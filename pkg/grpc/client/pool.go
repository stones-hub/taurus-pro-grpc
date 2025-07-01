// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Package client 提供了一个高性能、线程安全的 gRPC 连接池实现。
// 连接池支持一元调用(Unary)和流式调用(Stream)两种模式，并对每种模式的连接进行独立管理。
// 主要特性：
// - 支持每个地址维护多个连接
// - 支持连接的并发复用和负载控制
// - 自动清理空闲和异常连接
// - 提供详细的连接池状态统计

// ConnPool 实现了一个支持多地址、多连接的 gRPC 连接池。
// 连接池按地址维护连接，每个地址可以有多个连接，每个连接可以被多个请求复用。
// 连接池通过 load 计数来实现负载均衡，并通过定期清理来移除空闲和异常连接。
type ConnPool struct {
	mu       sync.RWMutex
	pools    map[string]*AddressPool // key是地址，value是该地址的连接池
	config   *PoolConfig             // 连接池配置
	cleanup  *time.Ticker            // 清理定时器
	stopChan chan struct{}           // 停止信号通道
}

// AddressPool 管理单个地址的连接池。
// 为了优化不同调用模式的性能，分别维护一元调用和流式调用的连接。
type AddressPool struct {
	mu          sync.RWMutex
	address     string
	unaryConns  []*ConnInfo // 一元调用连接池
	streamConns []*ConnInfo // 流式调用连接池
}

// ConnInfo 记录单个连接的详细信息。
// 使用 atomic.Int32 确保负载计数的并发安全。
type ConnInfo struct {
	conn      *grpc.ClientConn   // gRPC连接
	lastUsed  time.Time          // 最后使用时间
	createdAt time.Time          // 创建时间
	state     connectivity.State // 连接状态
	load      atomic.Int32       // 当前负载（原子操作）
	isStream  bool               // 是否是流式连接
}

// PoolConfig 定义连接池的配置参数。
// 包括连接数量控制和生命周期管理两个方面。
type PoolConfig struct {
	// 连接数量控制
	MinConnsPerAddr int   // 每个地址最小连接数，保证服务可用性
	MaxConnsPerAddr int   // 每个地址最大连接数，防止资源耗尽
	MaxIdleConns    int   // 每个地址最大空闲连接数，超过此数量的空闲连接将被清理
	MaxLoadPerConn  int32 // 每个连接最大负载，超过此负载将创建新连接或返回错误

	// 连接生命周期
	ConnMaxLifetime time.Duration // 连接最大生命周期，超过此时间的空闲连接将被清理
	ConnMaxIdleTime time.Duration // 连接最大空闲时间，超过此时间的空闲连接将被清理
	DialTimeout     time.Duration // 连接超时时间
}

// DefaultPoolConfig 返回默认配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		// 连接数量控制
		MinConnsPerAddr: 2,   // 默认每个地址保持2个连接
		MaxConnsPerAddr: 10,  // 默认每个地址最多10个连接
		MaxIdleConns:    2,   // 默认每个地址最多2个空闲连接
		MaxLoadPerConn:  100, // 默认每个连接最多100个并发请求

		// 连接生命周期
		ConnMaxLifetime: time.Hour,        // 连接最长生存1小时
		ConnMaxIdleTime: time.Minute * 30, // 空闲超过30分钟清理
		DialTimeout:     time.Second * 5,  // 连接超时5秒
	}
}

// NewConnPool 创建连接池
func NewConnPool(config *PoolConfig) *ConnPool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	pool := &ConnPool{
		pools:    make(map[string]*AddressPool),
		config:   config,
		cleanup:  time.NewTicker(time.Minute),
		stopChan: make(chan struct{}),
	}

	go pool.cleanupLoop()
	return pool
}

