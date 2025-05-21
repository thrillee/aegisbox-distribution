# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=manager-api ./cmd/manager-api

# Stage 2: Final stage
FROM gcr.io/distroless/base-debian11:latest

# Set working directory
WORKDIR /app

# Copy only the binary from the builder stage
COPY --from=builder --chown=nonroot:nonroot /app/manager-api /app/manager-api

# Expose the required port
EXPOSE 8001

# Use nonroot user for security
USER nonroot

# Command to run the application
ENTRYPOINT ["/app/manager-api"]
