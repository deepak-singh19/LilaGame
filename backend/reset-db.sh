#!/bin/bash

echo "Resetting Nakama database..."

# Stop the containers
docker compose down

# Remove the database volume
docker volume rm lila_games_postgres_data 2>/dev/null || true

# Start the containers again
docker compose up -d

echo "Database reset complete!"
echo "You can now use any username without conflicts."
