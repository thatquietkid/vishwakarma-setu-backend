#!/bin/bash
set -e

# Parse DSN or use defaults matching docker-compose
DB_HOST="db"
DB_USER="vishwakarma_user"
DB_PASSWORD="password"
DB_NAME="vishwakarma_db"

export PGPASSWORD=$DB_PASSWORD

echo "⏳ Waiting for Postgres at $DB_HOST:5432..."

# Loop until Postgres is ready to accept connections
until psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "Postgres is unavailable - sleeping..."
  sleep 2
done

echo "✅ Postgres is up and running!"

# (Optional) Run migrations or other setup here if needed
# The app will likely handle AutoMigrate, so we just exit successfully.