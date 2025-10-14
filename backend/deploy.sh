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

# Enable required APIs
echo "📋 Enabling required APIs..."
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable sql-component.googleapis.com

# Create Cloud SQL instance for PostgreSQL
echo "🗄️ Creating Cloud SQL instance..."
gcloud sql instances create $DB_INSTANCE_NAME \
    --database-version=POSTGRES_13 \
    --tier=db-f1-micro \
    --region=$REGION \
    --root-password=your-secure-password \
    --storage-type=SSD \
    --storage-size=10GB \
    --backup-start-time=03:00

# Create database
echo "📊 Creating database..."
gcloud sql databases create nakama --instance=$DB_INSTANCE_NAME

# Create Cloud Run service
echo "🏗️ Building and deploying to Cloud Run..."
gcloud builds submit --config cloudbuild.yaml .

# Get the Cloud SQL connection name
CONNECTION_NAME=$(gcloud sql instances describe $DB_INSTANCE_NAME --format="value(connectionName)")

# Update Cloud Run service with Cloud SQL connection
echo "🔗 Updating service with database connection..."
gcloud run services update $SERVICE_NAME \
    --region=$REGION \
    --add-cloudsql-instances=$CONNECTION_NAME \
    --set-env-vars="NAKAMA_DATABASE_URL=postgres://postgres:your-secure-password@/nakama?host=/cloudsql/$CONNECTION_NAME&sslmode=disable"

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)")

echo "✅ Deployment complete!"
echo "🌐 Service URL: $SERVICE_URL"
echo "📊 Database: $CONNECTION_NAME"
echo ""
echo "Next steps:"
echo "1. Update your frontend to use: $SERVICE_URL"
echo "2. Test the WebSocket connection at: $SERVICE_URL"
echo "3. Access the console at: $SERVICE_URL/console"
