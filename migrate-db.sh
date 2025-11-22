#!/bin/bash

echo "🔄 Running Database Migrations"
echo "==============================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "❌ .env file not found. Run ./setup-db.sh first"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL not set in .env file"
    exit 1
fi

echo "📊 Connecting to database..."
echo ""

# Run migrations
cd server
go run migrate.go database.go

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 Database migration complete!"
else
    echo ""
    echo "❌ Migration failed. Please check the error above."
    exit 1
fi

