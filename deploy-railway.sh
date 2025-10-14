#!/bin/bash

echo "🚀 Deploying Tic-Tac-Toe Backend to Railway..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing..."
    
    # Install Railway CLI
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install railway
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -fsSL https://railway.app/install.sh | sh
    else
        echo "❌ Unsupported OS. Please install Railway CLI manually from https://docs.railway.app/develop/cli"
        exit 1
    fi
fi

# Login to Railway
echo "🔐 Logging in to Railway..."
railway login

# Create new project
echo "📦 Creating Railway project..."
railway init ttt-backend

# Add PostgreSQL database
echo "🗄️ Adding PostgreSQL database..."
railway add postgresql

# Deploy the backend
echo "🚀 Deploying backend..."
railway up

# Get the deployment URL
echo "🌐 Getting deployment URL..."
RAILWAY_URL=$(railway domain)

echo "✅ Deployment complete!"
echo "🌐 Backend URL: https://$RAILWAY_URL"
echo "📊 Console URL: https://$RAILWAY_URL:7351"

# Update Vercel environment variables
echo "🔄 Updating Vercel environment variables..."
echo "Please update your Vercel environment variables:"
echo "VITE_NAKAMA_HOST = $RAILWAY_URL"
echo "VITE_NAKAMA_PORT = 443"
echo "VITE_NAKAMA_SSL = true"
echo "VITE_NAKAMA_SERVER_KEY = defaultkey"
