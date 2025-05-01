    # Stage 1: Builder stage using golang image
    FROM golang:1.21-alpine AS builder

    # Set working directory
    WORKDIR /src

    # Install goose
    # Ensure you have the correct import path if you use goose as a library
    # Or use the version matching your project if installed via go mod
    RUN go install github.com/pressly/goose/v3/cmd/goose@latest

    # Stage 2: Final stage using alpine
    FROM alpine:latest

    # Set working directory
    WORKDIR /app

    # Copy the goose binary from the builder stage
    COPY --from=builder /go/bin/goose /usr/local/bin/goose

    # Copy your migration files into the container
    # Adjust the source path './db/migration' if your migrations are located elsewhere
    COPY ./db/migration /app/migrations

    # Copy the entrypoint script (we will create this next)
    COPY migration-entrypoint.sh /app/migration-entrypoint.sh

    # Make the entrypoint script executable
    RUN chmod +x /app/migration-entrypoint.sh

    # Set the entrypoint script to run when the container starts
    ENTRYPOINT ["/app/migration-entrypoint.sh"]

    # Default command (can be overridden in docker-compose)
    CMD ["up"]
    
