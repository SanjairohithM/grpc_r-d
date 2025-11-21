package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorilla/websocket"
)

var grpcClient pb.GreeterClient

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
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	
	grpcClient = pb.NewGreeterClient(conn)
	
	// Serve static files (public folder is in parent directory)
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	
	// API endpoints
	http.HandleFunc("/api/unary", enableCORS(handleUnary))
	http.HandleFunc("/api/server-stream", enableCORS(handleServerStream))
	http.HandleFunc("/api/client-stream", enableCORS(handleClientStream))
	http.HandleFunc("/api/bidirectional", handleBidirectional)
	
	log.Println("🚀 HTTP Gateway running on http://localhost:3000")
	log.Println("📡 Connected to gRPC server on localhost:8080")
	log.Println("🌐 Open http://localhost:3000 in your browser")
	
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// 1. UNARY RPC - POST /api/unary
func handleUnary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req UnaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("[HTTP Gateway] Unary request: %s", req.Name)
	
	// Call gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	grpcResp, err := grpcClient.SayHello(ctx, &pb.HelloRequest{Name: req.Name})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	resp := UnaryResponse{Message: grpcResp.Message}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
	
	// Call gRPC stream
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	
	// Call gRPC stream
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	
	// Connect to gRPC stream
	ctx, cancel := context.WithCancel(context.Background())
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

