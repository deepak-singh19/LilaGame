#!/bin/bash
set -e

# Configuration
PROJECT_ID="lila-games-ttt"
REGION="us-central1"
SERVICE_NAME="ttt-backend"
DB_INSTANCE_NAME="ttt-postgres"

echo "🚀 Deploying Tic-Tac-Toe Backend to Google Cloud"

# Set project
gcloud config set project $PROJECT_ID

# Get the Cloud SQL connection name
CONNECTION_NAME=$(gcloud sql instances describe $DB_INSTANCE_NAME --format="value(connectionName)")
echo "📊 Using existing database: $CONNECTION_NAME"

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
    --add-cloudsql-instances=$CONNECTION_NAME \
    --set-env-vars "NAKAMA_DATABASE_URL=postgres://postgres:SecurePassword123!@/nakama?host=/cloudsql/$CONNECTION_NAME&sslmode=disable"

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)")

echo "✅ Backend deployment complete!"
echo "🌐 Service URL: $SERVICE_URL"
echo "📊 Console: $SERVICE_URL/console"
echo "🔗 Database: $CONNECTION_NAME"
echo ""
echo "Next step: Deploy frontend to Vercel with backend URL: $SERVICE_URL"
