package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	pb "grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type server struct {
	pb.UnimplementedGreeterServer
	db *gorm.DB
}

// 1. UNARY RPC - ⛔ DISABLED: This endpoint has been disabled
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("[Unary] ⛔ Access blocked - Service disabled for: %s", in.Name)
	
	// Return gRPC error status
	return nil, status.Errorf(codes.Unimplemented, "Unary API endpoint has been disabled")
}

// 2. SERVER STREAMING RPC - OPTIMIZED: One request, multiple responses from server
func (s *server) SayHelloServerStream(in *pb.HelloRequest, stream pb.Greeter_SayHelloServerStreamServer) error {
	startTime := time.Now()
	log.Printf("[Server Streaming] 📥 Received request from: %s", in.Name)
	
	// ⚡ OPTIMIZATION: Check context for cancellation
	ctx := stream.Context()
	
	// Send multiple responses to the client
	for i := 1; i <= 5; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			log.Printf("[Server Streaming] ⚠️  Context cancelled")
			return ctx.Err()
		default:
		}
		
		msg := fmt.Sprintf("Hello %s - Message %d of 5", in.Name, i)
		response := &pb.HelloReply{Message: msg}
		
		if err := stream.Send(response); err != nil {
			log.Printf("[Server Streaming] ❌ Send error: %v", err)
			return err
		}
		
		time.Sleep(1 * time.Second) // Simulate real-time data
	}
	
	duration := time.Since(startTime)
	log.Printf("[Server Streaming] ⚡ Completed in %v", duration)
	return nil
}

// 3. CLIENT STREAMING RPC - OPTIMIZED: Batch operations for multiple users
func (s *server) SayHelloClientStream(stream pb.Greeter_SayHelloClientStreamServer) error {
	startTime := time.Now()
	log.Printf("[Client Streaming] 📥 Waiting for client messages...")
	
	var names []string
	
	// Receive multiple messages from client
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			log.Printf("[Client Streaming] ✅ Received %d names", len(names))
			
			// ⚡ OPTIMIZATION: Process all users concurrently with goroutines
			var wg sync.WaitGroup
			userChan := make(chan *User, len(names))
			
			for _, name := range names {
				wg.Add(1)
				go func(n string) {
					defer wg.Done()
					user, err := GetOrCreateUser(s.db, n)
					if err == nil {
						userChan <- user
					}
				}(name)
			}
			
			// Wait for all users to be processed
			go func() {
				wg.Wait()
				close(userChan)
			}()
			
			// Collect user IDs
			var users []*User
			for user := range userChan {
				users = append(users, user)
			}
			
			// ⚡ OPTIMIZATION: Batch insert greetings asynchronously
			go func() {
				greetings := make([]Greeting, len(users))
				for i, user := range users {
					greetings[i] = Greeting{
						Message: fmt.Sprintf("Hello %s", user.Name),
						UserID:  &user.ID,
					}
				}
				if len(greetings) > 0 {
					s.db.CreateInBatches(greetings, 100) // Batch insert 100 at a time
				}
			}()
			
			// Build response
			allNames := strings.Join(names, ", ")
			totalTime := time.Since(startTime)
			
			log.Printf("[Client Streaming] ⚡ Processed %d users in %v", len(names), totalTime)
			
			return stream.SendAndClose(&pb.HelloReply{
				Message: fmt.Sprintf("Hello to all: %s! (Total: %d people, %v)", allNames, len(names), totalTime),
			})
		}
		if err != nil {
			return err
		}
		
		log.Printf("[Client Streaming] 📨 Received: %s", req.Name)
		names = append(names, req.Name)
	}
}

