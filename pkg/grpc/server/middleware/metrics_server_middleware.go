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
package middleware

import (
	"context"
	"crypto/md5"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// 监控中间件, 引入opentelemetry的监控
func MetricsMiddleware(tracer trace.Tracer) attributes.UnaryMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()

			// 使用 MD5 生成 16 字节的 TraceID
			hash := md5.Sum([]byte(uuid.New().String()))
			var traceID trace.TraceID
			copy(traceID[:], hash[:])

			// 获取 gRPC 方法信息
			fullMethod, _ := grpc.Method(ctx)
			service := path.Dir(fullMethod)[1:]
			method := path.Base(fullMethod)

			// 创建新的spanContext
			spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceID,
			})

			// 将SpanContext注入到上下文中
			ctx = trace.ContextWithSpanContext(ctx, spanCtx)

			// 获取 peer 信息
			peer, _ := peer.FromContext(ctx)
			peerAddr := "unknown"
			if peer != nil {
				peerAddr = peer.Addr.String()
			}

			// 创建新的 span，并记录请求的详细信息
			spanName := "grpc." + service + "." + method
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithAttributes(
					attribute.String("rpc.system", "grpc"),
					attribute.String("rpc.service", service),
					attribute.String("rpc.method", method),
					attribute.String("rpc.peer.address", peerAddr),
					attribute.String("rpc.trace_id", traceID.String()),
					attribute.String("rpc.at_time", time.Now().Format(time.RFC3339)),
				),
			)
			defer span.End()

			// 使用新的带有追踪信息的上下文调用下一个处理函数
			resp, err := next(ctx, req)

			// 记录处理时间和响应状态
			duration := time.Since(start)
			statusCode := "OK"
			if err != nil {
				statusCode = status.Code(err).String()
				span.SetAttributes(attribute.String("error", err.Error()))
			}

			span.SetAttributes(
				attribute.String("rpc.status", statusCode),
				attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
			)

			return resp, err
		}
	}
}
