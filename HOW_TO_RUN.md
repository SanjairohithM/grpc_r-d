# 🚀 How to Run Your Services

## ⚠️ **Important: Correct Commands**

When running the gateway, you must include **all files** because functions are split across multiple files.

---

## ✅ **Correct Way to Run**

### **Option 1: Use `go run .` (Recommended)**

```bash
# Start gRPC Server
cd server
go run main.go database.go

# Start HTTP Gateway
cd gateway
go run .  # ← This includes all .go files in the directory
```

### **Option 2: Use Start Scripts**

```bash
# Terminal 1: Start Server
./start-server.sh

# Terminal 2: Start Gateway
./start-gateway.sh
```

### **Option 3: Specify All Files**

```bash
# Start Gateway (explicit)
cd gateway
go run main.go middleware.go connection.go
```

---

## ❌ **Wrong Way (Will Cause Errors)**

```bash
# ❌ DON'T DO THIS - Missing files!
cd gateway
go run main.go  # ← This will fail with "undefined" errors
```

**Error you'll see:**
```
undefined: initGRPCConnection
undefined: rateLimitMiddleware
undefined: enableGzip
undefined: requestLogger
```

---

## 📁 **File Structure**

```
gateway/
├── main.go          ← Main entry point
├── middleware.go    ← Gzip, Rate Limit, Logging
└── connection.go    ← gRPC connection pooling

server/
├── main.go          ← Main entry point
└── database.go      ← Database models & helpers
```

---

## 🎯 **Quick Start**

```bash
# Terminal 1
cd server && go run main.go database.go

# Terminal 2  
cd gateway && go run .
```

That's it! ✅


