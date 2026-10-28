# Aegisbox

Aegisbox is an intelligent SMPP gateway and manager API service written in Go. It features:
- **smpp-gateway**: Handles SMPP connections with service providers and routes SMS to MNOs with intelligent failover
- **manager-api**: REST API for managing the gateway (Gin-based)

## 🚀 Key Features

### Intelligent Routing System
- **Multi-Route Support**: Configure multiple MNO routes per destination prefix for redundancy
- **Automatic Failover**: Seamlessly switches to alternate routes on primary route failure
- **Health-Based Selection**: Routes chosen based on real-time performance metrics
- **Configurable Retry Policy**: Per-group maximum retry configuration (default: 2 retries + primary = 3 attempts)
- **DLR-Based Health Monitoring**: Tracks delivery rates using 1000-message ring buffer
- **Auto-Blocking**: Automatically blocks poorly performing sender IDs (below 50% delivery rate)
- **Automatic Recovery**: Routes auto-recover after consecutive successful deliveries
- **Safety Mechanisms**: Auto-unblock inactive senders after 7 days to prevent permanent blocking

### Background Workers
- **DLR Timeout Worker**: Marks segments with missing DLRs after 5 minutes (runs every minute)
- **Sender Auto-Unblock Worker**: Automatically unblocks inactive senders (runs every 6 hours)
- **Ring Buffer Cleanup Worker**: Maintains ring buffer by deleting old records (runs daily, 90-day retention)

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

## Single Binary

`cmd/aegisbox` runs the whole backend from one binary and one image (`deploy/aegisbox/Dockerfile`). Migrations are embedded in the binary.

| Command | What it runs |
|---------|--------------|
| `aegisbox` or `aegisbox all` | Pending migrations, then the manager API and SMPP gateway in one process |
| `aegisbox migrate` | Migrations only, then exits |
| `aegisbox api` | Manager API only |
| `aegisbox gateway` | SMPP gateway only |

In `all` mode, if the API or the gateway fails, both shut down and the process exits with an error so the container restarts.

Set `MANAGER_API_KEY` in production. Every manager API route except `/health` and the MNO DLR webhooks (`/api/v1/webhooks/*`) then requires it in the `X-API-Key` header (or `Authorization: Bearer`). Without it the API is open to anyone who can reach it, including wallet credits and manual DLRs, and a warning is logged at startup. Locally: `make run/aegisbox`. The separate `cmd/manager-api`, `cmd/smpp-gateway` and `cmd/migration` entrypoints still work and run the same code.

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

## Testing

### Run All Tests

```bash
# Run all tests with race detector
make test

# Run tests with coverage
make test/cover

# Run specific package tests
go test -v ./internal/sms/...
```

### Intelligent Routing Tests

The failover logic is covered by comprehensive unit tests:

```bash
# Run failover-specific tests
go test -v ./internal/sms/ -run "TestFailover|TestIntelligent|TestMaxRetry"
```

**Test Coverage:**
- ✅ Single route success (no failover needed)
- ✅ Failover to second route after primary fails
- ✅ Failover to third route after first two fail
- ✅ All routes exhausted (failure handling)
- ✅ Router candidate selection
- ✅ Max retry limits respected

**Sample Test Output:**
```
=== RUN   TestFailoverWithMultipleRoutes
=== RUN   TestFailoverWithMultipleRoutes/First_route_succeeds_-_no_failover
=== RUN   TestFailoverWithMultipleRoutes/First_route_fails,_second_succeeds
=== RUN   TestFailoverWithMultipleRoutes/First_two_routes_fail,_third_succeeds
=== RUN   TestFailoverWithMultipleRoutes/All_routes_fail
--- PASS: TestFailoverWithMultipleRoutes (0.00s)
=== RUN   TestIntelligentRouterCandidateSelection
--- PASS: TestIntelligentRouterCandidateSelection (0.00s)
=== RUN   TestMaxRetryRespected
--- PASS: TestMaxRetryRespected (0.00s)
PASS
```

## Project Structure

```
aegisbox/
├── cmd/
│   ├── smpp-gateway/    # SMPP Gateway entrypoint
│   └── manager-api/     # Manager API entrypoint
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database operations (sqlc generated)
│   ├── sms/             # SMS processing & intelligent routing
│   │   ├── processor.go              # Message processing pipeline
│   │   ├── intelligent_router.go     # Multi-route selection
│   │   ├── sender_health_tracker.go  # Sender ID performance tracking
│   │   ├── route_health_tracker.go   # Route-level health monitoring
│   │   ├── dlr_timeout_worker.go     # DLR timeout detection
│   │   ├── sender_auto_unblock_worker.go  # Auto-unblock inactive senders
│   │   └── ring_buffer_cleanup_worker.go  # Ring buffer maintenance
│   └── ...
├── deploy/
│   ├── smpp-gateway/    # Gateway Dockerfile
│   ├── manager-api/     # Manager API Dockerfile
│   ├── migration/       # Migration Dockerfile
│   └── docker-compose.yml
├── db/
│   ├── query/           # sqlc queries (including routing_flex.sql)
│   └── migration/       # SQL migrations (including intelligent routing)
└── .github/workflows/   # CI/CD pipelines
```

## Intelligent Routing Configuration

### Database Schema

The intelligent routing system uses these key tables:

