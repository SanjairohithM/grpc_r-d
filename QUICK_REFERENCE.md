# gRPC Patterns Quick Reference 🚀

## Pattern Selection Guide

```
Need a simple API call?          → Use Unary RPC
Need server to push updates?     → Use Server Streaming
Need to upload large data?       → Use Client Streaming  
Need real-time bidirectional?    → Use Bidirectional Streaming
```

## Proto Syntax Quick Reference

```protobuf
// 1. Unary
rpc MethodName (Request) returns (Response) {}

// 2. Server Streaming (add 'stream' to response)
rpc MethodName (Request) returns (stream Response) {}

// 3. Client Streaming (add 'stream' to request)
rpc MethodName (stream Request) returns (Response) {}

// 4. Bidirectional (add 'stream' to both)
rpc MethodName (stream Request) returns (stream Response) {}
```

## Server Implementation Patterns

### 1. Unary
```go
func (s *server) Method(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    return &pb.Response{}, nil
}
```

### 2. Server Streaming
```go
func (s *server) Method(req *pb.Request, stream pb.Service_MethodServer) error {
    for i := 0; i < 5; i++ {
        stream.Send(&pb.Response{})
    }
    return nil
}
```

### 3. Client Streaming
```go
func (s *server) Method(stream pb.Service_MethodServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.Response{})
        }
        // Process req
    }
}
```

### 4. Bidirectional
```go
func (s *server) Method(stream pb.Service_MethodServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        stream.Send(&pb.Response{})
    }
}
```

## Client Implementation Patterns

### 1. Unary
```go
resp, err := client.Method(ctx, &pb.Request{})
```

### 2. Server Streaming
```go
stream, err := client.Method(ctx, &pb.Request{})
for {
    resp, err := stream.Recv()
    if err == io.EOF { break }
}
```

### 3. Client Streaming
```go
stream, err := client.Method(ctx)
for _, item := range items {
    stream.Send(&pb.Request{})
}
resp, err := stream.CloseAndRecv()
```

### 4. Bidirectional
```go
stream, err := client.Method(ctx)

// Receive goroutine
go func() {
    for {
        resp, err := stream.Recv()
    }
}()

// Send
for _, item := range items {
    stream.Send(&pb.Request{})
}
stream.CloseSend()
```

## Common Commands

```bash
# Generate proto files
protoc --go_out=. --go-grpc_out=. proto/*.proto

# Run server
cd server && go run main.go

# Run client
cd client && go run main.go

# Kill port 8080
lsof -ti:8080 | xargs kill -9

# Install dependencies
go mod tidy
```

## Performance Tips

1. **Reuse connections** - Don't create new connection per request
2. **Use streaming** - For real-time or large data
3. **Enable compression** - For large payloads
4. **Set timeouts** - Always use context with timeout
5. **Handle errors** - Check io.EOF for stream end

## Error Handling

```go
// Check for stream end
if err == io.EOF {
    // Stream closed normally
}

// Check for errors
if err != nil {
    // Handle error
}
```

## Testing Checklist

- [ ] Test Unary RPC
- [ ] Test Server Streaming (receive all 5 messages)
- [ ] Test Client Streaming (send multiple, receive summary)
- [ ] Test Bidirectional (chat back and forth)
- [ ] Test error cases
- [ ] Test with large data
- [ ] Check logs on both sides

## Real-World Examples

| Pattern | Example | Why |
|---------|---------|-----|
| Unary | User Login | Simple request/response |
| Server Stream | Stock Prices | Server pushes updates |
| Client Stream | File Upload | Client sends chunks |
| Bidirectional | Chat App | Both send anytime |

---

**Keep this as your reference while building gRPC services!**

