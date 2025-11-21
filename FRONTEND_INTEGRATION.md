# Frontend Integration with gRPC ⚡

## Architecture Overview

```
┌──────────────┐      HTTP/REST      ┌─────────────┐      gRPC       ┌──────────────┐
│   Frontend   │ ─────────────────> │   Gateway   │ ──────────────> │  gRPC Server │
│ (Browser UI) │ <───────────────── │ (Port 3000) │ <────────────── │  (Port 8080) │
└──────────────┘      JSON/SSE       └─────────────┘    Protobuf     └──────────────┘
                      WebSocket
```

## Components

### 1. **gRPC Server** (Port 8080)
- Pure gRPC implementation
- Handles all 4 communication patterns
- Uses Protocol Buffers

### 2. **HTTP Gateway** (Port 3000)
- Translates HTTP/REST → gRPC
- Serves static frontend files
- Handles:
  - REST API for Unary RPC
  - Server-Sent Events (SSE) for Server Streaming
  - REST API for Client Streaming
  - WebSocket for Bidirectional Streaming

### 3. **Frontend** (Browser)
- Beautiful, modern UI
- Real-time interactions
- Visual demonstrations of all 4 patterns

## Running the Complete System

### Step 1: Start gRPC Server
```bash
cd server
go run main.go
```
You should see: `gRPC server listening on :8080`

### Step 2: Start HTTP Gateway
```bash
cd gateway
go run main.go
```
You should see:
```
🚀 HTTP Gateway running on http://localhost:3000
📡 Connected to gRPC server on localhost:8080
🌐 Open http://localhost:3000 in your browser
```

### Step 3: Open Browser
```
http://localhost:3000
```

## API Endpoints

### 1. Unary RPC
**Endpoint:** `POST /api/unary`

**Request:**
```json
{
  "name": "Alice"
}
```

**Response:**
```json
{
  "message": "Hello Alice"
}
```

**Frontend Code:**
```javascript
const response = await fetch('http://localhost:3000/api/unary', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Alice' })
});
const data = await response.json();
console.log(data.message); // "Hello Alice"
```

---

### 2. Server Streaming
**Endpoint:** `GET /api/server-stream?name=Bob`

**Technology:** Server-Sent Events (SSE)

**Frontend Code:**
```javascript
const eventSource = new EventSource('http://localhost:3000/api/server-stream?name=Bob');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log(data.message); // "Hello Bob - Message 1 of 5"
};

eventSource.addEventListener('done', function(event) {
    console.log('Stream completed');
    eventSource.close();
});
```

**Response Stream:**
```
data: {"message":"Hello Bob - Message 1 of 5"}

data: {"message":"Hello Bob - Message 2 of 5"}

data: {"message":"Hello Bob - Message 3 of 5"}

...

event: done
data: {"message":"Stream complete"}
```

---

### 3. Client Streaming
**Endpoint:** `POST /api/client-stream`

**Request:**
```json
["Alice", "Bob", "Charlie"]
```

**Response:**
```json
{
  "message": "Hello to all: Alice, Bob, Charlie! (Total: 3 people)"
}
```

**Frontend Code:**
```javascript
const names = ['Alice', 'Bob', 'Charlie'];

const response = await fetch('http://localhost:3000/api/client-stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(names)
});
const data = await response.json();
console.log(data.message);
```

---

### 4. Bidirectional Streaming
**Endpoint:** `ws://localhost:3000/api/bidirectional`

**Technology:** WebSocket

**Frontend Code:**
```javascript
const ws = new WebSocket('ws://localhost:3000/api/bidirectional');

ws.onopen = function() {
    console.log('Connected');
    // Send message
    ws.send(JSON.stringify({ name: 'Hello' }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data.message);
};

// Send more messages
ws.send(JSON.stringify({ name: 'How are you?' }));
```

**Message Flow:**
```
Client → Server: {"name": "Hello"}
Server → Client: {"message": "Echo: Hello Hello!"}
Client → Server: {"name": "How are you?"}
Server → Client: {"message": "Echo: Hello How are you?!"}
```

## Technology Stack

### Backend
- **Go** - Programming language
- **gRPC** - Remote procedure calls
- **Protocol Buffers** - Data serialization
- **net/http** - HTTP server
- **gorilla/websocket** - WebSocket support

### Frontend
- **HTML5** - Structure
- **CSS3** - Modern styling with gradients and animations
- **Vanilla JavaScript** - No frameworks, pure JS
- **Fetch API** - HTTP requests
- **EventSource** - Server-Sent Events
- **WebSocket API** - Bidirectional communication

