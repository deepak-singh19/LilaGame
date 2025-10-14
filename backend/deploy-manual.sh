#!/bin/bash
set -e

echo "🚀 Manual deployment to Cloud Run"

# Configuration
PROJECT_ID="learned-age-471809-j4"
REGION="us-central1"
SERVICE_NAME="ttt-backend"

# Set project
gcloud config set project $PROJECT_ID

# Build and push image
echo "📦 Building and pushing image..."
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME .

# Deploy to Cloud Run
echo "🚀 Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
    --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
    --region $REGION \
    --platform managed \
    --allow-unauthenticated \
    --port 7350 \
    --memory 1Gi \
    --cpu 1 \
    --max-instances 5 \
    --min-instances 0 \
    --set-env-vars "NAKAMA_DATABASE_URL=postgres://postgres:Madleo@123@34.31.236.20:5432/nakama?sslmode=disable" \
    --set-env-vars "NAKAMA_DATABASE_ADDRESS=postgres://postgres:Madleo@123@34.31.236.20:5432/nakama?sslmode=disable"

echo "✅ Deployment complete!"
