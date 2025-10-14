#!/bin/bash

# Hostinger VPS Deployment Script for Tic-Tac-Toe Backend
# Make sure to replace these variables with your actual Hostinger VPS details

VPS_HOST="your-vps-ip-address"
VPS_USER="root"
VPS_PASSWORD="your-vps-password"
APP_NAME="ttt-backend"

echo "🚀 Deploying Tic-Tac-Toe Backend to Hostinger VPS..."

# Upload the backend files
echo "📤 Uploading backend files..."
scp -r backend/ $VPS_USER@$VPS_HOST:/opt/$APP_NAME/

# Connect to VPS and set up the application
echo "🔧 Setting up application on VPS..."
ssh $VPS_USER@$VPS_HOST << 'EOF'
    cd /opt/ttt-backend
    
    # Update system
    apt update && apt upgrade -y
    
    # Install Docker
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    
    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    # Create docker-compose.yml for production
    cat > docker-compose.yml << 'DOCKERFILE'
version: '3.8'
services:
  postgres:
    image: postgres:13-alpine
    container_name: ttt_postgres
    environment:
      - POSTGRES_DB=nakama
      - POSTGRES_USER=nakama
      - POSTGRES_PASSWORD=your-secure-password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  nakama:
    build: .
    container_name: ttt_nakama
    depends_on:
      - postgres
    entrypoint:
      - "/bin/sh"
      - "-ecx"
      - >
        /nakama/nakama migrate up --database.address nakama:your-secure-password@postgres:5432/nakama?sslmode=disable &&
        exec /nakama/nakama --config /nakama/data/config.yml --database.address nakama:your-secure-password@postgres:5432/nakama?sslmode=disable
    ports:
      - "7349:7349"
      - "7350:7350"
      - "7351:7351"
    restart: unless-stopped

volumes:
  postgres_data:
DOCKERFILE

    # Start the application
    docker-compose up -d --build
    
    echo "✅ Backend deployed successfully!"
    echo "🌐 Backend URL: http://$VPS_HOST:7350"
    echo "📊 Console URL: http://$VPS_HOST:7351"
EOF

echo "🎉 Deployment complete!"
echo "🌐 Your backend is now available at: http://$VPS_HOST:7350"
echo "📊 Console available at: http://$VPS_HOST:7351"
