package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "grpc-example/proto"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

type UnaryRequest struct {
	Name string `json:"name"`
}

type UnaryResponse struct {
	Message string `json:"message"`
}

func main() {
	// ⚡ Initialize optimized gRPC connection with pooling
	if err := initGRPCConnection(); err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer closeGRPCConnection()
	
	// Create HTTP server with optimizations
	srv := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	
	// ⚡ Apply middleware chain: Rate Limit → Gzip → CORS → Logger → Handler
	http.HandleFunc("/api/unary", 
		rateLimitMiddleware(
			enableGzip(
				enableCORS(
					requestLogger(handleUnary),
				),
			),
		),
	)
	
	http.HandleFunc("/api/server-stream", 
		rateLimitMiddleware(
			enableCORS(
				requestLogger(handleServerStream),
			),
		),
	)
	
	http.HandleFunc("/api/client-stream", 
		rateLimitMiddleware(
			enableGzip(
				enableCORS(
					requestLogger(handleClientStream),
				),
			),
		),
	)
	
	http.HandleFunc("/api/bidirectional", 
		rateLimitMiddleware(
			requestLogger(handleBidirectional),
		),
	)
	
	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	
	log.Println("🚀 HTTP Gateway (API) running on http://localhost:8081")
	log.Println("📡 Connected to gRPC server on localhost:8080 with connection pooling")
	log.Println("⚡ Optimizations: Gzip, Rate Limiting, Connection Pooling, Request Logging")
	log.Println("🔗 CORS enabled for Next.js on http://localhost:3000")
	
	// ⚡ Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Shutting down server gracefully...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("✅ Server exited gracefully")
}

// CORS middleware - configured for Next.js
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow Next.js frontend (port 3000)
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:3001" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// 1. UNARY RPC - POST /api/unary
// ⛔ DISABLED: This endpoint has been disabled
func handleUnary(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP Gateway] ⛔ Unary API access blocked - Service disabled")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	
	errorResp := map[string]string{
		"error":   "Service Unavailable",
		"message": "Unary API endpoint has been disabled",
		"status":  "503",
	}
	json.NewEncoder(w).Encode(errorResp)
}

// 2. SERVER STREAMING RPC - GET /api/server-stream?name=xxx
// Uses Server-Sent Events (SSE)
func handleServerStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Guest"
	}
	
	log.Printf("[HTTP Gateway] Server streaming request: %s", name)
	
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	
	// ⚡ Use request context with timeout (better resource management)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	stream, err := grpcClient.SayHelloServerStream(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Stream messages to client
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Fprintf(w, "event: done\ndata: {\"message\": \"Stream complete\"}\n\n")
			flusher.Flush()
			break
		}
		if err != nil {
			log.Printf("Stream error: %v", err)
			break
		}
		
		data := map[string]string{"message": msg.Message}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
	}
}

// 3. CLIENT STREAMING RPC - POST /api/client-stream
func handleClientStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var names []string
	if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("[HTTP Gateway] Client streaming request with %d names", len(names))
	
	// ⚡ Use request context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	stream, err := grpcClient.SayHelloClientStream(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Send all names
	for _, name := range names {
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	// Get response
	grpcResp, err := stream.CloseAndRecv()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	resp := UnaryResponse{Message: grpcResp.Message}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// 4. BIDIRECTIONAL STREAMING RPC - WebSocket /api/bidirectional
func handleBidirectional(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()
	
	log.Println("[HTTP Gateway] Bidirectional WebSocket connection established")
	
	// ⚡ Use request context for better cancellation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	
	stream, err := grpcClient.SayHelloBidirectional(ctx)
	if err != nil {
		log.Printf("gRPC stream error: %v", err)
		return
	}
	
	// Goroutine to receive from gRPC and send to WebSocket
	go func() {
		for {
			grpcResp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("gRPC receive error: %v", err)
				return
			}
			
			data := map[string]string{"message": grpcResp.Message}
			if err := ws.WriteJSON(data); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()
	
	// Receive from WebSocket and send to gRPC
	for {
		var msg map[string]string
		if err := ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		name := msg["name"]
		if name == "" {
			continue
		}
		
		log.Printf("[HTTP Gateway] Bidirectional message: %s", name)
		
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Printf("gRPC send error: %v", err)
			break
		}
	}
	
	stream.CloseSend()
}

