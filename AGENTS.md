# Aegisbox Distribution Guide

This document provides comprehensive information about Aegisbox for AI agents and developers working with this deployment package.

## Project Overview

Aegisbox is an enterprise-grade SMPP (Short Message Peer-to-Peer) gateway and management platform. It enables organizations to:

- **Send and receive SMS** via SMPP protocol with mobile network operators (MNOs)
- **Manage gateway operations** through a REST API
- **Monitor system health** via Prometheus metrics and Grafana dashboards
- **Aggregate logs** from all services using Loki and Promtail

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Aegisbox Architecture                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────┐    ┌─────────────────┐    ┌─────────────────────────┐   │
│   │   Portal    │───▶│   Manager API   │───▶│     PostgreSQL         │   │
│   │  (Frontend) │    │   (REST API)    │    │   (Primary Database)    │   │
│   └─────────────┘    └────────┬────────┘    └─────────────────────────┘   │
│                               │                                             │
│   ┌─────────────┐            │                                             │
│   │   Grafana   │◀───────────┼─────────────────────────────────────────┐  │
│   │  (Dashboards)│           │                                         │  │
│   └─────────────┘           │                                         │  │
│                              ▼                                         │  │
│                    ┌─────────────────┐                                 │  │
│                    │  SMPP Gateway   │                                 │  │
│                    │  (SMS Routing)  │                                 │  │
│                    └────────┬────────┘                                 │  │
│                             │                                          │  │
│                             ▼                                          │  │
│                    ┌─────────────────┐                                  │  │
│                    │  MNO/SMSC       │◀─── SMPP Protocol ─────────────│  │
│                    │  Connections    │                                  │  │
│                    └─────────────────┘                                  │  │
│                              │                                          │  │
│                              │                                          │  │
│              ┌───────────────┼───────────────┐                          │  │
│              │               │               │                          │  │
│              ▼               ▼               ▼                          │  │
│        ┌──────────┐   ┌──────────┐   ┌──────────┐                      │  │
│        │Prometheus│   │   Loki   │   │Promtail │                      │  │
│        │(Metrics) │   │  (Logs)  │   │(Log Scr)│                      │  │
│        └──────────┘   └──────────┘   └──────────┘                      │  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Portal (Frontend)

**Image:** `ghcr.io/thrillee/aegisbox-portal:docker-prod`

A web-based UI for managing the Aegisbox platform. Built with React/Next.js.