// 4. BIDIRECTIONAL STREAMING RPC - OPTIMIZED: Both client and server send multiple messages
func (s *server) SayHelloBidirectional(stream pb.Greeter_SayHelloBidirectionalServer) error {
	log.Printf("[Bidirectional] 📥 Starting bidirectional stream...")
	ctx := stream.Context()
	
	// ⚡ OPTIMIZATION: Use goroutine for concurrent send/receive
	recvChan := make(chan *pb.HelloRequest, 10)
	errChan := make(chan error, 1)
	
	// Receive goroutine
	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				close(recvChan)
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			recvChan <- req
		}
	}()
	
	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Bidirectional] ⚠️  Context cancelled")
			return ctx.Err()
			
		case err := <-errChan:
			if err != nil {
				log.Printf("[Bidirectional] ❌ Receive error: %v", err)
				return err
			}
			
		case req, ok := <-recvChan:
			if !ok {
				log.Printf("[Bidirectional] ✅ Client closed the stream")
				return nil
			}
			
			// Send immediate response
			response := &pb.HelloReply{
				Message: fmt.Sprintf("Echo: Hello %s! (received at %s)", req.Name, time.Now().Format("15:04:05")),
			}
			
			if err := stream.Send(response); err != nil {
				log.Printf("[Bidirectional] ❌ Send error: %v", err)
				return err
			}
		}
	}
}

func main() {
	// Load .env file from project root (parent directory)
	envPath := filepath.Join("..", ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		} else {
			log.Println("✅ Loaded .env file")
		}
	} else {
		// Try current directory
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	}
	
	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()
	
	// Note: We use HTTP server for gRPC-Web, which internally uses the gRPC server
	// No need for separate listener - grpcweb handles it
	
	// ⚡ OPTIMIZED gRPC Server with keepalive and performance settings
	srv := grpc.NewServer(
		// ⚡ Keepalive enforcement - prevents dead connections
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // Minimum time between pings
			PermitWithoutStream: true,            // Allow pings without active streams
		}),
		
		// ⚡ Keepalive parameters - keeps connections alive
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute, // Close idle connections after 15min
			MaxConnectionAge:      30 * time.Minute, // Close connections after 30min
			MaxConnectionAgeGrace: 5 * time.Second,  // Grace period for closing
			Time:                  5 * time.Second,  // Send keepalive ping every 5s
			Timeout:               1 * time.Second, // Wait 1s for ping ack
		}),
		
		// ⚡ Message size limits
		grpc.MaxRecvMsgSize(4*1024*1024),  // 4MB max receive
		grpc.MaxSendMsgSize(4*1024*1024),  // 4MB max send
		grpc.MaxConcurrentStreams(1000),   // Max concurrent streams
	)
	
	pb.RegisterGreeterServer(srv, &server{db: DB})
	
	// ⚡ Wrap gRPC server with gRPC-Web support for browser clients
	wrappedServer := grpcweb.WrapServer(srv,
		grpcweb.WithOriginFunc(func(origin string) bool {
			// Allow requests from Next.js frontend
			return origin == "http://localhost:3000" || 
			       origin == "http://localhost:3001" ||
			       origin == "" // Allow same-origin requests
		}),
		grpcweb.WithWebsockets(true), // Enable WebSocket support for bidirectional streaming
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true // Allow WebSocket connections
		}),
	)
	
	// Create HTTP server that serves gRPC-Web (for browsers)
	// Regular gRPC clients can still connect directly to the gRPC server
	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Serve gRPC-Web requests
			wrappedServer.ServeHTTP(w, r)
		}),
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start gRPC server in a separate goroutine (for native clients)
	// Note: This won't work on the same port, so we'll use HTTP server for both
	// The grpcweb wrapper handles both gRPC-Web and can proxy to gRPC
	
	fmt.Println("⚡ gRPC + gRPC-Web server listening on :8080")
	fmt.Println("✅ Database connected and ready")
	fmt.Println("⚡ Optimizations: Keepalive, Connection Pooling, Max Streams: 1000")
	fmt.Println("🌐 gRPC-Web enabled for browser clients (no gateway needed!)")
	
	// ⚡ Graceful shutdown
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Shutting down server gracefully...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	
	// Also stop gRPC server gracefully
	srv.GracefulStop()
	
	log.Println("✅ Server exited gracefully")
}
