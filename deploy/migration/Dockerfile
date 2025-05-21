    # Stage 1: Builder stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install goose and dependencies
RUN apk add --no-cache git postgresql-client
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Stage 2: Final stage
FROM alpine:latest

# Install PostgreSQL client for health checks
RUN apk add --no-cache postgresql-client

# Set working directory
WORKDIR /app

# Copy the goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy your migration files
COPY ./db/migration /app/migrations

# Copy the entrypoint script
COPY ./deploy/migration/migration-entrypoint.sh /app/migration-entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /app/migration-entrypoint.sh

# Set the entrypoint script to run when the container starts
ENTRYPOINT ["/app/migration-entrypoint.sh"]

# Default command (can be overridden in docker-compose)
CMD ["up"]
