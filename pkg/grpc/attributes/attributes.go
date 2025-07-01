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

package attributes

import (
	"context"

	"google.golang.org/grpc"
)

// ------------------------------------------------------------
// 客户端拦截器链
// ------------------------------------------------------------
// 拦截器链 客户端一元请求
func ChainUnaryClient(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		chain := invoker
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.UnaryInvoker, interceptor grpc.UnaryClientInterceptor) grpc.UnaryInvoker {
				return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
					return interceptor(ctx, method, req, reply, cc, next, opts...)
				}
			}(chain, interceptors[i])
		}
		return chain(ctx, method, req, reply, cc, opts...)
	}
}

// 拦截器链 客户端流请求
func ChainStreamClient(interceptors ...grpc.StreamClientInterceptor) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		chain := streamer
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.Streamer, interceptor grpc.StreamClientInterceptor) grpc.Streamer {
				return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
					return interceptor(ctx, desc, cc, method, next, opts...)
				}
			}(chain, interceptors[i])
		}
		return chain(ctx, desc, cc, method, opts...)
	}
}

// ------------------------------------------------------------
// 服务端拦截器链和中间件链
// ------------------------------------------------------------

// 定义中间件类型
type UnaryMiddleware func(grpc.UnaryHandler) grpc.UnaryHandler

type StreamMiddleware func(grpc.StreamHandler) grpc.StreamHandler

// 拦截器链 服务端一元请求
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	// interceptors = [] func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (resp any, err error)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// handler =  func(ctx context.Context, req any) (any, error)
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.UnaryHandler, interceptor grpc.UnaryServerInterceptor) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return interceptor(ctx, req, info, next)
				}
			}(chain, interceptors[i])
		}

		return chain(ctx, req)
	}
}

// 拦截器链 服务端流请求
func ChainStreamServer(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.StreamHandler, interceptor grpc.StreamServerInterceptor) grpc.StreamHandler {
				return func(srv interface{}, ss grpc.ServerStream) error {
					return interceptor(srv, ss, info, next)
				}
			}(chain, interceptors[i])
		}
		return chain(srv, ss)
	}
}

// 同时处理中间件和拦截器
func ChainUnaryInterceptorWithMiddlewareServer(mids []UnaryMiddleware, interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		midChain := handler
		for i := len(mids) - 1; i >= 0; i-- {
			midChain = mids[i](midChain)
		}

		chain := midChain
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.UnaryHandler, interceptor grpc.UnaryServerInterceptor) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return interceptor(ctx, req, info, next)
				}
			}(chain, interceptors[i])
		}

		return chain(ctx, req)
	}
}

func ChainStreamInterceptorWithMiddlewareServer(mids []StreamMiddleware, interceptors []grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		midChain := handler
		for i := len(mids) - 1; i >= 0; i-- {
			midChain = mids[i](midChain)
		}

		chain := midChain
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = func(next grpc.StreamHandler, interceptor grpc.StreamServerInterceptor) grpc.StreamHandler {
				return func(srv interface{}, ss grpc.ServerStream) error {
					return interceptor(srv, ss, info, next)
				}
			}(chain, interceptors[i])
		}
		return chain(srv, ss)
	}
}

// Notice: This function is EXPERIMENTAL and may be changed or removed in the future.
// middlewares chain
func ChainUnaryMiddlewareServer(middlewares ...UnaryMiddleware) UnaryMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Notice: This function is EXPERIMENTAL and may be changed or removed in the future.
// middlewares chain
func ChainStreamMiddlewareServer(middlewares ...StreamMiddleware) StreamMiddleware {
	return func(next grpc.StreamHandler) grpc.StreamHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