// GetConn 获取或创建一个可用的连接。
// 连接获取策略：
// 1. 优先从现有连接中选择负载较低的连接
// 2. 如果没有可用连接且未达到最大连接数，创建新连接
// 3. 如果达到最大连接数且所有连接都已满载，返回错误
func (p *ConnPool) GetConn(address string, isStream bool, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	p.mu.Lock()
	pool, exists := p.pools[address]
	if !exists {
		pool = &AddressPool{
			address:     address,
			unaryConns:  make([]*ConnInfo, 0, p.config.MinConnsPerAddr),
			streamConns: make([]*ConnInfo, 0, p.config.MinConnsPerAddr),
		}
		p.pools[address] = pool
	}
	p.mu.Unlock()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.unaryConns
	if isStream {
		conns = pool.streamConns
	}

	// 1. 使用一致性哈希或简单的负载均衡算法来选择连接
	if len(conns) > 0 {
		// 使用时间戳作为随机种子，确保分布均匀
		seed := time.Now().UnixNano()
		startIndex := int(seed % int64(len(conns)))

		// 从随机位置开始遍历，遍历一圈
		for i := 0; i < len(conns); i++ {
			index := (startIndex + i) % len(conns)
			connInfo := conns[index]

			if connInfo.conn.GetState() == connectivity.Ready {
				currentLoad := connInfo.load.Load()
				if currentLoad < p.config.MaxLoadPerConn {
					connInfo.lastUsed = time.Now()
					connInfo.load.Add(1)
					return connInfo.conn, nil
				}
			}
		}
	}

	// 2. 如果没有可用连接且未达到最大连接数，则创建新连接
	if len(conns) < p.config.MaxConnsPerAddr {
		conn, err := grpc.Dial(address, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %v", err)
		}

		connInfo := &ConnInfo{
			conn:      conn,
			lastUsed:  time.Now(),
			createdAt: time.Now(),
			state:     connectivity.Ready,
			isStream:  isStream,
		}
		connInfo.load.Store(1)

		if isStream {
			pool.streamConns = append(pool.streamConns, connInfo)
		} else {
			pool.unaryConns = append(pool.unaryConns, connInfo)
		}
		return conn, nil
	}

	// 3. 如果达到最大连接数，直接返回错误
	return nil, fmt.Errorf("connection pool exhausted: address=%s, max_conns=%d, all connections are at max load",
		address, p.config.MaxConnsPerAddr)
}

// ReleaseConn 释放连接的占用。
// 注意：这里不是真正关闭连接，而是将连接的负载计数减1。
// 连接的实际关闭由清理程序负责。
func (p *ConnPool) ReleaseConn(conn *grpc.ClientConn) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pool := range p.pools {
		pool.mu.Lock()

		// 检查并释放一元连接
		for _, connInfo := range pool.unaryConns {
			if connInfo.conn == conn {
				connInfo.load.Add(-1) // 只需要减少负载计数
				pool.mu.Unlock()
				return
			}
		}

		// 检查并释放流式连接
		for _, connInfo := range pool.streamConns {
			if connInfo.conn == conn {
				connInfo.load.Add(-1) // 只需要减少负载计数
				pool.mu.Unlock()
				return
			}
		}

		pool.mu.Unlock()
	}
}

