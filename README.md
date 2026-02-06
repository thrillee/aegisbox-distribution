# Aegisbox

Aegisbox is an SMPP gateway and manager API service written in Go. It consists of two binaries:
- **smpp-gateway**: Handles SMPP connections with service providers and routes SMS to MNOs
- **manager-api**: REST API for managing the gateway (Gin-based)

## Prerequisites

- Go 1.23.0+
- Docker & Docker Compose
- PostgreSQL 16 (if running without Docker)

## Local Development

### 1. Clone and Setup

```bash
git clone https://github.com/thrillee/aegisbox.git
cd aegisbox
```

### 2. Configure Environment

Copy the example environment file and update values:

```bash
cp deploy/.env.example deploy/.env
```

Required environment variables:
- `DATABASE_URL` - PostgreSQL connection string
- `API_ADDR` - Manager API listen address (default: :8001)
- Other service-specific variables

### 3. Run with Docker Compose

```bash
# Start all services
cd deploy
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

Services will be available at:
- **Manager API**: http://localhost:8001
- **SMPP Gateway**: localhost:2775
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

### 4. Run Locally (without Docker)

```bash
# Run tests
make test

# Build binaries
make build

# Run specific service
./tmp/bin/smpp-gateway
./tmp/bin/manager-api

# Live reload with air
make run/live/gateway
make run/live/manager-api
```

## Production Deployment

### Pulling Images from GHCR

On your remote server, authenticate with GitHub Container Registry and pull the images:

```bash
# Login to GHCR (use GitHub PAT with read:packages scope)
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Pull images
docker pull ghcr.io/thrillee/aegisbox/migration:latest
docker pull ghcr.io/thrillee/aegisbox/manager-api:latest
docker pull ghcr.io/thrillee/aegisbox/smpp-gateway:latest

# Or pull all at once
docker pull ghcr.io/thrillee/aegisbox/migration:latest \
  ghcr.io/thrillee/aegisbox/manager-api:latest \
  ghcr.io/thrillee/aegisbox/smpp-gateway:latest
```

### Deploy with Docker Compose

```bash
# On the remote server
git clone https://github.com/thrillee/aegisbox.git
cd aegisbox
git checkout docker-prod

# Copy and configure environment
cp deploy/.env.example deploy/.env
vim deploy/.env  # Update DATABASE_URL and other values

# Update images to use GHCR versions
sed -i 's|image: migration:latest|image: ghcr.io/thrillee/aegisbox/migration:latest|g' deploy/docker-compose.yml
sed -i 's|image: manager-api:latest|image: ghcr.io/thrillee/aegisbox/manager-api:latest|g' deploy/docker-compose.yml
sed -i 's|image: smpp-gateway:latest|image: ghcr.io/thrillee/aegisbox/smpp-gateway:latest|g' deploy/docker-compose.yml

# Start services
cd deploy
docker compose pull
docker compose up -d

# View logs
docker compose logs -f
```

### Quick Pull & Deploy Script

```bash
#!/bin/bash
REPO_DIR="/opt/aegisbox"
cd $REPO_DIR

# Backup current environment
cp deploy/.env deploy/.env.backup 2>/dev/null || true

# Pull latest changes
git fetch origin docker-prod
git checkout docker-prod
git pull origin docker-prod

# Update Docker images
docker compose -f deploy/docker-compose.yml pull

# Restart services
docker compose -f deploy/docker-compose.yml up -d

echo "Deployment complete. Check logs with: docker compose -f deploy/docker-compose.yml logs -f"
```

## Automated Builds

Docker images are automatically built and pushed to GHCR when changes are pushed to the `docker-prod` branch:

- Images: `ghcr.io/thrillee/aegisbox/{migration,manager-api,smpp-gateway}`
- Tags: SHA-based tags + `latest`

## Available Make Commands

```bash
make tidy          # Format code and tidy dependencies
make audit         # Run quality control checks
make test          # Run all tests with race detector
make build         # Build both binaries
make production/build  # Build Linux AMD64 binaries
make docker/build  # Build Docker images locally
make docker/deploy # Deploy with Docker Compose
make docker/down   # Remove all services and volumes
make db/migrate    # Run database migrations
```

## Project Structure

```
aegisbox/
├── cmd/
│   ├── smpp-gateway/    # SMPP Gateway entrypoint
│   └── manager-api/     # Manager API entrypoint
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database operations
│   ├── sms/             # SMS handling
│   └── ...
├── deploy/
│   ├── smpp-gateway/    # Gateway Dockerfile
│   ├── manager-api/     # Manager API Dockerfile
│   ├── migration/       # Migration Dockerfile
│   └── docker-compose.yml
├── db/
│   ├── query/           # sqlc queries
│   └── migration/       # SQL migrations
└── .github/workflows/   # CI/CD pipelines
```
