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
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"google.golang.org/grpc"
)

// ServiceRegistrar 服务注册接口
type ServiceRegistrar interface {
	RegisterService(server *grpc.Server)
}

// 服务注册表
var (
	serviceRegistry          = make(map[string]ServiceRegistrar)
	serviceMiddleware        = make([]attributes.UnaryMiddleware, 0)
	serviceStreamMiddleware  = make([]attributes.StreamMiddleware, 0)
	serviceInterceptor       = make([]grpc.UnaryServerInterceptor, 0)
	serviceStreamInterceptor = make([]grpc.StreamServerInterceptor, 0)
)

// RegisterService 注册服务
func RegisterService(name string, service ServiceRegistrar) {
	serviceRegistry[name] = service
}

// GetRegisteredServices 获取所有注册的服务
func GetRegisteredServices() map[string]ServiceRegistrar {
	return serviceRegistry
}

func RegisterMiddleware(middleware attributes.UnaryMiddleware) {
	serviceMiddleware = append(serviceMiddleware, middleware)
}

func RegisterStreamMiddleware(middleware attributes.StreamMiddleware) {
	serviceStreamMiddleware = append(serviceStreamMiddleware, middleware)
}

func RegisterInterceptor(interceptor grpc.UnaryServerInterceptor) {
	serviceInterceptor = append(serviceInterceptor, interceptor)
}

func RegisterStreamInterceptor(interceptor grpc.StreamServerInterceptor) {
	serviceStreamInterceptor = append(serviceStreamInterceptor, interceptor)
}

func GetServiceMiddleware() []attributes.UnaryMiddleware {
	return serviceMiddleware
}

func GetServiceStreamMiddleware() []attributes.StreamMiddleware {
	return serviceStreamMiddleware
}

func GetServiceInterceptor() []grpc.UnaryServerInterceptor {
	return serviceInterceptor
}

func GetServiceStreamInterceptor() []grpc.StreamServerInterceptor {
	return serviceStreamInterceptor
}
