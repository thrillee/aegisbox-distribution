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

The Docker images are hosted on GitHub Container Registry (GHCR). To pull them on any server:

#### 1. Generate a GitHub Personal Access Token

Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic) and create a token with:
- `read:packages` scope

#### 2. Login to GHCR

```bash
# Replace <username> with your GitHub username
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

#### 3. Pull Images

```bash
# Pull all three images
docker pull ghcr.io/thrillee/aegisbox/migration:latest
docker pull ghcr.io/thrillee/aegisbox/manager-api:latest
docker pull ghcr.io/thrillee/aegisbox/smpp-gateway:latest
```

#### 4. Verify Pulled Images

```bash
docker images | grep ghcr.io/thrillee/aegisbox
```

Expected output:
```
ghcr.io/thrillee/aegisbox/migration       latest    ...
ghcr.io/thrillee/aegisbox/manager-api     latest    ...
ghcr.io/thrillee/aegisbox/smpp-gateway    latest    ...
```

#### Pull Without Updating docker-compose.yml

If you want to use GHCR images without modifying your local `docker-compose.yml`, tag them locally:

```bash
docker tag ghcr.io/thrillee/aegisbox/migration:latest migration:latest
docker tag ghcr.io/thrillee/aegisbox/manager-api:latest manager-api:latest
docker tag ghcr.io/thrillee/aegisbox/smpp-gateway:latest smpp-gateway:latest
```

### Deploy with Docker Compose

#### Option 1: Use Pre-built GHCR Images (Recommended)

Create `deploy/docker-compose.override.yml`:

```bash
cd deploy
cat > docker-compose.override.yml << 'EOF'
services:
  migration:
    image: ghcr.io/thrillee/aegisbox/migration:latest
  manager-api:
    image: ghcr.io/thrillee/aegisbox/manager-api:latest
  smpp-gateway:
    image: ghcr.io/thrillee/aegisbox/smpp-gateway:latest
EOF
```

Then deploy:

```bash
# On the remote server
git clone https://github.com/thrillee/aegisbox.git
cd aegisbox
git checkout docker-prod

# Copy and configure environment
cp deploy/.env.example deploy/.env
vim deploy/.env  # Update DATABASE_URL and other values

# Start services (docker-compose will use override.yml for image references)
cd deploy
docker compose pull
docker compose up -d

# View logs
docker compose logs -f
```

#### Option 2: Modify docker-compose.yml Directly

```bash
# On the remote server
git clone https://github.com/thrillee/aegisbox.git
cd aegisbox
git checkout docker-prod

# Copy and configure environment
cp deploy/.env.example deploy/.env
vim deploy/.env  # Update DATABASE_URL and other values

# Edit docker-compose.yml and update image lines:
# - image: migration:latest        →  - image: ghcr.io/thrillee/aegisbox/migration:latest
# - image: manager-api:latest      →  - image: ghcr.io/thrillee/aegisbox/manager-api:latest
# - image: smpp-gateway:latest     →  - image: ghcr.io/thrillee/aegisbox/smpp-gateway:latest

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
set -e

REPO_DIR="/opt/aegisbox"
cd $REPO_DIR

# Backup current environment
cp deploy/.env deploy/.env.backup 2>/dev/null || true

# Pull latest changes from docker-prod branch
git fetch origin docker-prod
git checkout docker-prod
git pull origin docker-prod

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Update Docker images using override file
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.override.yml pull

# Restart services
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.override.yml up -d

echo "Deployment complete!"
echo "View logs: docker compose -f deploy/docker-compose.yml logs -f"
```

## Automated Builds

Docker images are automatically built and pushed to GHCR when code is pushed to the `docker-prod` branch.

### Image References

| Service | GHCR Image |
|---------|------------|
| Migration | `ghcr.io/thrillee/aegisbox/migration` |
| Manager API | `ghcr.io/thrillee/aegisbox/manager-api` |
| SMPP Gateway | `ghcr.io/thrillee/aegisbox/smpp-gateway` |

### Tag Format

Images are tagged with:
- `sha-{short_sha}` - Each commit gets a unique tag (e.g., `sha-abc1234`)
- `docker-prod` - Points to latest `docker-prod` branch build

To pull a specific image by SHA, check the GitHub Actions workflow run.

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
