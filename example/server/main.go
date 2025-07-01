package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/stones-hub/taurus-pro-grpc/example/proto/echo"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/server"
	"google.golang.org/grpc/keepalive"
)

// echoServer 实现 Echo 服务
type echoServer struct {
	pb.UnimplementedEchoServer
}

// UnaryEcho 实现一元 Echo
func (s *echoServer) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: fmt.Sprintf("Echo: %s", req.Message)}, nil
}

// StreamEcho 实现流式 Echo
func (s *echoServer) StreamEcho(stream pb.Echo_StreamEchoServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		res := &pb.EchoResponse{Message: fmt.Sprintf("Stream Echo: %s", req.Message)}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func loadTLSConfig() (*tls.Config, error) {
	// 加载服务器证书和私钥
	cert, err := tls.LoadX509KeyPair("../certs/certs/server.crt", "../certs/certs/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	// 加载 CA 证书用于验证客户端证书
	caCert, err := os.ReadFile("../certs/certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}, nil
}

func main() {
	// 加载 TLS 配置
	tlsConfig, err := loadTLSConfig()
	if err != nil {
		log.Fatalf("failed to load TLS config: %v", err)
	}

	// 创建服务器选项
	opts := []server.ServerOption{
		server.WithAddress(":50051"),
		server.WithTLS(tlsConfig),
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

	// 注册 Echo 服务
	pb.RegisterEchoServer(s.Server(), &echoServer{})

	// 启动服务器
	log.Printf("Starting Echo server with TLS on :50051")
	if err := s.Start(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
