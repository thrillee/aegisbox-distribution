# One image for the whole backend: database migrations, manager API and SMPP gateway.
# Default command "all" runs migrations, then the API and gateway in one process.
# Other commands: migrate, api, gateway.

# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build for the platform Docker is building for (amd64 or arm64)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags='-s -w' -o=aegisbox ./cmd/aegisbox

# Stage 2: Final stage
FROM gcr.io/distroless/base-debian12:nonroot

# Set working directory
WORKDIR /app

# Copy only the binary; migrations are embedded in it
COPY --from=builder --chown=nonroot:nonroot /app/aegisbox /app/aegisbox

# Manager API, SMPP server, SP HTTP API, MNO DLR receiver
EXPOSE 8001 6886 8000 9001

USER nonroot

ENTRYPOINT ["/app/aegisbox"]
CMD ["all"]
