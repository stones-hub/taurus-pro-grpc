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
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 定义客户端接口
type Client interface {
	// GetConn 获取连接，isStream 参数指定是否为流式连接
	GetConn(address string, isStream bool) (*grpc.ClientConn, error)
	// ReleaseConn 释放连接
	ReleaseConn(*grpc.ClientConn)
	// CloseAddress 关闭指定地址的所有连接
	CloseAddress(address string) error
	// Close 关闭客户端
	Close() error
	// Options 获取配置
	Options() *ClientOptions
}

// GrpcClient 统一的gRPC客户端实现
type GrpcClient struct {
	opts *ClientOptions
	pool *ConnPool
}

// NewClient 创建新的客户端
func NewClient(opts ...ClientOption) (Client, error) {
	options := DefaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 创建连接池时使用新的配置结构
	pool := NewConnPool(&PoolConfig{
		// 连接数量控制
		MinConnsPerAddr: 2, // 默认每个地址保持2个连接
		MaxConnsPerAddr: options.Pool.MaxOpenConns,
		MaxIdleConns:    options.Pool.MaxIdleConns,
		MaxLoadPerConn:  options.Pool.MaxLoadPerConn,

		// 连接生命周期
		ConnMaxLifetime: options.Pool.ConnMaxLifetime,
		ConnMaxIdleTime: options.Pool.ConnMaxIdleTime,
		DialTimeout:     options.Timeout,
	})

	return &GrpcClient{
		opts: options,
		pool: pool,
	}, nil
}

func (c *GrpcClient) getDialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{}

	// TLS配置
	if c.opts.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.opts.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// KeepAlive配置
	if c.opts.KeepAlive != nil {
		opts = append(opts, grpc.WithKeepaliveParams(*c.opts.KeepAlive))
	}

	// 一元拦截器
	if len(c.opts.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.WithUnaryInterceptor(attributes.ChainUnaryClient(c.opts.UnaryInterceptors...)))
	}

	// 流式拦截器
	if len(c.opts.StreamInterceptors) > 0 {
		opts = append(opts, grpc.WithStreamInterceptor(attributes.ChainStreamClient(c.opts.StreamInterceptors...)))
	}

	return opts
}

// GetConn 获取连接，由调用者指定是否为流式连接
func (c *GrpcClient) GetConn(address string, isStream bool) (*grpc.ClientConn, error) {
	return c.pool.GetConn(address, isStream, c.getDialOptions()...)
}

// ReleaseConn 释放连接
func (c *GrpcClient) ReleaseConn(conn *grpc.ClientConn) {
	c.pool.ReleaseConn(conn)
}

// CloseAddress 关闭指定地址的所有连接
func (c *GrpcClient) CloseAddress(address string) error {
	return c.pool.CloseAddress(address)
}

// Close 关闭客户端
func (c *GrpcClient) Close() error {
	return c.pool.Close()
}

// Options 获取配置
func (c *GrpcClient) Options() *ClientOptions {
	return c.opts
}
