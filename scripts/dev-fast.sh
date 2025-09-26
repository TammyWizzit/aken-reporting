#!/bin/bash
# Ultra-fast development startup script

set -e

echo "🔨 Building Go application locally (super fast)..."
CGO_ENABLED=0 GOOS=linux go build -o main .

echo "🐳 Building Docker image (takes ~10 seconds)..."
docker-compose -f docker-compose.fast.yml build dev-fast

echo "🚀 Starting development environment..."
docker-compose -f docker-compose.fast.yml up dev-fast -d

echo "⏳ Waiting for services to start..."
sleep 5

echo "🏥 Testing health endpoint..."
curl -s http://localhost:8090/api/v2/health | jq . || echo "Service still starting..."

echo ""
echo "✅ Development environment ready!"
echo "🌐 API available at: http://localhost:8090"
echo "🏥 Health check: http://localhost:8090/api/v2/health"
echo "📈 Transactions: http://localhost:8090/api/v2/transactions"
echo ""
echo "📊 To view logs: docker-compose -f docker-compose.fast.yml logs -f dev-fast"
echo "🛑 To stop: docker-compose -f docker-compose.fast.yml down"