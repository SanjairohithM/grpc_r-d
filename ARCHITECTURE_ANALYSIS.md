# Architecture Analysis

## 🏗️ Current Architecture Overview

Your application follows a **3-tier microservices architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND LAYER (Next.js)                      │
│  Port: 3000                                                      │
│  - React Components (page.tsx)                                  │
│  - API Client (lib/grpc-api.ts)                                 │
│  - UI/UX Layer                                                   │
└───────────────────────┬─────────────────────────────────────────┘
                        │ HTTP/1.1 (REST-like)
                        │ - GET /api/server-stream (SSE)
                        │ - POST /api/client-stream (JSON)
                        │ - WebSocket /api/bidirectional
                        │
┌───────────────────────▼─────────────────────────────────────────┐
│              GATEWAY LAYER (Go HTTP Server)                     │
│  Port: 8081                                                      │
│  - Protocol Translation (HTTP/1.1 → gRPC/HTTP/2)               │
│  - CORS Handling                                                 │
│  - Rate Limiting                                                 │
│  - Request Logging                                               │
│  - Gzip Compression                                              │
│  - Middleware Chain                                               │
└───────────────────────┬─────────────────────────────────────────┘
                        │ gRPC/HTTP/2
                        │ - Unary RPC (disabled)
                        │ - Server Streaming
                        │ - Client Streaming
                        │ - Bidirectional Streaming
                        │
┌───────────────────────▼─────────────────────────────────────────┐
│              gRPC SERVER LAYER (Go)                              │
│  Port: 8080                                                      │
│  - Business Logic                                                │
│  - Database Operations (GORM + PostgreSQL)                           │
│  - Connection Pooling                                            │
│  - In-memory Caching                                             │
│  - Optimized Streaming                                            │
└───────────────────────┬─────────────────────────────────────────┘
                        │ SQL
                        │
┌───────────────────────▼─────────────────────────────────────────┐
│              DATABASE LAYER (PostgreSQL/Supabase)                │
│  - User Management                                               │
│  - Greeting Storage                                               │
│  - Connection Pooling (pgbouncer)                               │
└─────────────────────────────────────────────────────────────────┘
```

## 📊 Communication Patterns

### 1. Server Streaming (One → Many)
```
Frontend → GET /api/server-stream?name=Bob
         ↓ (HTTP/1.1 + SSE)
Gateway → SayHelloServerStream(name: "Bob")
         ↓ (gRPC/HTTP/2)
Server  → Streams 5 messages back
         ↓
Gateway → Converts to SSE format
         ↓
Frontend ← Receives messages via EventSource
```

**Technology Stack:**
- Frontend: `EventSource` API (Server-Sent Events)
- Gateway: SSE headers + streaming response
- Server: gRPC server streaming

### 2. Client Streaming (Many → One)
```
Frontend → POST /api/client-stream
         ↓ (HTTP/1.1 + JSON array)
Gateway → Receives array, sends to gRPC stream
         ↓ (gRPC/HTTP/2 client streaming)
Server  → Processes all names concurrently
         ↓ (Batch operations, goroutines)
Server  → Returns single aggregated response
         ↓
Gateway → Converts to JSON
         ↓
Frontend ← Receives single JSON response
```

**Technology Stack:**
- Frontend: `fetch()` with JSON body
- Gateway: JSON decode → gRPC client stream
- Server: Concurrent processing with goroutines

### 3. Bidirectional Streaming (Many ↔ Many)
```
Frontend → WebSocket connection to /api/bidirectional
         ↓ (WebSocket protocol)
Gateway → Upgrades HTTP to WebSocket
         ↓
Gateway → Creates gRPC bidirectional stream
         ↓ (gRPC/HTTP/2 bidirectional)
Server  → Handles concurrent send/receive
         ↓
Gateway → Bridges WebSocket ↔ gRPC stream
         ↓
Frontend ← Real-time bidirectional communication
```

**Technology Stack:**
- Frontend: `WebSocket` API
- Gateway: `gorilla/websocket` + gRPC bidirectional stream
- Server: Concurrent goroutines for send/receive

## 🔧 Component Details

### Frontend (`frontend/`)
- **Framework**: Next.js 16 (React 19)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v3
- **API Client**: `lib/grpc-api.ts`
  - Uses native browser APIs (EventSource, WebSocket, fetch)
  - No gRPC-Web (uses HTTP gateway instead)
  - Simple, maintainable approach

### Gateway (`gateway/`)
- **Language**: Go
- **Purpose**: Protocol translation layer
- **Key Features**:
  - HTTP/1.1 → gRPC/HTTP/2 translation
  - CORS handling for browser compatibility
  - Rate limiting (100 req/s, burst 200)
  - Request logging with timing
  - Gzip compression (for non-streaming endpoints)
  - Connection pooling to gRPC server
  - Graceful shutdown

**Middleware Chain:**
```
Request → Rate Limit → Gzip → CORS → Logger → Handler
```

**Endpoints:**
- `/health` - Health check
- `/api/server-stream` - Server streaming (SSE)
- `/api/client-stream` - Client streaming (POST JSON)
- `/api/bidirectional` - Bidirectional (WebSocket)

### gRPC Server (`server/`)
- **Language**: Go
- **Protocol**: gRPC/HTTP/2
- **Key Features**:
  - Connection pooling (25 idle, 100 max)
  - Keepalive settings
  - In-memory user caching (`sync.Map`)
  - Concurrent processing (goroutines)
  - Batch database operations
  - Graceful shutdown

**RPC Methods:**
- `SayHello` - Unary (disabled)
- `SayHelloServerStream` - Server streaming
- `SayHelloClientStream` - Client streaming
- `SayHelloBidirectional` - Bidirectional streaming

### Database (`server/database.go`)
- **ORM**: GORM
- **Database**: PostgreSQL (via Supabase)
- **Features**:
  - Connection pooling
  - Auto-migration
  - Indexes on User.name and User.email
  - Optimized queries

## ⚡ Performance Optimizations

### Gateway:
1. **Connection Pooling**: Reuses gRPC connections
2. **Gzip Compression**: 70-90% size reduction
3. **Rate Limiting**: Prevents abuse
4. **Request Logging**: Performance monitoring

### Server:
1. **In-Memory Caching**: `sync.Map` for user lookups
2. **Concurrent Processing**: Goroutines for parallel operations
3. **Batch Inserts**: `CreateInBatches` for database writes
4. **Connection Pooling**: Database connection reuse
5. **Indexes**: Fast database queries

### Database:
1. **Connection Pooling**: pgbouncer (via Supabase)
2. **Indexes**: Unique indexes on User.name and User.email
3. **Optimized Queries**: GORM with connection limits

## 🔄 Data Flow Examples

### Example 1: Server Streaming
```
User clicks "Start Streaming" with name "Bob"
  ↓
