#!/bin/bash
set -e

# Configuration
PROJECT_ID="lila-games-ttt"
REGION="us-central1"
SERVICE_NAME="ttt-backend"

echo "🚀 Deploying Tic-Tac-Toe Backend to Google Cloud Run (Simple)"

# Set project
gcloud config set project $PROJECT_ID

# Enable required APIs
echo "📋 Enabling required APIs..."
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com

# Build and push Docker image
echo "🏗️ Building and pushing Docker image..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME .

# Deploy to Cloud Run
echo "🚀 Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
    --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
    --region $REGION \
    --platform managed \
    --allow-unauthenticated \
    --port 7350 \
    --memory 2Gi \
    --cpu 2 \
    --max-instances 10 \
    --set-env-vars "NAKAMA_DATABASE_URL=postgres://postgres:password@/nakama?host=/cloudsql/$PROJECT_ID:us-central1:ttt-postgres&sslmode=disable"

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)")

echo "✅ Deployment complete!"
echo "🌐 Service URL: $SERVICE_URL"
echo "📊 Console: $SERVICE_URL/console"
echo ""
echo "Note: This uses a temporary database. For production, set up Cloud SQL."
