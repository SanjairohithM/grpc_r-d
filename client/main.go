package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to server
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	
	client := pb.NewGreeterClient(conn)
	
	// Interactive menu
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Println("\n========================================")
		fmt.Println("  gRPC Communication Patterns Demo")
		fmt.Println("========================================")
		fmt.Println("1. Unary RPC (Simple Request-Response)")
		fmt.Println("2. Server Streaming RPC")
		fmt.Println("3. Client Streaming RPC")
		fmt.Println("4. Bidirectional Streaming RPC")
		fmt.Println("5. Exit")
		fmt.Print("\nSelect option (1-5): ")
		
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			testUnaryRPC(client)
		case "2":
			testServerStreamingRPC(client)
		case "3":
			testClientStreamingRPC(client)
		case "4":
			testBidirectionalStreamingRPC(client)
		case "5":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option. Please select 1-5")
		}
	}
}

// 1. UNARY RPC - Simple request/response
func testUnaryRPC(client pb.GreeterClient) {
	fmt.Println("\n--- Testing Unary RPC ---")
	fmt.Print("Enter your name: ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	response, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("✓ Response: %s\n", response.Message)
}

// 2. SERVER STREAMING RPC - Server sends multiple responses
func testServerStreamingRPC(client pb.GreeterClient) {
	fmt.Println("\n--- Testing Server Streaming RPC ---")
	fmt.Print("Enter your name: ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	stream, err := client.SayHelloServerStream(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Println("Receiving messages from server...")
	
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("✓ Server finished sending messages")
			break
		}
		if err != nil {
			log.Printf("Error receiving: %v", err)
			return
		}
		
		fmt.Printf("✓ Received: %s\n", response.Message)
	}
}

// 3. CLIENT STREAMING RPC - Client sends multiple requests
func testClientStreamingRPC(client pb.GreeterClient) {
	fmt.Println("\n--- Testing Client Streaming RPC ---")
	fmt.Println("Enter names (type 'done' to finish):")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	stream, err := client.SayHelloClientStream(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	count := 0
	
	for {
		fmt.Print("Name: ")
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		
		if name == "done" {
			break
		}
		
		if name == "" {
			continue
		}
		
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Printf("Error sending: %v", err)
			return
		}
		
		count++
		fmt.Printf("✓ Sent: %s (%d)\n", name, count)
	}
	
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error receiving response: %v", err)
		return
	}
	
	fmt.Printf("\n✓ Server Response: %s\n", response.Message)
}

// 4. BIDIRECTIONAL STREAMING RPC - Both send multiple messages
func testBidirectionalStreamingRPC(client pb.GreeterClient) {
	fmt.Println("\n--- Testing Bidirectional Streaming RPC ---")
	fmt.Println("Chat mode activated! Type messages (type 'exit' to quit)")
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	stream, err := client.SayHelloBidirectional(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Goroutine to receive messages from server
	waitc := make(chan struct{})
	go func() {
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Error receiving: %v", err)
				close(waitc)
				return
			}
			
			fmt.Printf("\n✓ Server: %s\n", response.Message)
			fmt.Print("You: ")
		}
	}()
	
	// Send messages to server
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		scanner.Scan()
		message := strings.TrimSpace(scanner.Text())
		
		if message == "exit" {
			stream.CloseSend()
			break
		}
		
		if message == "" {
			continue
		}
		
		if err := stream.Send(&pb.HelloRequest{Name: message}); err != nil {
			log.Printf("Error sending: %v", err)
			return
		}
	}
	
	<-waitc
	fmt.Println("✓ Chat session ended")
}
