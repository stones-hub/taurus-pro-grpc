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
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// 一元模式和流式模式区别
// 一元模式:
// 1. 同一个链接被多个协程同时复用，是排队执行的， 前提是你没有关闭链接
// 流模式:
// 1. 同一个链接被多个协程同时复用，是并发执行的， 前提是你没有关闭链接

// PoolOptions 连接池配置
type PoolOptions struct {
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大打开连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	MaxLoadPerConn  int32         // 每个连接的最大负载
}

// DefaultPoolOptions 返回默认连接池配置
func DefaultPoolOptions() *PoolOptions {
	return &PoolOptions{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		MaxLoadPerConn:  1000,
	}
}

// ClientOptions 包含所有客户端配置
type ClientOptions struct {
	// 基础配置
	Timeout   time.Duration // 连接超时时间
	TLSConfig *tls.Config   // TLS配置

	// 连接池配置
	Pool *PoolOptions

	// 通用配置
	KeepAlive          *keepalive.ClientParameters    // 保活配置
	UnaryInterceptors  []grpc.UnaryClientInterceptor  // 一元拦截器
	StreamInterceptors []grpc.StreamClientInterceptor // 流式拦截器
}

// DefaultClientOptions 返回默认配置
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Timeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Pool: DefaultPoolOptions(),
		KeepAlive: &keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		},
	}
}

// ClientOption 定义配置选项函数
type ClientOption func(*ClientOptions)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithTLS 设置TLS配置
func WithTLS(config *tls.Config) ClientOption {
	return func(o *ClientOptions) {
		o.TLSConfig = config
	}
}

// WithInsecure 设置非安全连接
func WithInsecure() ClientOption {
	return func(o *ClientOptions) {
		o.TLSConfig = nil
	}
}

// WithKeepAlive 设置保活配置
func WithKeepAlive(config *keepalive.ClientParameters) ClientOption {
	return func(o *ClientOptions) {
		o.KeepAlive = config
	}
}

// WithPool 设置连接池配置
func WithPool(pool *PoolOptions) ClientOption {
	return func(o *ClientOptions) {
		o.Pool = pool
	}
}

// WithPoolConfig 设置连接池详细配置
func WithPoolConfig(maxIdle, maxOpen int, maxLifetime, maxIdleTime time.Duration, maxLoad int32) ClientOption {
	return func(o *ClientOptions) {
		o.Pool = &PoolOptions{
			MaxIdleConns:    maxIdle,
			MaxOpenConns:    maxOpen,
			ConnMaxLifetime: maxLifetime,
			ConnMaxIdleTime: maxIdleTime,
			MaxLoadPerConn:  maxLoad,
		}
	}
}

// WithUnaryInterceptor 添加一元拦截器
func WithUnaryInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(o *ClientOptions) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, interceptor)
	}
}

// WithStreamInterceptor 添加流式拦截器
func WithStreamInterceptor(interceptor grpc.StreamClientInterceptor) ClientOption {
	return func(o *ClientOptions) {
		o.StreamInterceptors = append(o.StreamInterceptors, interceptor)
	}
}