## Protocol Translations

### Unary RPC
```
HTTP POST → gRPC Unary → HTTP Response
JSON → Protobuf → JSON
```

### Server Streaming
```
HTTP GET → gRPC Server Stream → SSE Stream
Query Params → Protobuf Stream → JSON Events
```

### Client Streaming
```
HTTP POST (Array) → gRPC Client Stream → HTTP Response
JSON Array → Multiple Protobuf Messages → JSON
```

### Bidirectional Streaming
```
WebSocket ↔ gRPC Bidirectional Stream
JSON Messages ↔ Protobuf Messages
```

## Performance Benefits

### Why This Architecture?

1. **Browser Compatibility** - Browsers can't directly call gRPC
2. **Best of Both Worlds** - Easy frontend + Fast backend
3. **Real-time Updates** - SSE and WebSocket for streaming
4. **Type Safety** - Protocol Buffers on backend
5. **Scalability** - gRPC's performance with web accessibility

### Performance Comparison

| Metric | Direct gRPC | REST/JSON | Our Setup |
|--------|-------------|-----------|-----------|
| Backend Speed | 10/10 | 5/10 | 10/10 |
| Browser Support | 0/10 | 10/10 | 10/10 |
| Streaming | 10/10 | 3/10 | 9/10 |
| Type Safety | 10/10 | 3/10 | 10/10 |
| **Overall** | 7.5/10 | 5.25/10 | **9.75/10** |

## Testing the Frontend

### Test Pattern 1: Unary RPC
1. Open http://localhost:3000
2. In "Unary RPC" card, enter your name
3. Click "Send Request"
4. See immediate response

### Test Pattern 2: Server Streaming
1. In "Server Streaming" card, enter your name
2. Click "Start Streaming"
3. Watch 5 messages arrive over 5 seconds
4. Click "Stop" to cancel early

### Test Pattern 3: Client Streaming
1. In "Client Streaming" card
2. Enter names and click "Add Name" (repeat multiple times)
3. Click "Send All"
4. See aggregated response

### Test Pattern 4: Bidirectional
1. In "Bidirectional Streaming" card
2. Click "Connect"
3. Type messages and click "Send"
4. See real-time echo responses
5. Click "Disconnect" when done

## Troubleshooting

### Port Already in Use
```bash
# Kill process on port 8080 (gRPC server)
lsof -ti:8080 | xargs kill -9

# Kill process on port 3000 (Gateway)
lsof -ti:3000 | xargs kill -9
```

### Gateway Can't Connect to gRPC
- Make sure gRPC server is running on port 8080
- Check server logs for errors

### Frontend Can't Connect
- Make sure gateway is running on port 3000
- Check browser console for errors
- Clear browser cache

### WebSocket Connection Failed
- Make sure gateway is running
- Check for firewall blocking WebSocket connections
- Try using `ws://localhost:3000` instead of `wss://`

## Advanced Features

### Adding HTTPS
1. Generate SSL certificates
2. Update gateway to use TLS
3. Change frontend URLs to https://

### Adding Authentication
1. Add JWT token generation
2. Validate tokens in gateway
3. Pass credentials to gRPC with metadata

### Production Deployment
1. Build static frontend
2. Deploy gateway with reverse proxy (nginx)
3. Use proper domain names
4. Enable CORS properly
5. Add rate limiting

## Project Structure

```
grpc-example/
├── server/           # gRPC server (port 8080)
│   └── main.go
├── gateway/          # HTTP Gateway (port 3000)
│   └── main.go
├── public/           # Frontend files
│   ├── index.html
│   ├── styles.css
│   └── script.js
├── proto/            # Protocol Buffers
│   ├── helloworld.proto
│   ├── helloworld.pb.go
│   └── helloworld_grpc.pb.go
├── client/           # CLI client (optional)
│   └── main.go
├── go.mod
└── go.sum
```

## Next Steps

1. ✅ All 4 patterns working with frontend
2. Add more complex data types
3. Implement authentication
4. Add database integration
5. Deploy to production
6. Monitor performance
7. Add error handling UI
8. Implement retry logic

## Resources

- [gRPC Official Docs](https://grpc.io/)
- [Server-Sent Events MDN](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [WebSocket API MDN](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)

---

**🎉 You now have a complete full-stack gRPC application with frontend integration!**

