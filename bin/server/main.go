package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/stones-hub/taurus-pro-grpc/bin/proto"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/server"
)

// helloServer 实现 Hello 服务
type helloServer struct {
	pb.UnimplementedHelloServer
}

// SayHello 实现一元 Hello
func (s *helloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Message: fmt.Sprintf("你好, %s!", req.Name)}, nil
}

func main() {
	// 创建服务器选项
	opts := []server.ServerOption{
		server.WithAddress(":50051"),
	}

	// 创建 gRPC 服务器
	s, cleanup, err := server.NewServer(opts...)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}
	defer cleanup()

	// 注册 Hello 服务
	pb.RegisterHelloServer(s.Server(), &helloServer{})

	// 启动服务器
	log.Printf("启动 gRPC 服务器，监听端口 :50051")
	if err := s.Start(); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}