// cleanupLoop 定期清理连接池中的连接。
// 清理策略：
// 1. 优先清理异常连接（TransientFailure 或 Shutdown 状态）
// 2. 清理空闲连接（load=0 且超过最大空闲时间或生命周期）
// 3. 保持最小连接数
func (p *ConnPool) cleanupLoop() {
	for {
		select {
		case <-p.cleanup.C:
			p.mu.RLock()
			for _, pool := range p.pools {
				pool.mu.Lock()
				now := time.Now()

				// 清理一元连接
				for i := 0; i < len(pool.unaryConns); i++ {
					conn := pool.unaryConns[i]

					// 1. 优先清理异常连接，不管 load 是多少
					if conn.conn.GetState() == connectivity.TransientFailure ||
						conn.conn.GetState() == connectivity.Shutdown {
						conn.conn.Close()
						pool.unaryConns[i] = pool.unaryConns[len(pool.unaryConns)-1]
						pool.unaryConns = pool.unaryConns[:len(pool.unaryConns)-1]
						i--
						continue
					}

					// 2. 清理空闲连接
					if conn.load.Load() == 0 &&
						(now.Sub(conn.lastUsed) > p.config.ConnMaxIdleTime ||
							now.Sub(conn.createdAt) > p.config.ConnMaxLifetime) {
						conn.conn.Close()
						pool.unaryConns[i] = pool.unaryConns[len(pool.unaryConns)-1]
						pool.unaryConns = pool.unaryConns[:len(pool.unaryConns)-1]
						i--
					}
				}

				// 清理流式连接，逻辑相同
				for i := 0; i < len(pool.streamConns); i++ {
					conn := pool.streamConns[i]

					if conn.conn.GetState() == connectivity.TransientFailure ||
						conn.conn.GetState() == connectivity.Shutdown {
						conn.conn.Close()
						pool.streamConns[i] = pool.streamConns[len(pool.streamConns)-1]
						pool.streamConns = pool.streamConns[:len(pool.streamConns)-1]
						i--
						continue
					}

					if conn.load.Load() == 0 &&
						(now.Sub(conn.lastUsed) > p.config.ConnMaxIdleTime ||
							now.Sub(conn.createdAt) > p.config.ConnMaxLifetime) {
						conn.conn.Close()
						pool.streamConns[i] = pool.streamConns[len(pool.streamConns)-1]
						pool.streamConns = pool.streamConns[:len(pool.streamConns)-1]
						i--
					}
				}
				pool.mu.Unlock()
			}
			p.mu.RUnlock()
		case <-p.stopChan:
			return
		}
	}
}

// Stats 返回连接池的详细统计信息。
// 统计指标包括：
// - 总连接数
// - 可用连接数（Ready状态）
// - 空闲连接数（load=0）
// - 异常连接数
// - 总负载
// 统计数据按一元调用和流式调用分别统计，并提供每个地址的详细统计。
func (p *ConnPool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"unary": map[string]interface{}{
			"by_address": make(map[string]interface{}),
			"total": map[string]interface{}{
				"total_connections":  0,
				"ready_connections":  0,        // Ready状态的连接数
				"idle_connections":   0,        // 空闲连接数（load=0）
				"failed_connections": 0,        // 失败的连接数（TransientFailure或Shutdown）
				"total_load":         int32(0), // 总负载
			},
		},
		"stream": map[string]interface{}{
			"by_address": make(map[string]interface{}),
			"total": map[string]interface{}{
				"total_connections":  0,
				"ready_connections":  0,
				"idle_connections":   0,
				"failed_connections": 0,
				"total_load":         int32(0),
			},
		},
	}

	unaryStats := stats["unary"].(map[string]interface{})
	streamStats := stats["stream"].(map[string]interface{})

	for addr, pool := range p.pools {
		pool.mu.RLock()

		// 统计一元连接
		unaryAddrStats := map[string]interface{}{
			"total_connections":  len(pool.unaryConns),
			"ready_connections":  0,
			"idle_connections":   0,
			"failed_connections": 0,
			"total_load":         int32(0),
		}

		// 统计每个一元连接的状态
		for _, connInfo := range pool.unaryConns {
			state := connInfo.conn.GetState()
			load := connInfo.load.Load()

			switch state {
			case connectivity.Ready:
				unaryAddrStats["ready_connections"] = unaryAddrStats["ready_connections"].(int) + 1
				if load == 0 {
					unaryAddrStats["idle_connections"] = unaryAddrStats["idle_connections"].(int) + 1
				}
			case connectivity.TransientFailure, connectivity.Shutdown:
				unaryAddrStats["failed_connections"] = unaryAddrStats["failed_connections"].(int) + 1
			}
			unaryAddrStats["total_load"] = unaryAddrStats["total_load"].(int32) + load
		}

		// 更新一元连接总统计
		unaryTotal := unaryStats["total"].(map[string]interface{})
		unaryTotal["total_connections"] = unaryTotal["total_connections"].(int) + unaryAddrStats["total_connections"].(int)
		unaryTotal["ready_connections"] = unaryTotal["ready_connections"].(int) + unaryAddrStats["ready_connections"].(int)
		unaryTotal["idle_connections"] = unaryTotal["idle_connections"].(int) + unaryAddrStats["idle_connections"].(int)
		unaryTotal["failed_connections"] = unaryTotal["failed_connections"].(int) + unaryAddrStats["failed_connections"].(int)
		unaryTotal["total_load"] = unaryTotal["total_load"].(int32) + unaryAddrStats["total_load"].(int32)

		unaryStats["by_address"].(map[string]interface{})[addr] = unaryAddrStats

		// 统计流式连接（逻辑相同）
		streamAddrStats := map[string]interface{}{
			"total_connections":  len(pool.streamConns),
			"ready_connections":  0,
			"idle_connections":   0,
			"failed_connections": 0,
			"total_load":         int32(0),
		}

		for _, connInfo := range pool.streamConns {
			state := connInfo.conn.GetState()
			load := connInfo.load.Load()

			switch state {
			case connectivity.Ready:
				streamAddrStats["ready_connections"] = streamAddrStats["ready_connections"].(int) + 1
				if load == 0 {
					streamAddrStats["idle_connections"] = streamAddrStats["idle_connections"].(int) + 1
				}
			case connectivity.TransientFailure, connectivity.Shutdown:
				streamAddrStats["failed_connections"] = streamAddrStats["failed_connections"].(int) + 1
			}
			streamAddrStats["total_load"] = streamAddrStats["total_load"].(int32) + load
		}

		// 更新流式连接总统计
		streamTotal := streamStats["total"].(map[string]interface{})
		streamTotal["total_connections"] = streamTotal["total_connections"].(int) + streamAddrStats["total_connections"].(int)
		streamTotal["ready_connections"] = streamTotal["ready_connections"].(int) + streamAddrStats["ready_connections"].(int)
		streamTotal["idle_connections"] = streamTotal["idle_connections"].(int) + streamAddrStats["idle_connections"].(int)
		streamTotal["failed_connections"] = streamTotal["failed_connections"].(int) + streamAddrStats["failed_connections"].(int)
		streamTotal["total_load"] = streamTotal["total_load"].(int32) + streamAddrStats["total_load"].(int32)

		streamStats["by_address"].(map[string]interface{})[addr] = streamAddrStats

		pool.mu.RUnlock()
	}

	return stats
}

