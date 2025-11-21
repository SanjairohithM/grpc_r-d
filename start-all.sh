#!/bin/bash

echo "🚀 Starting gRPC Full-Stack Application"
echo "========================================"
echo ""

# Check if ports are available
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 8080 is already in use. Killing existing process..."
    lsof -ti:8080 | xargs kill -9
fi

if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 3000 is already in use. Killing existing process..."
    lsof -ti:3000 | xargs kill -9
fi

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

echo "🌐 Starting HTTP Gateway (Port 3000)..."
cd gateway
go run main.go &
GATEWAY_PID=$!
cd ..

echo ""
echo "✅ All services started!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Application is ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📡 gRPC Server:    http://localhost:8080"
echo "🌐 HTTP Gateway:   http://localhost:3000"
echo "💻 Frontend UI:    http://localhost:3000"
echo ""
echo "Open your browser and navigate to:"
echo "👉 http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Trap Ctrl+C to kill both processes
trap "echo ''; echo '🛑 Stopping all services...'; kill $GRPC_PID $GATEWAY_PID 2>/dev/null; exit" INT

# Wait for processes
wait

