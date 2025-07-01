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
package server

import (
	"crypto/tls"
	"time"

	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// ServerOption 定义服务器配置选项
type ServerOption func(*ServerOptions)

// ServerOptions 包含所有服务器配置
type ServerOptions struct {
	// 基础配置
	Address   string      // 服务器地址
	TLSConfig *tls.Config // TLS配置

	// 高级配置
	KeepAlive          *keepalive.ServerParameters    // KeepAlive配置
	MaxConns           int                            // 最大连接数
	UnaryInterceptors  []grpc.UnaryServerInterceptor  // 一元拦截器
	StreamInterceptors []grpc.StreamServerInterceptor // 流拦截器

	// 自定义配置中间件
	UnaryMiddlewares  []attributes.UnaryMiddleware  // 一元中间件
	StreamMiddlewares []attributes.StreamMiddleware // 流中间件
}

// DefaultServerOptions 返回默认配置
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Address: ":50051",
		KeepAlive: &keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,  // 空闲连接最长保持时间
			MaxConnectionAge:      10 * time.Minute, // 连接在接收到关闭信号，还能保持的时间
			MaxConnectionAgeGrace: 5 * time.Second,  // MaxConnectionAgeGrace是MaxConnectionAge之后的一个附加周期, 过了这个周期强制关闭
			Time:                  2 * time.Hour,    // 服务器2小时后发送ping，判断是否连接存活
			Timeout:               20 * time.Second, // 在Time参数时间后，发送了ping后，如果20秒内没有收到客户端的pong，则关闭连接
		},
		UnaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		StreamInterceptors: make([]grpc.StreamServerInterceptor, 0),
		UnaryMiddlewares:   make([]attributes.UnaryMiddleware, 0),
		StreamMiddlewares:  make([]attributes.StreamMiddleware, 0),
	}
}

// WithAddress 设置服务器地址
func WithAddress(addr string) ServerOption {
	return func(o *ServerOptions) {
		o.Address = addr
	}
}

// WithTLS 设置TLS配置
func WithTLS(config *tls.Config) ServerOption {
	return func(o *ServerOptions) {
		o.TLSConfig = config
	}
}

// WithKeepAlive 设置KeepAlive配置
func WithKeepAlive(config *keepalive.ServerParameters) ServerOption {
	return func(o *ServerOptions) {
		o.KeepAlive = config
	}
}

// WithMaxConns 设置最大连接数
func WithMaxConns(maxConns int) ServerOption {
	return func(o *ServerOptions) {
		o.MaxConns = maxConns
	}
}

// WithUnaryInterceptor 添加一元拦截器
func WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) ServerOption {
	return func(o *ServerOptions) {
		o.UnaryInterceptors = append(o.UnaryInterceptors, interceptor)
	}
}

// WithStreamInterceptor 添加流拦截器
func WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerOption {
	return func(o *ServerOptions) {
		o.StreamInterceptors = append(o.StreamInterceptors, interceptor)
	}
}

// WithUnaryMiddleware 添加一元中间件
func WithUnaryMiddleware(middleware attributes.UnaryMiddleware) ServerOption {
	return func(o *ServerOptions) {
		o.UnaryMiddlewares = append(o.UnaryMiddlewares, middleware)
	}
}

// WithStreamMiddleware 添加流中间件
func WithStreamMiddleware(middleware attributes.StreamMiddleware) ServerOption {
	return func(o *ServerOptions) {
		o.StreamMiddlewares = append(o.StreamMiddlewares, middleware)
	}
}