**Environment Variables:**
- `PORTAL_API_URL` - Backend API URL (default: http://manager-api:8001)
- `PORTAL_BASE_URL` - Base path for routing (default: /)

**Ports:** 3001 (mapped to container port 3000)

---

### 2. Manager API (Backend)

**Image:** `ghcr.io/thrillee/aegisbox/manager-api:docker-prod`

REST API for managing the gateway. Built with Go + Gin framework.

**Features:**
- CRUD operations for MNO connections
- SMS routing configuration
- Segment management
- System health monitoring
- API key authentication

**Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string
- `API_ADDR` - Listen address (default: :8001)
- `LOG_LEVEL` - Logging level (info, debug, warn, error)

**Ports:** 8001

**Health Check:**
```bash
curl http://localhost:8001/health
```

---

### 3. SMPP Gateway (SMS Router)

**Image:** `ghcr.io/thrillee/aegisbox/smpp-gateway:docker-prod`

Core SMS routing engine. Handles SMPP connections to Mobile Network Operators.

**Features:**
- Bind to multiple MNOs (transceiver, transmitter, receiver modes)
- Automatic message routing based on destination MSISDN
- Connection health monitoring
- Reconnection logic with exponential backoff
- Message throttling and rate limiting
- DeliverSM / SubmitSM message handling

**SMPP Protocol Versions:** 3.3, 3.4, 5.0

**Environment Variables:**
- `SMPP_SYSTEM_ID` - System identifier for MNO connections
- `SMPP_PASSWORD` - Password for MNO authentication
- `SMPP_SYSTEM_TYPE` - System type (optional)
- `SMPP_BIND_MODE` - 0=receiver, 1=transmitter, 2=transceiver
- `SMPP_ADDR_TON` - Type of Number for source address
- `SMPP_ADDR_NPI` - Numbering Plan Indicator
- `SMPP_RECONNECT_DELAY` - Reconnection delay (default: 5s)
- `SMPP_ENQUIRE_LINK_INTERVAL` - Heartbeat interval (default: 30s)
- `SMPP_RESPONSE_TIMEOUT` - Response timeout (default: 10s)

**Ports:** 2775

---

### 4. Database Migration

**Image:** `ghcr.io/thrillee/aegisbox/migration:docker-prod`

Runs database schema migrations on startup. Uses goose for migrations.

**Behavior:**
- Runs once at startup (`restart: "no"`)
- Waits for PostgreSQL to be healthy before running
- Applies all migrations in `/migrations`
- Exits with success (0) or failure (1)

---

### 5. PostgreSQL

**Image:** `postgres:16-alpine`

Primary database storing all configuration and routing data.

**Ports:** 5432

---

### 6. Monitoring Stack

#### Prometheus

**Image:** `prom/prometheus:v3`

Collects and stores metrics from all services.

**Targets Scraped:**
- `prometheus:9090` - Self metrics
- `manager-api:8001/metrics` - API metrics
- `smpp-gateway:2775/metrics` - Gateway metrics

**Ports:** 9090

**Query Examples:**
```promql
# Gateway uptime
up{job="smpp-gateway"}

# Messages processed rate
rate(smpp_messages_processed_total[5m])

# Active MNO connections
smpp_active_connections
```

---

#### Loki

**Image:** `grafana/loki:3`

Horizontal log aggregation system.

**Sources:**
- Docker container logs (via Promtail)
- Application logs in `/var/log/aegisbox`
- System logs `/var/log/**/*.log`

**Ports:** 3100

**Query Examples:**
```logql
# All Aegisbox logs
{job="aegisbox-logs"}

# Error logs in last hour
{job="aegisbox-logs"} | json | level="error" | json.duration > 1000
```

---

#### Promtail

**Image:** `grafana/promtail:3`

Agent that scrapes logs and sends to Loki.

**Configuration File:** `./promtail/config.yml`

**Scrape Targets:**
- Docker container JSON logs
- Application-specific log files
- System log files

---

#### Grafana

**Image:** `grafana/grafana:11`

Visualization and alerting platform.

**Pre-configured:**
- Loki datasource (logs)
- Prometheus datasource (metrics)
- Aegisbox Overview dashboard

**Ports:** 3000

**Default Credentials:** admin/admin

---

## Deployment

### Prerequisites

- Docker Engine 24+
- Docker Compose V2
- Git
- 4GB+ RAM
- 10GB+ Disk space

### Quick Deployment

```bash
# 1. Clone
git clone https://github.com/thrillee/aegisbox-distribution.git
cd aegisbox-distribution

# 2. Configure
cp .env.example .env
# Edit .env with your configuration

# 3. Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <github-username> --password-stdin

# 4. Deploy
docker compose up -d

# 5. Verify
docker compose ps
```

### Environment Configuration

#### Required Changes

```bash
# Database
POSTGRES_PASSWORD=<secure-password>
DATABASE_URL=postgres://aegisbox:<password>@postgres:5432/aegisbox_prod?sslmode=disable

# SMPP Connection
SMPP_SYSTEM_ID=<your-system-id>
SMPP_PASSWORD=<your-password>

# Grafana
GRAFANA_PASSWORD=<secure-password>
```

#### Optional Changes

```bash
# Port mappings (if default ports conflict)
API_PORT=8001
SMPP_PORT=2775
PORTAL_PORT=3001
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
LOKI_PORT=3100
POSTGRES_PORT=5432
```

### Service Management

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose stop

# View logs
docker compose logs -f
docker compose logs -f smpp-gateway

# Restart a single service
docker compose restart smpp-gateway

# View service status
docker compose ps

# Update images and restart
docker compose pull
docker compose up -d
```

### Complete Cleanup

```bash
# Stop and remove containers, networks
docker compose down

# Stop and remove containers, networks, AND volumes
docker compose down -v
```

## Common Operations

### Adding an MNO Connection

Via Manager API:

```bash
curl -X POST http://localhost:8001/api/v1/mno-connections \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <api-key>" \
  -d '{
    "name": "MTN Nigeria",
    "host": "smpp.mtn.com",
    "port": 2775,
    "system_id": "mtn_gateway",
    "password": "secure_password",
    "system_type": "",
    "bind_mode": 2,
    "addr_ton": 1,
    "addr_npi": 1,
    "timeout": 30,
    "enquire_link_interval": 30
  }'
