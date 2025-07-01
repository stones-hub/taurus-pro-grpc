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
	"fmt"
	"log"
	"net"
	"time"

	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// Server gRPC服务器封装
type Server struct {
	server *grpc.Server   // gRPC服务器实例
	opts   *ServerOptions // 服务器配置
}

// NewServer 创建新的gRPC服务器
func NewServer(opts ...ServerOption) (*Server, func(), error) {
	options := DefaultServerOptions()
	for _, opt := range opts {
		opt(options)
	}

	serverOpts := []grpc.ServerOption{
		// 基础资源限制
		// MaxRecvMsgSize 限制服务器接收的最大消息大小，默认值为10MB
		// 如果客户端发送的消息超过此大小，请求会被拒绝
		// 根据业务需求调整，比如文件上传场景可能需要更大的值
		grpc.MaxRecvMsgSize(1024 * 1024 * 10),

		// MaxSendMsgSize 限制服务器发送的最大消息大小，默认值为10MB
		// 如果服务器响应的消息超过此大小，响应会被拒绝
		// 通常与 MaxRecvMsgSize 设置相同的值保持一致性
		grpc.MaxSendMsgSize(1024 * 1024 * 10),

		// MaxConcurrentStreams 限制每个HTTP2连接上的最大并发流数量
		// 即一个客户端连接能同时处理的最大请求数
		// 默认值1000适用于大多数场景，可根据服务器资源情况调整
		grpc.MaxConcurrentStreams(1000),

		// 连接保活策略配置
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			// MinTime 指定客户端发送keepalive ping的最小间隔时间
			// 如果客户端发送过于频繁的ping，服务器会关闭连接
			// 5秒的间隔可以在保持连接活跃和避免资源浪费之间取得平衡
			MinTime: time.Second * 5,

			// PermitWithoutStream 允许客户端在没有活动流的情况下发送keepalive ping
			// true: 即使没有正在进行的请求也保持连接
			// false: 只有在有活动请求时才允许发送keepalive ping
			// 建议设置为true以维持长连接，特别是在微服务架构中
			PermitWithoutStream: true,
		}),
	}

	// TLS配置
	if options.TLSConfig != nil {
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(options.TLSConfig)))
	}

	// KeepAlive配置
	if options.KeepAlive != nil {
		serverOpts = append(serverOpts, grpc.KeepaliveParams(*options.KeepAlive))
	}

	// 用户自定义拦截器配置
	if len(options.UnaryMiddlewares) > 0 {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(
			attributes.ChainUnaryInterceptorWithMiddlewareServer(
				options.UnaryMiddlewares,
				options.UnaryInterceptors,
			),
		))
	} else if len(options.UnaryInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(
			attributes.ChainUnaryServer(options.UnaryInterceptors...),
		))
	}

	if len(options.StreamMiddlewares) > 0 {
		serverOpts = append(serverOpts, grpc.StreamInterceptor(
			attributes.ChainStreamInterceptorWithMiddlewareServer(
				options.StreamMiddlewares,
				options.StreamInterceptors,
			),
		))
	} else if len(options.StreamInterceptors) > 0 {
		serverOpts = append(serverOpts, grpc.StreamInterceptor(
			attributes.ChainStreamServer(options.StreamInterceptors...),
		))
	}

	// 创建服务器实例
	server := grpc.NewServer(serverOpts...)

	// 注册健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	grpcServer := &Server{
		server: server,
		opts:   options,
	}

	return grpcServer, func() {
		grpcServer.server.GracefulStop()
		log.Println("gRPC server stopped successfully")
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	log.Println("Starting gRPC server on", s.opts.Address)
	lis, err := net.Listen("tcp", s.opts.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return s.server.Serve(lis)
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// Server 获取原始服务器实例
func (s *Server) Server() *grpc.Server {
	return s.server
}

/*
# 1. 首先确保安装了必要的工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# 2. 生成gRPC代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/controller/gRPC/proto/user/user.proto
*/
