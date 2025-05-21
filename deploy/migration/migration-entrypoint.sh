#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

# Check if required environment variables are set
if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
  echo "Error: One or more database environment variables (DB_USER, DB_PASSWORD, DB_NAME, DB_HOST, DB_PORT) are not set."
  exit 1
fi

# Add retry logic for database connectivity
MAX_RETRIES=10
RETRY_INTERVAL=5
retries=0

echo "Waiting for RDS database to be ready..."
while [ $retries -lt $MAX_RETRIES ]; do
  if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
    echo "Database is ready!"
    break
  fi
  
  retries=$((retries + 1))
  echo "Database not ready yet. Retry $retries/$MAX_RETRIES. Waiting $RETRY_INTERVAL seconds..."
  sleep $RETRY_INTERVAL
done

if [ $retries -eq $MAX_RETRIES ]; then
  echo "Error: Could not connect to the database after $MAX_RETRIES attempts."
  exit 1
fi

# Construct the connection string for goose
CONN_STR="user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=disable"

echo "Running database migrations..."

# Execute goose command
# Pass the command (e.g., "up") as arguments to this script ($@)
goose -dir /app/migrations postgres "$CONN_STR" $@

echo "Database migrations finished."

# Exit successfully
exit 0
