#!/bin/bash

# Install dependencies
npm install

# Set environment variables for production
export VITE_NAKAMA_HOST="localhost"
export VITE_NAKAMA_PORT="7350"
export VITE_NAKAMA_SSL="false"
export VITE_NAKAMA_SERVER_KEY="defaultkey"

# Build the frontend
npm run build

# Start the development server
npm run dev -- --host 0.0.0.0 --port 3000
