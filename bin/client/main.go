package main

import (
	"context"
	"log"

	pb "github.com/stones-hub/taurus-pro-grpc/bin/proto"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/client"
)

func main() {
	// 创建客户端选项，使用不安全连接
	opts := []client.ClientOption{
		client.WithInsecure(), // 禁用 TLS
	}

	// 创建客户端
	c, err := client.NewClient(opts...)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer c.Close()

	// 获取连接
	conn, err := c.GetConn("localhost:50051", false)
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer c.ReleaseConn(conn)

	// 创建 Hello 客户端
	helloClient := pb.NewHelloClient(conn)

	// 调用 SayHello
	ctx := context.Background()
	res, err := helloClient.SayHello(ctx, &pb.HelloRequest{Name: "张三"})
	if err != nil {
		log.Fatalf("调用 SayHello 失败: %v", err)
	}
	log.Printf("服务器响应: %s", res.Message)
}
