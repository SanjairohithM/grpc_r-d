#!/bin/bash

echo "🔧 Setting up Database Configuration"
echo "====================================="
echo ""

# Check if .env exists
if [ -f .env ]; then
    echo "⚠️  .env file already exists"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping .env creation"
        exit 0
    fi
fi

echo "📝 Creating .env file..."
echo ""
echo "Please enter your Supabase database password:"
read -s PASSWORD

cat > .env << EOF
# Supabase Database Connection
# Connection pooling (for queries)
DATABASE_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:${PASSWORD}@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true"

# Direct connection (for migrations)
DIRECT_URL="postgresql://postgres.bvtsauqbkrsnyfrrayuh:${PASSWORD}@aws-1-ap-northeast-2.pooler.supabase.com:5432/postgres"
EOF

echo ""
echo "✅ .env file created successfully!"
echo ""
echo "📊 Testing database connection..."

# Test connection using Go
cd server
go run -tags test -exec "true" << 'GOTEST'
package main

import (
    "fmt"
    "log"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL not set")
    }
    
    db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    
    sqlDB, _ := db.DB()
    if err := sqlDB.Ping(); err != nil {
        log.Fatalf("Ping failed: %v", err)
    }
    
    fmt.Println("✅ Database connection successful!")
}
GOTEST

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 Database setup complete!"
    echo ""
    echo "Next steps:"
    echo "1. Run migrations: ./migrate-db.sh"
    echo "2. Start server: cd server && go run main.go"
else
    echo ""
    echo "❌ Database connection failed. Please check your credentials."
fi

