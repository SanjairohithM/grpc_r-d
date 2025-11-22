#!/bin/bash

echo "🚀 Starting Full-Stack gRPC Application with Next.js"
echo "====================================================="
echo ""

# Kill existing processes
echo "🧹 Cleaning up existing processes..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
lsof -ti:8081 | xargs kill -9 2>/dev/null
lsof -ti:3000 | xargs kill -9 2>/dev/null

echo ""
echo "📦 Installing dependencies..."
go mod tidy

echo ""
echo "🔧 Starting gRPC Server (Port 8080)..."
cd server
go run main.go &
GRPC_PID=$!
cd ..

sleep 2

echo "🌐 Starting HTTP Gateway (API Port 8081)..."
cd gateway
go run main.go &
GATEWAY_PID=$!
cd ..

sleep 2

echo "⚛️  Starting Next.js Frontend (Port 3000)..."
cd frontend

# Create .env.local if it doesn't exist
if [ ! -f .env.local ]; then
    echo "NEXT_PUBLIC_API_URL=http://localhost:8081/api" > .env.local
    echo "✓ Created .env.local"
fi

npm run dev &
NEXTJS_PID=$!
cd ..

echo ""
echo "✅ All services started!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Application is ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📡 gRPC Server:     http://localhost:8080  (HTTP/2)"
echo "🌐 HTTP Gateway:    http://localhost:8081  (API)"
echo "⚛️  Next.js Frontend: http://localhost:3000  (UI)"
echo ""
echo "Open your browser and navigate to:"
echo "👉 http://localhost:3000"
echo ""
echo "Architecture:"
echo "  Browser → Next.js (3000) → Gateway (8081) → gRPC (8080)"
echo "           HTTP/1.1          HTTP/1.1         HTTP/2"
echo ""
echo "Press Ctrl+C to stop all services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Trap Ctrl+C to kill all processes
trap "echo ''; echo '🛑 Stopping all services...'; kill $GRPC_PID $GATEWAY_PID $NEXTJS_PID 2>/dev/null; exit" INT

# Wait for processes
wait