Frontend: EventSource("http://localhost:8081/api/server-stream?name=Bob")
  ↓
Gateway: Receives GET request, calls gRPC SayHelloServerStream
  ↓
Server: Streams 5 messages over 5 seconds
  ↓
Gateway: Converts each gRPC message to SSE format
  ↓
Frontend: Receives messages via EventSource.onmessage
```

### Example 2: Client Streaming
```
User adds names: ["Alice", "Bob", "Charlie"] and clicks "Send All"
  ↓
Frontend: POST /api/client-stream with JSON array
  ↓
Gateway: Decodes JSON, creates gRPC client stream
  ↓
Server: Receives all names, processes concurrently with goroutines
  ↓
Server: Batch creates users, inserts greetings asynchronously
  ↓
Server: Returns aggregated response
  ↓
Gateway: Converts to JSON
  ↓
Frontend: Displays result
```

### Example 3: Bidirectional
```
User clicks "Connect" for bidirectional chat
  ↓
Frontend: WebSocket("ws://localhost:8081/api/bidirectional")
  ↓
Gateway: Upgrades HTTP to WebSocket, creates gRPC bidirectional stream
  ↓
Server: Sets up concurrent send/receive goroutines
  ↓
User types message "Hello"
  ↓
Frontend: ws.send(JSON.stringify({name: "Hello"}))
  ↓
Gateway: Receives WebSocket message, sends to gRPC stream
  ↓
Server: Receives message, processes, sends echo response
  ↓
Gateway: Receives gRPC response, sends to WebSocket
  ↓
Frontend: Receives message via ws.onmessage
```

## 🐛 Issues Fixed

### 1. Bidirectional WebSocket CORS
**Problem**: WebSocket upgrade requests need CORS headers
**Solution**: Added `enableCORS` middleware to `/api/bidirectional` route

### 2. Server Streaming CORS
**Problem**: SSE endpoints need proper CORS headers
**Solution**: Enhanced CORS middleware with localhost support

### 3. Health Check CORS
**Problem**: Health endpoint didn't have CORS
**Solution**: Added CORS middleware to `/health` endpoint

## 📈 Scalability Considerations

### Current Architecture:
- **Single Gateway Instance**: Can handle ~100 req/s (rate limited)
- **Single gRPC Server**: Can handle 1000 concurrent streams
- **Database**: Connection pool of 100 max connections

### Scaling Options:
1. **Horizontal Scaling**: Multiple gateway instances behind load balancer
2. **gRPC Server Clustering**: Multiple server instances
3. **Database Read Replicas**: For read-heavy workloads
4. **Caching Layer**: Redis for frequently accessed data
5. **CDN**: For static frontend assets

## 🔐 Security Features

1. **CORS**: Restricts origins to localhost (development)
2. **Rate Limiting**: Prevents abuse
3. **Input Validation**: Gateway validates requests
4. **Connection Limits**: Prevents resource exhaustion
5. **Graceful Shutdown**: Clean connection termination

## 🚀 Deployment Architecture

### Development:
```
localhost:3000 (Next.js) → localhost:8081 (Gateway) → localhost:8080 (gRPC)
```

### Production (Recommended):
```
CDN/Edge → Load Balancer → [Gateway Instances] → [gRPC Server Instances] → Database Cluster
```

## 📝 Key Design Decisions

1. **Gateway Pattern**: Chosen for browser compatibility (browsers can't use gRPC directly)
2. **SSE for Server Streaming**: Simpler than WebSocket for one-way streaming
3. **WebSocket for Bidirectional**: Required for true bidirectional communication
4. **JSON over WebSocket**: Simpler than protobuf for frontend
5. **Go for Backend**: High performance, excellent concurrency support

## ✅ Architecture Strengths

1. **Clear Separation**: Each layer has distinct responsibilities
2. **Protocol Flexibility**: Gateway handles protocol translation
3. **Performance Optimized**: Connection pooling, caching, batching
4. **Scalable**: Can scale each layer independently
5. **Maintainable**: Clear code organization

## 🔄 Future Improvements

1. **gRPC-Web**: Direct browser-to-gRPC (requires proto generation)
2. **Service Mesh**: For advanced routing and observability
3. **GraphQL Gateway**: Alternative API layer
4. **WebSocket Cluster**: For horizontal WebSocket scaling
5. **Message Queue**: For async processing

