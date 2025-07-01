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
	"fmt"
	"log"

	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/attributes"
	"google.golang.org/grpc"
)

// 日志中间件
func LoggingMiddleware() attributes.UnaryMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			fmt.Printf("Request: %v\n", req)

			resp, err := next(ctx, req)

			log.Printf("Response: %v, Error: %v", resp, err)
			return resp, err
		}
	}
}
