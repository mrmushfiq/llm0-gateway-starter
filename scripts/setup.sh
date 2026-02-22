#!/bin/bash

# Setup script for LLM Gateway Starter

set -e

echo "🚀 Setting up LLM Gateway Starter"
echo "=================================="
echo ""

# Check prerequisites
echo "1️⃣  Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.21+ from https://golang.org/dl/"
    exit 1
fi
echo "   ✓ Go $(go version | awk '{print $3}')"

if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker not found. You'll need Docker to run PostgreSQL and Redis."
    echo "   Install from https://docker.com"
else
    echo "   ✓ Docker $(docker --version | awk '{print $3}' | tr -d ',')"
fi

if ! command -v psql &> /dev/null; then
    echo "⚠️  psql not found. You'll need it to run migrations."
    echo "   Install: brew install postgresql (macOS) or apt install postgresql-client (Linux)"
else
    echo "   ✓ psql"
fi

echo ""

# Install Go dependencies
echo "2️⃣  Installing Go dependencies..."
go mod download
echo "   ✓ Dependencies installed"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "3️⃣  Creating .env file..."
    cp .env.example .env
    echo "   ✓ Created .env file"
    echo "   ⚠️  Please edit .env and add your API keys!"
    echo ""
else
    echo "3️⃣  .env file already exists"
    echo ""
fi

# Start Docker services
echo "4️⃣  Starting PostgreSQL and Redis (Docker)..."
if command -v docker &> /dev/null; then
    docker-compose up -d postgres redis
    echo "   ✓ PostgreSQL and Redis started"
    
    # Wait for PostgreSQL to be ready
    echo "   Waiting for PostgreSQL to be ready..."
    sleep 5
    
    # Run migrations
    echo ""
    echo "5️⃣  Running database migrations..."
    docker exec -i gateway_postgres psql -U gateway -d gateway < migrations/001_initial_schema.sql
    echo "   ✓ Migrations applied"
else
    echo "   ⚠️  Docker not available. Please start PostgreSQL and Redis manually."
    echo "   Then run: psql \$DATABASE_URL -f migrations/001_initial_schema.sql"
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env and add your API keys (OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY)"
echo "  2. Run: go run cmd/gateway/main.go"
echo "  3. Test: ./scripts/test_basic.sh"
echo ""
echo "Test API key: gw_test_abc123"
