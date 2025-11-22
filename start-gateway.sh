#!/bin/bash

cd "$(dirname "$0")/gateway"
echo "🚀 Starting HTTP Gateway..."
go run .

