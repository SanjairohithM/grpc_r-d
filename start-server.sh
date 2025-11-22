#!/bin/bash

cd "$(dirname "$0")/server"
echo "🚀 Starting gRPC Server..."
go run main.go database.go