- **routing_assignments**: Defines MNO routes for prefix groups (priority, weight, max retries)
- **route_health**: Tracks route-level success/failure rates and availability
- **sender_id_health**: Tracks sender ID performance per MNO
- **route_delivery_attempts**: Ring buffer for route health calculation (1000 last attempts)
- **sender_id_delivery_attempts**: Ring buffer for sender health calculation (1000 last attempts)
- **routing_config**: Configuration per routing group (strategies, thresholds, timeouts)
- **health_ring_buffer_config**: Global health tracking configuration

### Routing Strategies

Configure via `routing_config.selection_strategy`:

| Strategy | Behavior |
|----------|----------|
| `priority_weighted` | Selects by priority first, then weighted random within same priority (default) |
| `health_based` | Selects route with highest health score |
| `round_robin` | Distributes traffic evenly across all viable routes |

### Configuration Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `max_route_retries` | 2 | Number of alternate routes to try on failure (3 total attempts) |
| `dlr_timeout_seconds` | 300 | Seconds to wait for DLR before marking as timeout |
| `sender_block_threshold` | 0.50 | Minimum delivery rate to avoid auto-blocking (50%) |
| `sender_unblock_threshold` | 0.70 | Delivery rate required for manual unblock (70%) |
| `route_availability_threshold` | 0.70 | Minimum health score for route to be considered viable (70%) |
| `health_calculation_window` | 1000 | Number of messages to use for health calculation |
| `route_failure_threshold` | 5 | Consecutive failures before route marked unavailable |
| `route_recovery_threshold` | 3 | Consecutive successes before route marked available |

### Example: Configuring Multi-Route Failover

```sql
-- 1. Create routing group for a service provider
INSERT INTO routing_groups (name, reference) 
VALUES ('Premium Routes', 'premium-routes');

-- 2. Configure routing settings
INSERT INTO routing_config (routing_group_id, selection_strategy, max_route_retries)
VALUES (1, 'priority_weighted', 2);  -- Try up to 3 routes total

-- 3. Define multiple routes for the same destination
INSERT INTO routing_assignments (routing_group_id, msisdn_prefix_group_id, mno_id, priority, weight, max_retry_attempts)
VALUES 
  (1, 1, 1, 1, 100, 2),  -- Primary: MNO-1, Priority 1
  (1, 1, 2, 2, 80, 2),   -- Backup 1: MNO-2, Priority 2
  (1, 1, 3, 3, 60, 2);   -- Backup 2: MNO-3, Priority 3

-- The system will automatically:
-- - Try MNO-1 first
-- - If MNO-1 fails, try MNO-2
-- - If MNO-2 fails, try MNO-3
-- - If all fail, mark message as failed and reverse debit
```

### Monitoring Route Health

```sql
-- View route health status
SELECT 
  ra.id AS assignment_id,
  m.name AS mno_name,
  rh.health_score,
  rh.success_rate,
  rh.total_attempts,
  rh.consecutive_failures,
  rh.is_available,
  rh.last_attempt_at
FROM route_health rh
JOIN routing_assignments ra ON ra.id = rh.routing_assignment_id
JOIN mnos m ON m.id = ra.mno_id
ORDER BY rh.health_score DESC;

-- View sender ID health
SELECT 
  si.sender_id,
  m.name AS mno_name,
  sh.delivery_rate,
  sh.dlr_response_rate,
  sh.health_score,
  sh.is_blocked,
  sh.blocked_at
FROM sender_id_health sh
JOIN approved_sender_ids si ON si.id = sh.sender_id
JOIN mnos m ON m.id = sh.mno_id
WHERE sh.is_blocked = true;
```

## Failover Behavior

### Successful Send (No Failover)
```
Route Selection → Send to Primary MNO → Success
                                      ↓
                  Record Route Health (success=true)
                                      ↓
                  Mark Message as send_attempted
```

### Failover Scenario
```
Route Selection → Try MNO-1 → Fail
                            ↓
                  Record Route Health (success=false)
                            ↓
                  Try MNO-2 → Fail
                            ↓
                  Record Route Health (success=false)
                            ↓
                  Try MNO-3 → Success
                            ↓
                  Record Route Health (success=true)
                            ↓
                  Mark Message as send_attempted
```

### All Routes Fail
```
Try All Routes → All Fail
               ↓
Record Route Health for all attempts (success=false)
               ↓
Mark Message as failed
               ↓
Reverse wallet debit
               ↓
Generate failure DLR
```

## Log Messages

Watch for these key log entries during operation:

```log
# System Initialization
INFO "Initializing health trackers..."
INFO "Initializing intelligent routing workers..."
INFO "DLR timeout worker started" interval=1m timeout_seconds=300
INFO "Sender auto-unblock worker started" interval=6h
INFO "Ring buffer cleanup worker started" interval=24h

# Routing & Failover
INFO "Retrieved routing candidates for failover" total_candidates=3 max_retries=2
INFO "Attempting failover to alternate route" attempt=2 mno_id=2 mno_name="MNO-2"
INFO "Send attempt completed" attempt=2 was_submitted=true
INFO "Message sent successfully" attempt=2 retry_attempts=1

# Health Tracking
INFO "Recorded delivery result in health tracker" sender_id=123 mno_id=2 status="delivered"
INFO "Route health updated" assignment_id=456 success=true health_score=0.85
WARN "Sender ID auto-blocked due to low delivery rate" sender_id=789 mno_id=3 delivery_rate=0.45
INFO "Route marked unavailable after consecutive failures" assignment_id=101 consecutive_failures=5

# Worker Activity
INFO "Processing DLR timeouts" found=15 marked_timeout=15
INFO "Auto-unblocked inactive senders" count=3
INFO "Ring buffer cleanup completed" deleted_route_attempts=15234 deleted_sender_attempts=23456
```
