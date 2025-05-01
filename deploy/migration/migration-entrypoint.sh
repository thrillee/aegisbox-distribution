    #!/bin/sh

    # Exit immediately if a command exits with a non-zero status.
    set -e

    # Check if required environment variables are set
    if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
      echo "Error: One or more database environment variables (DB_USER, DB_PASSWORD, DB_NAME, DB_HOST, DB_PORT) are not set."
      exit 1
    fi

    # Construct the connection string for goose
    # Ensure sslmode is set appropriately for your environment (e.g., disable, require)
    CONN_STR="user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT} sslmode=disable"

    echo "Running database migrations..."

    # Execute goose command
    # Pass the command (e.g., "up") as arguments to this script ($@)
    # The default is "up" from the Dockerfile CMD, but can be overridden.
    goose -dir /app/migrations postgres "$CONN_STR" $@

    echo "Database migrations finished."

    # Exit successfully
    exit 0
