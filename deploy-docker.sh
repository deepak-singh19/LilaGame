#!/bin/bash
set -e

echo "🚀 Deploying Tic-Tac-Toe with Docker"

# Configuration
PROJECT_NAME="lila-games"
BACKEND_IMAGE="lila-backend"
FRONTEND_IMAGE="lila-frontend"

echo "📦 Building backend Docker image..."
cd backend
docker build -t $BACKEND_IMAGE .
cd ..

echo "📦 Building frontend Docker image..."
cd frontend/tac-tac-toe
docker build -t $FRONTEND_IMAGE .
cd ../..

echo "🐳 Starting services with Docker Compose..."
docker-compose -f docker-compose.prod.yml up -d

echo "⏳ Waiting for services to start..."
sleep 10

echo "✅ Deployment complete!"
echo ""
echo "🌐 Services running:"
echo "  - Frontend: http://localhost:3000"
echo "  - Backend: http://localhost:7350"
echo "  - Nakama Console: http://localhost:7351/console"
echo ""
echo "📊 Check logs:"
echo "  docker-compose -f docker-compose.prod.yml logs -f"
echo ""
echo "🛑 Stop services:"
echo "  docker-compose -f docker-compose.prod.yml down"