// Close 关闭连接池。
// 1. 停止清理定时器
// 2. 关闭所有连接
// 3. 清空连接池
// 返回最后一个发生的错误（如果有）。
func (p *ConnPool) Close() error {
	p.cleanup.Stop()
	close(p.stopChan)

	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for _, pool := range p.pools {
		pool.mu.Lock()

		// 关闭所有一元连接
		for _, connInfo := range pool.unaryConns {
			if err := connInfo.conn.Close(); err != nil {
				lastErr = err
			}
		}
		pool.unaryConns = nil

		// 关闭所有流式连接
		for _, connInfo := range pool.streamConns {
			if err := connInfo.conn.Close(); err != nil {
				lastErr = err
			}
		}
		pool.streamConns = nil

		pool.mu.Unlock()
	}

	p.pools = nil
	return lastErr
}

// CloseAddress 关闭指定地址的所有连接
// 返回最后一个发生的错误（如果有）
func (p *ConnPool) CloseAddress(address string) error {
	p.mu.Lock()
	pool, exists := p.pools[address]
	if !exists {
		p.mu.Unlock()
		return nil
	}

	// 从连接池映射中删除，防止新的连接被创建
	delete(p.pools, address)
	p.mu.Unlock()

	// 关闭该地址的所有连接
	pool.mu.Lock()
	defer pool.mu.Unlock()

	var lastErr error

	// 关闭一元连接
	for _, connInfo := range pool.unaryConns {
		if err := connInfo.conn.Close(); err != nil {
			lastErr = fmt.Errorf("close unary connection error: %v", err)
		}
		// 确保连接状态更新
		connInfo.state = connectivity.Shutdown
	}
	pool.unaryConns = nil

	// 关闭流式连接
	for _, connInfo := range pool.streamConns {
		if err := connInfo.conn.Close(); err != nil {
			lastErr = fmt.Errorf("close stream connection error: %v", err)
		}
		// 确保连接状态更新
		connInfo.state = connectivity.Shutdown
	}
	pool.streamConns = nil

	return lastErr
}
