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
package interceptor

import (
	"context"
	"fmt"

	"github.com/stones-hub/taurus-pro-grpc/pkg/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerValidationInterceptor 创建一个gRPC一元服务验证拦截器
func UnaryServerValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 验证请求参数, 对于请求参数 req做验证， 切记 生成的proto文件， 需要添加validate标签
		if err := validate.ValidateStruct(req); err != nil {
			// 如果是验证错误，返回InvalidArgument状态
			if valErrs, ok := err.(validate.ValidationErrors); ok {
				errMsg := fmt.Sprintf("请求参数验证失败: %s", valErrs.Error())
				return nil, status.Error(codes.InvalidArgument, errMsg)
			}
			// 其他错误
			return nil, status.Error(codes.Internal, fmt.Sprintf("请求验证出现内部错误: %v", err))
		}

		// 验证通过，继续处理请求
		return handler(ctx, req)
	}
}

// StreamServerValidationInterceptor 创建一个gRPC流服务验证拦截器
func StreamServerValidationInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{
			ServerStream: ss,
		}
		return handler(srv, wrapper)
	}
}

// recvWrapper 包装流服务器，用于验证每个接收到的消息
type recvWrapper struct {
	grpc.ServerStream
}

// RecvMsg 拦截并验证接收到的消息
func (s *recvWrapper) RecvMsg(m interface{}) error {
	// 首先调用原始的RecvMsg
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	// 验证接收到的消息
	if err := validate.ValidateStruct(m); err != nil {
		if valErrs, ok := err.(validate.ValidationErrors); ok {
			errMsg := fmt.Sprintf("请求参数验证失败: %s", valErrs.Error())
			return status.Error(codes.InvalidArgument, errMsg)
		}
		return status.Error(codes.Internal, fmt.Sprintf("请求验证出现内部错误: %v", err))
	}

	return nil
}
