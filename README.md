# gRPC Communication Patterns - Lightning Fast APIs ⚡

This project demonstrates all 4 gRPC communication patterns with working examples.

## 🚀 Quick Start

### 1. Start the Server
```bash
cd server
go run main.go
```

### 2. Run the Client (Interactive Menu)
```bash
cd client
go run main.go
```

## 📚 Four gRPC Communication Patterns

### 1. Unary RPC (Simple Request-Response)
**Pattern:** Client sends ONE request → Server sends ONE response

**Use Cases:**
- Simple API calls (like REST)
- User authentication
- CRUD operations
- Quick data lookups

**Example:**
```go
// Client sends
HelloRequest{Name: "Alice"}

// Server responds
HelloReply{Message: "Hello Alice"}
```

**When to use:** Standard API operations, similar to REST endpoints

**Performance:** ~10x faster than REST/JSON due to Protocol Buffers

---

### 2. Server Streaming RPC
**Pattern:** Client sends ONE request → Server sends MULTIPLE responses (stream)

**Use Cases:**
- Live stock prices
- Real-time notifications
- Log streaming
- Live sports scores
- Downloading large datasets

**Example:**
```go
// Client sends once
HelloRequest{Name: "Bob"}

// Server sends multiple responses
HelloReply{Message: "Hello Bob - Message 1 of 5"}
HelloReply{Message: "Hello Bob - Message 2 of 5"}
HelloReply{Message: "Hello Bob - Message 3 of 5"}
...
```

**When to use:** 
- Real-time data feeds
- Server needs to push updates to client
- Large file downloads split into chunks

**Benefits:**
- Efficient for continuous data flow
- Client doesn't need to poll
- Single connection for multiple messages

---

### 3. Client Streaming RPC
**Pattern:** Client sends MULTIPLE requests (stream) → Server sends ONE response

**Use Cases:**
- File uploads
- Batch data insertion
- IoT sensor data collection
- Audio/video uploads
- Log aggregation

**Example:**
```go
// Client sends multiple requests
HelloRequest{Name: "Alice"}
HelloRequest{Name: "Bob"}
HelloRequest{Name: "Charlie"}

// Server responds once after receiving all
HelloReply{Message: "Hello to all: Alice, Bob, Charlie! (Total: 3 people)"}
```

**When to use:**
- Uploading large files in chunks
- Batch processing
- Aggregating data from client

**Benefits:**
- Efficient file uploads
- Reduces connection overhead
- Server can process data as it arrives

---

### 4. Bidirectional Streaming RPC
**Pattern:** Both client and server send MULTIPLE messages independently

**Use Cases:**
- Chat applications
- Real-time collaboration (Google Docs style)
- Online multiplayer games
- Video conferencing
- Live trading platforms

**Example:**
```go
// Client and server send messages independently

Client → Server: HelloRequest{Name: "Alice"}
Server → Client: HelloReply{Message: "Echo: Hello Alice!"}
Client → Server: HelloRequest{Name: "How are you?"}
Server → Client: HelloReply{Message: "Echo: Hello How are you?!"}
```

**When to use:**
- Real-time bidirectional communication
- Chat systems
- Live collaboration
- Gaming

**Benefits:**
- Full duplex communication
- Lowest latency
- Both parties can send/receive anytime

---

## 🎯 Performance Comparison

| Pattern | Use Case | Connection | Latency |
|---------|----------|------------|---------|
| **Unary** | Simple API | Single | Low |
| **Server Streaming** | Real-time updates | Single | Very Low |
| **Client Streaming** | File upload | Single | Low |
| **Bidirectional** | Chat/Gaming | Single | Lowest |

## 💡 Why gRPC is Lightning Fast

1. **HTTP/2** - Multiplexing, header compression
2. **Protocol Buffers** - Binary serialization (70% smaller than JSON)
3. **Single Connection** - All streams share one TCP connection
4. **No Overhead** - Direct binary communication
5. **Efficient Streaming** - Native support for real-time data

### Speed Comparison
```
REST/JSON:  ████████████████████ (100ms)
gRPC:       ████ (20ms)

5-10x faster in most scenarios!
```

## 📖 Testing the Patterns

### Test Pattern 1: Unary RPC
```bash
# Run client and select option 1
# Enter a name when prompted
# Expect: Immediate response
```

### Test Pattern 2: Server Streaming
```bash
# Run client and select option 2
# Enter a name when prompted
# Expect: 5 messages received over 5 seconds
```

### Test Pattern 3: Client Streaming
```bash
# Run client and select option 3
# Enter multiple names
# Type 'done' when finished
# Expect: Single aggregated response
```

### Test Pattern 4: Bidirectional Streaming
```bash
# Run client and select option 4
# Type messages back and forth
# Expect: Real-time echo responses
# Type 'exit' to quit
```

## 🛠️ Project Structure

```
grpc-example/
├── proto/
│   ├── helloworld.proto          # Protocol Buffer definition
│   ├── helloworld.pb.go          # Generated Go code
│   └── helloworld_grpc.pb.go     # Generated gRPC code
├── server/
│   └── main.go                   # Server with all 4 patterns
├── client/
│   └── main.go                   # Interactive client
├── go.mod
└── README.md
```

## 🔧 Proto Definition

```protobuf
service Greeter {
  // 1. Unary
  rpc SayHello (HelloRequest) returns (HelloReply) {}
  
  // 2. Server Streaming
  rpc SayHelloServerStream (HelloRequest) returns (stream HelloReply) {}
  
  // 3. Client Streaming
  rpc SayHelloClientStream (stream HelloRequest) returns (HelloReply) {}
  
  // 4. Bidirectional
  rpc SayHelloBidirectional (stream HelloRequest) returns (stream HelloReply) {}
}
```

## 🎓 Key Concepts

### Stream vs Unary
- **Unary:** Like a function call - one in, one out
- **Stream:** Like a pipe - continuous flow of data

### When to Use Each Pattern

| Need | Pattern |
|------|---------|
| Simple API call | Unary |
| Server pushes updates | Server Streaming |
| Upload large data | Client Streaming |
| Real-time chat | Bidirectional |

### gRPC vs REST

| Feature | REST | gRPC |
|---------|------|------|
| Protocol | HTTP/1.1 | HTTP/2 |
| Format | JSON | Protocol Buffers |
| Streaming | Limited | Native |
| Speed | 1x | 5-10x |
| Size | 100% | 30% |

## 🚀 Real-World Applications

### Unary RPC
- Authentication APIs
- CRUD operations
- Payment processing
- User management

### Server Streaming
- Stock tickers
- News feeds
- IoT sensor monitoring
- Live sports updates

### Client Streaming
- File uploads
- Batch inserts
- Telemetry data
- Log shipping

### Bidirectional Streaming
- WhatsApp/Slack (chat)
- Google Docs (collaboration)
- Multiplayer games
- Video calls (WebRTC signaling)

## 📝 Regenerating Proto Files

If you modify `proto/helloworld.proto`:

```bash
protoc --go_out=. --go-grpc_out=. proto/helloworld.proto
```

## 🐛 Troubleshooting

### Port already in use
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

### Connection refused
- Make sure server is running first
- Check server is listening on :8080

## 📚 Learn More

- [gRPC Official Docs](https://grpc.io/docs/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)

## 🎉 Summary

You now have a complete working example of all 4 gRPC patterns!

**Next Steps:**
1. Run the server: `cd server && go run main.go`
2. Run the client: `cd client && go run main.go`
3. Try all 4 patterns
4. Build your own lightning-fast APIs!

---

**Happy coding! ⚡**

