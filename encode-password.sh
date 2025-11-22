#!/bin/bash

echo "🔐 Password URL Encoder for Supabase"
echo "====================================="
echo ""
echo "Enter your Supabase database password:"
read -s PASSWORD

# Try Python first
if command -v python3 &> /dev/null; then
    ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$PASSWORD'))")
elif command -v python &> /dev/null; then
    ENCODED=$(python -c "import urllib.parse; print(urllib.parse.quote('$PASSWORD'))")
elif command -v node &> /dev/null; then
    ENCODED=$(node -e "console.log(encodeURIComponent('$PASSWORD'))")
else
    echo "❌ Error: Need Python or Node.js to encode password"
    echo ""
    echo "Please use online tool: https://www.urlencoder.org/"
    exit 1
fi

echo ""
echo "✅ URL-encoded password:"
echo "$ENCODED"
echo ""
echo "📝 Add this to your .env file:"
echo ""
echo "DATABASE_URL=\"postgresql://postgres.bvtsauqbkrsnyfrrayuh:${ENCODED}@aws-1-ap-northeast-2.pooler.supabase.com:6543/postgres?pgbouncer=true\""
echo "DIRECT_URL=\"postgresql://postgres.bvtsauqbkrsnyfrrayuh:${ENCODED}@aws-1-ap-northeast-2.pooler.supabase.com:5432/postgres\""
echo ""

