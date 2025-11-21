package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "grpc-example/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// 1. UNARY RPC - Simple request/response
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("[Unary] Received request from: %s", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

// 2. SERVER STREAMING RPC - One request, multiple responses from server
func (s *server) SayHelloServerStream(in *pb.HelloRequest, stream pb.Greeter_SayHelloServerStreamServer) error {
	log.Printf("[Server Streaming] Received request from: %s", in.Name)
	
	// Send multiple responses to the client
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Hello %s - Message %d of 5", in.Name, i)
		response := &pb.HelloReply{Message: msg}
		
		if err := stream.Send(response); err != nil {
			return err
		}
		
		log.Printf("[Server Streaming] Sent: %s", msg)
		time.Sleep(1 * time.Second) // Simulate real-time data
	}
	
	return nil
}

// 3. CLIENT STREAMING RPC - Client sends multiple requests, one response
func (s *server) SayHelloClientStream(stream pb.Greeter_SayHelloClientStreamServer) error {
	log.Printf("[Client Streaming] Waiting for client messages...")
	
	var names []string
	
	// Receive multiple messages from client
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			log.Printf("[Client Streaming] Client finished. Received %d names", len(names))
			
			// Send single response
			allNames := ""
			for i, name := range names {
				if i > 0 {
					allNames += ", "
				}
				allNames += name
			}
			
			return stream.SendAndClose(&pb.HelloReply{
				Message: fmt.Sprintf("Hello to all: %s! (Total: %d people)", allNames, len(names)),
			})
		}
		if err != nil {
			return err
		}
		
		log.Printf("[Client Streaming] Received: %s", req.Name)
		names = append(names, req.Name)
	}
}

// 4. BIDIRECTIONAL STREAMING RPC - Both client and server send multiple messages
func (s *server) SayHelloBidirectional(stream pb.Greeter_SayHelloBidirectionalServer) error {
	log.Printf("[Bidirectional] Starting bidirectional stream...")
	
	for {
		// Receive message from client
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[Bidirectional] Client closed the stream")
			return nil
		}
		if err != nil {
			return err
		}
		
		log.Printf("[Bidirectional] Received from client: %s", req.Name)
		
		// Send immediate response
		response := &pb.HelloReply{
			Message: fmt.Sprintf("Echo: Hello %s! (received at %s)", req.Name, time.Now().Format("15:04:05")),
		}
		
		if err := stream.Send(response); err != nil {
			return err
		}
		
		log.Printf("[Bidirectional] Sent response to: %s", req.Name)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	fmt.Println("gRPC server listening on :8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