```

### Checking Gateway Status

```bash
# API health
curl http://localhost:8001/health

# SMPP gateway metrics
curl http://localhost:2775/metrics

# View connected MNOs
curl http://localhost:8001/api/v1/mno-connections \
  -H "Authorization: Bearer <api-key>"
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f smpp-gateway
docker compose logs -f manager-api

# Last 100 lines
docker compose logs --tail 100 smpp-gateway
```

### Monitoring

1. Open Grafana: http://localhost:3000
2. Login: admin/admin
3. Navigate to Dashboards → Aegisbox Overview
4. View real-time metrics and logs in Explore

## Troubleshooting

### Services Not Starting

```bash
# Check if ports are in use
lsof -i :8001

# Check Docker status
docker compose ps

# View detailed logs
docker compose logs
```

### Database Connection Issues

```bash
# Verify PostgreSQL is running
docker compose exec postgres pg_isready -U aegisbox

# Check database logs
docker compose logs postgres

# Test connection from another service
docker compose exec manager-api wget -qO- http://localhost:8001/health
```

### SMPP Connection Failures

```bash
# Check SMPP gateway logs
docker compose logs smpp-gateway

# Verify MNO credentials
docker compose exec smpp-gateway env | grep SMPP

# Test connectivity to MNO
docker compose exec smpp-gateway nc -zv smpp.mtn.com 2775
```

### Grafana Dashboards Empty

1. Check Loki datasource configuration
2. Verify Promtail is running and scraping logs
3. Check Loki logs via Explore

### High Memory Usage

```bash
# View container resource usage
docker stats

# Check disk usage
df -h

# Prune unused Docker resources
docker system prune -a
```

## Security Recommendations

1. **Change all default passwords** in `.env`
2. **Use secrets management** (Vault, AWS Secrets Manager) in production
3. **Enable SSL/TLS** for:
   - Manager API (reverse proxy with nginx/Caddy)
   - PostgreSQL (sslmode=require)
   - Grafana (HTTPS)
4. **Restrict network access** using firewall rules
5. **Use strong API keys** and rotate periodically
6. **Enable authentication** for Prometheus and Loki (optional)

## File Structure

```
aegisbox-distribution/
├── docker-compose.yml              # Main deployment configuration
├── .env.example                   # Environment variable template
├── .gitignore                     # Git ignore rules
├── README.md                       # Quick start guide
├── AGENTS.md                       # This file
│
├── prometheus/
│   └── prometheus.yml              # Prometheus scrape configuration
│
├── loki/
│   └── local-config.yaml          # Loki configuration
│
├── promtail/
│   └── config.yml                  # Log scraping configuration
│
└── grafana/
    └── provisioning/
        ├── datasources/
        │   └── ds.yml             # Datasource configurations
        └── dashboards/
            ├── dashboards.yml     # Dashboard provider config
            └── aegisbox-overview.json
```

## Image References

| Service | Image | Tag |
|---------|-------|-----|
| Portal | `ghcr.io/thrillee/aegisbox-portal` | `docker-prod` |
| Manager API | `ghcr.io/thrillee/aegisbox/manager-api` | `docker-prod` |
| SMPP Gateway | `ghcr.io/thrillee/aegisbox/smpp-gateway` | `docker-prod` |
| Migration | `ghcr.io/thrillee/aegisbox/migration` | `docker-prod` |
| PostgreSQL | `postgres` | `16-alpine` |
| Grafana | `grafana/grafana` | `11` |
| Prometheus | `prom/prometheus` | `v3` |
| Loki | `grafana/loki` | `3` |
| Promtail | `grafana/promtail` | `3` |

## Support

For issues with:
- **SMPP Gateway:** Check logs and verify MNO credentials
- **Manager API:** Verify PostgreSQL connectivity
- **Monitoring:** Check Prometheus and Loki services
- **Deployment:** Review Docker Compose configuration

## Related Repositories

- [aegisbox](https://github.com/thrillee/aegisbox) - Main source code
- [aegisbox-portal](https://github.com/thrillee/aegisbox-portal) - Frontend UI
- [aegisbox-distribution](https://github.com/thrillee/aegisbox-distribution) - This distribution package
