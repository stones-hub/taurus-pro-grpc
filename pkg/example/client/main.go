package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/stones-hub/taurus-pro-grpc/pkg/example/proto/echo"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/client"
	"google.golang.org/grpc/keepalive"
)

func loadTLSConfig() (*tls.Config, error) {
	// 加载 CA 证书
	caCert, err := os.ReadFile("../certs/certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// 加载客户端证书和私钥（这里我们复用服务器的证书和私钥）
	cert, err := tls.LoadX509KeyPair("../certs/certs/server.crt", "../certs/certs/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}, nil
}

func main() {
	// 加载 TLS 配置
	tlsConfig, err := loadTLSConfig()
	if err != nil {
		log.Fatalf("failed to load TLS config: %v", err)
	}

	// 创建客户端选项
	opts := []client.ClientOption{
		client.WithTLS(tlsConfig),
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

	// 创建 Echo 客户端
	echoClient := pb.NewEchoClient(conn)

	// 一元调用示例
	ctx := context.Background()
	res, err := echoClient.UnaryEcho(ctx, &pb.EchoRequest{Message: "Hello, Server!"})
	if err != nil {
		log.Fatalf("failed to call UnaryEcho: %v", err)
	}
	log.Printf("UnaryEcho Response: %s", res.Message)

	// 流式调用示例
	stream, err := echoClient.StreamEcho(ctx)
	if err != nil {
		log.Fatalf("failed to call StreamEcho: %v", err)
	}

	messages := []string{"Hello", "Stream", "World"}
	for _, msg := range messages {
		if err := stream.Send(&pb.EchoRequest{Message: msg}); err != nil {
			log.Fatalf("failed to send message: %v", err)
		}

		res, err := stream.Recv()
		if err != nil {
			log.Fatalf("failed to receive response: %v", err)
		}
		log.Printf("StreamEcho Response: %s", res.Message)
	}

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("failed to close stream: %v", err)
	}
}
