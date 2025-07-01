#!/bin/bash

# 确保脚本在错误时退出
set -e

# 生成 Hello 服务的 protobuf 代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    hello.proto 