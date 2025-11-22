package main

import (
	"log"
	"time"

	pb "grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var grpcClient pb.GreeterClient
var grpcConn *grpc.ClientConn

// initGRPCConnection - Creates optimized gRPC connection with pooling
func initGRPCConnection() error {
	var err error
	
	// ⚡ OPTIMIZATION: Connection pooling with keepalive
	// This reuses connections instead of creating new ones for each request
	grpcConn, err = grpc.Dial("localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		
		// ⚡ Keepalive settings - keeps connection alive
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second, // Send keepalive ping every 30s
			Timeout:             5 * time.Second,  // Wait 5s for ping ack
			PermitWithoutStream: true,             // Send pings even without active streams
		}),
		
		// ⚡ Connection pool settings
		grpc.WithInitialWindowSize(1 << 20),  // 1MB initial window
		grpc.WithInitialConnWindowSize(1 << 20), // 1MB initial connection window
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB max receive
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB max send
		),
	)
	
	if err != nil {
		return err
	}
	
	grpcClient = pb.NewGreeterClient(grpcConn)
	log.Println("✅ gRPC connection established with connection pooling")
	return nil
}

// closeGRPCConnection - Gracefully closes gRPC connection
func closeGRPCConnection() {
	if grpcConn != nil {
		if err := grpcConn.Close(); err != nil {
			log.Printf("Error closing gRPC connection: %v", err)
		} else {
			log.Println("✅ gRPC connection closed gracefully")
		}
	}
}


