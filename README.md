# Taurus Pro gRPC

[![Go Version](https://img.shields.io/badge/Go-1.24.2+-blue.svg)](https://golang.org)
[![gRPC Version](https://img.shields.io/badge/gRPC-1.73.0+-green.svg)](https://grpc.io)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/stones-hub/taurus-pro-grpc)](https://goreportcard.com/report/github.com/stones-hub/taurus-pro-grpc)

**Taurus Pro gRPC** 是一个企业级的 Go gRPC 框架，提供了完整的客户端和服务器端解决方案，包含连接池管理、拦截器、中间件、TLS 支持、认证授权、监控指标等核心功能。

## ✨ 特性

### 🚀 核心功能
- **统一客户端管理**: 支持连接池、负载均衡、自动重连
- **高性能服务器**: 内置资源限制、连接保活、健康检查
- **TLS 安全通信**: 完整的 TLS 配置支持，包括双向认证
- **拦截器系统**: 丰富的客户端和服务器端拦截器
- **中间件支持**: 日志记录、监控指标、限流、恢复等中间件

### 🔧 客户端特性
- **连接池管理**: 智能连接池，支持最小/最大连接数配置
- **负载均衡**: 基于连接负载的智能分发
- **重试机制**: 可配置的重试策略和超时控制
- **认证拦截器**: 自动添加认证信息到请求头
- **日志拦截器**: 详细的请求/响应日志记录

### 🛡️ 服务器特性
- **资源限制**: 可配置的消息大小、并发流数限制
- **连接保活**: 灵活的 KeepAlive 策略配置
- **健康检查**: 内置 gRPC 健康检查服务
- **认证授权**: 基于 Token 的认证拦截器
- **限流控制**: 可配置的请求限流策略
- **监控指标**: 内置 Prometheus 指标收集

### 📊 监控与可观测性
- **OpenTelemetry 集成**: 完整的链路追踪支持
- **指标收集**: 请求计数、延迟、错误率等关键指标
- **结构化日志**: 统一的日志格式和级别控制

### 🔍 数据验证
- **多语言支持**: 中文和英文错误消息
- **字段验证**: 丰富的验证标签和自定义规则
- **错误翻译**: 自动化的错误消息本地化

## 📦 安装

### 前置要求
- Go 1.24.2 或更高版本
- Protocol Buffers 编译器 (protoc)

### 安装依赖
```bash
go mod download
```

### 生成 Protocol Buffers 代码
```bash
# 在项目根目录执行
cd bin/proto && ./generate.sh

# 或在示例目录执行
cd example/proto && ./generate.sh
```

## 🚀 快速开始

### 1. 启动服务器

```go
package main

import (
    "log"
    "github.com/stones-hub/taurus-pro-grpc/pkg/grpc/server"
    "google.golang.org/grpc/keepalive"
    "time"
)

func main() {
    // 创建服务器选项
    opts := []server.ServerOption{
        server.WithAddress(":50051"),
        server.WithKeepAlive(&keepalive.ServerParameters{
            MaxConnectionIdle: time.Minute * 5,
            Time:              time.Second * 60,
            Timeout:           time.Second * 20,
        }),
    }

    // 创建 gRPC 服务器
    s, cleanup, err := server.NewServer(opts...)
    if err != nil {
        log.Fatalf("failed to create server: %v", err)
    }
    defer cleanup()

    // 注册您的服务
    // pb.RegisterYourServiceServer(s.Server(), &yourService{})

    // 启动服务器
    log.Printf("Starting gRPC server on :50051")
    if err := s.Start(); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}
```

### 2. 创建客户端

```go
package main

import (
    "log"
    "time"
    "github.com/stones-hub/taurus-pro-grpc/pkg/grpc/client"
    "google.golang.org/grpc/keepalive"
)

func main() {
    // 创建客户端选项
    opts := []client.ClientOption{
        client.WithKeepAlive(&keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             5 * time.Second,
            PermitWithoutStream: true,
        }),
    }

    // 创建客户端
    c, err := client.NewClient(opts...)
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
    defer c.Close()

    // 获取连接
    conn, err := c.GetConn("localhost:50051", false)
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer c.ReleaseConn(conn)

    // 使用连接创建 gRPC 客户端
    // client := pb.NewYourServiceClient(conn)
}
```

## 🔐 TLS 配置

### 生成证书
```bash
cd example/certs && ./generate.sh
```

### 服务器端 TLS
```go
tlsConfig, err := loadTLSConfig()
if err != nil {
    log.Fatalf("failed to load TLS config: %v", err)
}

opts := []server.ServerOption{
    server.WithAddress(":50051"),
    server.WithTLS(tlsConfig),
}
```

### 客户端 TLS
```go
tlsConfig, err := loadTLSConfig()
if err != nil {
    log.Fatalf("failed to load TLS config: %v", err)
}

opts := []client.ClientOption{
    client.WithTLS(tlsConfig),
}
```

## 🎯 拦截器使用

### 客户端拦截器
```go
// 认证拦截器
authInterceptor := interceptor.AuthInterceptor("your-token")

// 日志拦截器
loggingInterceptor := interceptor.LoggingClientInterceptor()

opts := []client.ClientOption{
    client.WithUnaryInterceptors(authInterceptor, loggingInterceptor),
}
```

### 服务器端拦截器
```go
// 认证拦截器
authInterceptor := interceptor.AuthServerInterceptor("your-token")

// 日志拦截器
loggingInterceptor := interceptor.LoggingServerInterceptor()

opts := []server.ServerOption{
    server.WithUnaryInterceptors(authInterceptor, loggingInterceptor),
}
```

## 📊 监控与指标

### 启用指标中间件
```go
opts := []server.ServerOption{
    server.WithMetricsMiddleware("your-service-name"),
}
```

### 启用链路追踪
```go
// 客户端
opts := []client.ClientOption{
    client.WithUnaryInterceptors(interceptor.TracingClientInterceptor()),
}

// 服务器端
opts := []server.ServerOption{
    server.WithUnaryInterceptors(interceptor.TracingServerInterceptor()),
}
```

## 🔍 数据验证

### 使用验证器
```go
import "github.com/stones-hub/taurus-pro-grpc/pkg/validate"

type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=0,lte=150"`
}

func validateUser(user *User) error {
    return validate.Core.Struct(user)
}
```

## 📁 项目结构

```
taurus-pro-grpc/
├── bin/                    # 二进制文件目录
│   ├── client/            # 客户端示例
│   ├── proto/             # Protocol Buffers 定义
│   └── server/            # 服务器示例
├── example/               # 完整示例
│   ├── certs/            # TLS 证书生成脚本
│   ├── client/           # 客户端示例
│   ├── proto/            # Protocol Buffers 示例
│   └── server/           # 服务器示例
├── pkg/                  # 核心包
│   ├── grpc/             # gRPC 核心功能
│   │   ├── attributes/   # 属性管理
│   │   ├── client/       # 客户端实现
│   │   └── server/       # 服务器实现
│   └── validate/         # 数据验证
├── go.mod               # Go 模块文件
├── go.sum               # Go 依赖校验文件
└── README.md            # 项目文档
```

## 🧪 运行示例

### 1. 生成证书
```bash
cd example/certs && ./generate.sh
```

### 2. 启动服务器
```bash
cd example/server && go run main.go
```

### 3. 运行客户端
```bash
cd example/client && go run main.go
```

## 📚 API 文档

### 客户端选项

| 选项 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `WithAddress` | string | 服务器地址 | ":50051" |
| `WithTLS` | *tls.Config | TLS 配置 | nil |
| `WithKeepAlive` | *keepalive.ClientParameters | KeepAlive 参数 | 默认配置 |
| `WithTimeout` | time.Duration | 连接超时 | 30s |
| `WithUnaryInterceptors` | []grpc.UnaryClientInterceptor | 一元拦截器 | 空切片 |
| `WithStreamInterceptors` | []grpc.StreamClientInterceptor | 流式拦截器 | 空切片 |

### 服务器选项

| 选项 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `WithAddress` | string | 监听地址 | ":50051" |
| `WithTLS` | *tls.Config | TLS 配置 | nil |
| `WithKeepAlive` | *keepalive.ServerParameters | KeepAlive 参数 | 默认配置 |
| `WithUnaryInterceptors` | []grpc.UnaryServerInterceptor | 一元拦截器 | 空切片 |
| `WithStreamInterceptors` | []grpc.StreamServerInterceptor | 流式拦截器 | 空切片 |
| `WithUnaryMiddlewares` | []UnaryServerMiddleware | 一元中间件 | 空切片 |

### 连接池配置

| 配置项 | 类型 | 描述 | 默认值 |
|--------|------|------|--------|
| `MinConnsPerAddr` | int | 每个地址最小连接数 | 2 |
| `MaxConnsPerAddr` | int | 每个地址最大连接数 | 100 |
| `MaxIdleConns` | int | 最大空闲连接数 | 10 |
| `MaxLoadPerConn` | int | 每个连接最大负载 | 1000 |
| `ConnMaxLifetime` | time.Duration | 连接最大生命周期 | 1h |
| `ConnMaxIdleTime` | time.Duration | 连接最大空闲时间 | 30m |

## 🔧 配置说明

### 性能调优
- **连接池大小**: 根据并发量和服务器资源调整
- **消息大小限制**: 根据业务需求设置合适的消息大小
- **KeepAlive 参数**: 根据网络环境调整保活策略
- **超时设置**: 根据业务 SLA 要求设置合适的超时时间

### 安全配置
- **TLS 版本**: 建议使用 TLS 1.2 或更高版本
- **证书管理**: 定期更新证书，使用强加密算法
- **认证策略**: 实现适当的认证和授权机制

## 🤝 贡献指南

我们欢迎社区贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发环境设置
```bash
# 克隆仓库
git clone https://github.com/stones-hub/taurus-pro-grpc.git
cd taurus-pro-grpc

# 安装依赖
go mod download

# 运行测试
go test ./...

# 生成代码
cd bin/proto && ./generate.sh
```

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 👥 作者

- **yelei** - *初始开发者* - [61647649@qq.com](mailto:61647649@qq.com)

## 🙏 致谢

感谢以下开源项目的支持：

- [gRPC](https://grpc.io/) - 高性能 RPC 框架
- [Protocol Buffers](https://developers.google.com/protocol-buffers) - 数据序列化格式
- [OpenTelemetry](https://opentelemetry.io/) - 可观测性框架
- [validator](https://github.com/go-playground/validator) - 数据验证库

## 📞 支持与反馈

如果您在使用过程中遇到问题或有改进建议，请：

1. 查看 [Issues](https://github.com/stones-hub/taurus-pro-grpc/issues) 页面
2. 创建新的 Issue 描述问题
3. 发送邮件至 [61647649@qq.com](mailto:61647649@qq.com)

---

**Taurus Pro gRPC** - 让 gRPC 开发更简单、更高效、更安全！ 🚀
