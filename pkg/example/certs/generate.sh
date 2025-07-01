#!/bin/bash

# 确保脚本在错误时退出
set -e

# 创建证书目录
mkdir -p certs

# 生成私钥和自签名证书
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost" -addext "subjectAltName = DNS:localhost"

# 复制证书作为客户端的CA证书
cp certs/server.crt certs/ca.crt

# 设置适当的权限
chmod 644 certs/server.crt certs/ca.crt
chmod 600 certs/server.key 