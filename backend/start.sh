#!/bin/bash

# Start PostgreSQL
pg_ctl -D /home/runner/postgres_data -l /home/runner/postgres.log start

# Wait for PostgreSQL to start
sleep 5

# Create database
createdb nakama

# Set environment variables
export DATABASE_URL="postgresql://postgres@localhost:5432/nakama?sslmode=disable"

# Build the Go plugin
go build --trimpath --buildmode=plugin -o ./backend.so

# Run Nakama migrations
/nakama/nakama migrate up --database.address "$DATABASE_URL"

# Start Nakama
exec /nakama/nakama --database.address "$DATABASE_URL" --database.driver postgres
