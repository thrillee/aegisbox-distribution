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
- Provider Health dashboard (NEW)

**Ports:** 3000

**Default Credentials:** admin/admin

---

#### Alertmanager

**Image:** `prom/alertmanager:latest`

Alert routing and notification manager for Prometheus.

**Features:**
- Alert deduplication and grouping
- Email notifications for critical alerts
- Slack integration (configurable)
- Alert inhibition rules to reduce noise
- Silencing and maintenance windows

**Ports:** 9093

**Configuration File:** `./alertmanager/alertmanager.yml`

**Alert Categories:**
- **Critical:** Service down, circuit breaker open, connection failures
- **Warning:** High latency, failure rate spikes, resource usage
- **Info:** Low traffic, protocol imbalance notifications

**Notification Channels (Configurable):**
- Email (default)
- Slack webhooks
- PagerDuty integration
- Custom webhooks

---

## Monitoring & Observability

### Overview

Aegisbox includes a comprehensive monitoring stack that provides:
- **Real-time metrics** via Prometheus
- **Log aggregation** via Loki
- **Visualization** via Grafana dashboards
- **Alerting** via Alertmanager
- **HTTP provider monitoring** for modern SMS APIs
- **Circuit breaker tracking** for reliability insights

### Key Metrics Tracked

#### 1. Message Processing Metrics

```promql
# Total messages submitted (SMPP + HTTP)
sum(rate(sms_messages_submitted_total[5m]))

# Success rate
sum(rate(sms_messages_submitted_total[5m])) / 
  (sum(rate(sms_messages_submitted_total[5m])) + sum(rate(sms_messages_failed_total[5m])))

# Messages by protocol
sum(rate(sms_messages_submitted_total[5m])) by (protocol)

# Messages by provider (HTTP)
sum(rate(http_provider_messages_sent_total[5m])) by (provider)
```

#### 2. HTTP Provider Metrics

```promql
# Provider status (http_ok, error, disconnected)
http_provider_status{provider="TERMII"}

# Provider latency (95th percentile)
histogram_quantile(0.95, rate(http_provider_latency_seconds_bucket[5m]))

# Provider failure rate
rate(http_provider_messages_failed_total[5m]) / 
  rate(http_provider_messages_sent_total[5m])

# DLR webhook success
rate(dlr_webhook_received_total[5m])
```

#### 3. Circuit Breaker Metrics

```promql
# Circuit breaker state (0=closed, 1=open/half-open)
circuit_breaker_state{mno_id="1", state="closed"}

# Failure count
rate(circuit_breaker_failures_total[5m])

# Success count
rate(circuit_breaker_successes_total[5m])

# State changes over time
changes(circuit_breaker_state[10m])
```

#### 4. SMPP Connection Metrics

```promql
# Connection status
smpp_connection_status{mno_id="1", status="bound"}

# SMPP latency
histogram_quantile(0.95, rate(smpp_submit_sm_duration_seconds_bucket[5m]))

# Connection uptime
up{job="smpp-gateway"}
```

### Grafana Dashboards

#### 1. Aegisbox Overview Dashboard

Basic system health overview:
- Service status (Manager API, SMPP Gateway)
- Overall message throughput
- System resource usage
- Quick health checks

**Access:** http://localhost:3001/d/aegisbox-overview

#### 2. Provider Health Dashboard (NEW)

Comprehensive provider and circuit breaker monitoring:

**System Overview Row:**
- Total messages/minute
- Overall success rate
- Gateway status
- API status

**HTTP Providers Row:**
- Messages per provider (TERMII, AfricaTalking, etc.)
- Provider status table
- Latency graphs (p95)
- Failure rate by provider

**Circuit Breakers Row:**
- Circuit breaker state table
- Failures & successes timeline
- State transition events

**Protocol Distribution Row:**
- Traffic pie chart (SMPP vs HTTP)
- Messages/min by protocol (stacked)

**Active Alerts Row:**
- Live alert table with severity colors
- Grouped by criticality

**Access:** http://localhost:3001/d/aegisbox-provider-health

### Alert Rules

Aegisbox includes 60+ pre-configured alert rules across 10 categories:

#### Service Health Alerts
- `ServiceDown` - Any service unreachable for >2min
- `ManagerAPIDown` - API unavailable for >1min
- `SMPPGatewayDown` - Gateway unavailable for >1min

#### SMPP Connection Alerts
- `SMPPConnectionDown` - MNO disconnected for >5min
- `SMPPConnectionFlapping` - >5 state changes in 10min
- `SMPPHighLatency` - p95 latency >5s for 5min

#### HTTP Provider Alerts
- `HTTPProviderDown` - Provider unavailable for >5min
- `HTTPProviderHighFailureRate` - >10% failures for 5min
- `HTTPProviderHighLatency` - p95 >10s for 5min
- `HTTPProviderNoTraffic` - No messages for 30min

#### Circuit Breaker Alerts
- `CircuitBreakerOpen` - Open for >5min (critical)
- `CircuitBreakerHalfOpen` - Stuck in half-open for >10min
- `CircuitBreakerHighFailureCount` - >10 failures/sec for 5min

#### Message Processing Alerts
- `HighMessageFailureRate` - >10% global failures for 10min
- `MessageProcessingStalled` - No messages for 15min
- `MessageQueueBacklog` - >1000 pending messages

#### DLR Webhook Alerts
- `DLRWebhookFailures` - >1 failure/sec for 10min
- `LowDLRRate` - <50% DLR rate for 2 hours

#### System Resource Alerts
- `HighMemoryUsage` - >2GB usage for 10min
- `HighGoroutineCount` - >1000 goroutines for 10min
- `HighErrorRate` - >10 5xx errors/sec

#### Database Alerts
- `DatabaseConnectionPoolExhausted` - No connections for 2min
- `HighDatabaseLatency` - p95 >1s for 5min
- `DatabaseConnectionErrors` - >1 error/sec

### Configuring Alert Notifications

#### Email Notifications

Edit `alertmanager/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'aegisbox-alerts@your-domain.com'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'
  smtp_require_tls: true

receivers:
  - name: 'critical-alerts'
    email_configs:
      - to: 'oncall@your-domain.com'
        send_resolved: true
```

**For Gmail:**
1. Enable 2FA on your Google account
2. Generate an App Password: https://myaccount.google.com/apppasswords
3. Use the app password in `smtp_auth_password`

#### Slack Notifications

Add Slack webhook configuration:

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'

receivers:
  - name: 'critical-alerts'
    slack_configs:
      - channel: '#aegisbox-critical'
        title: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true
```

**Get Slack Webhook:**
1. Go to https://api.slack.com/messaging/webhooks
2. Create an incoming webhook for your workspace
3. Copy the webhook URL

#### PagerDuty Integration

For critical production alerts:

```yaml
receivers:
  - name: 'critical-alerts'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        description: '{{ .GroupLabels.alertname }}'
```

### Monitoring Best Practices

#### 1. Regular Dashboard Review

**Daily:**
- Check Provider Health dashboard
- Review active alerts in Alertmanager
- Verify all services are up
- Check circuit breaker states

**Weekly:**
- Analyze message throughput trends
- Review failure rate patterns
- Check DLR reception rates
- Review resource usage trends

**Monthly:**
- Capacity planning based on metrics
- Alert rule tuning
- Dashboard refinements

#### 2. Alert Fatigue Management

Aegisbox includes alert inhibition rules to reduce noise:

```yaml
inhibit_rules:
  # If service is down, don't alert on latency
  - source_match:
      alertname: 'ServiceDown'
    target_match:
      severity: 'warning'
    equal: ['job']
  
  # If circuit breaker is open, don't alert on failures
  - source_match:
      alertname: 'CircuitBreakerOpen'
    target_match:
      alertname: 'CircuitBreakerHighFailureCount'
    equal: ['mno_id']
```

**Tips:**
- Tune alert thresholds based on your traffic patterns
- Use `group_wait` and `group_interval` to batch related alerts
- Silence alerts during maintenance windows
- Adjust `repeat_interval` for non-critical alerts

#### 3. Circuit Breaker Monitoring

Circuit breakers protect your system from cascading failures:

**States:**
- **Closed** (green): Normal operation, all requests go through
- **Open** (red): Too many failures, requests blocked
- **Half-Open** (yellow): Testing if service recovered

**When Circuit Breaker Opens:**
1. Check Provider Health dashboard for the affected MNO/provider
2. Review logs in Loki: `{job="aegisbox-logs"} | json | mno_id="X" | level="error"`
3. Verify network connectivity to provider
4. Check provider's status page
5. Wait for automatic recovery or manually intervene

**Configuration** (in main application):
```go
CircuitBreakerConfig{
    FailureThreshold: 5,    // Open after 5 consecutive failures
    SuccessThreshold: 3,    // Close after 3 consecutive successes
    Timeout:          30s,  // Wait 30s before trying half-open
    RequestTimeout:   10s,  // Individual request timeout
}
```

#### 4. HTTP Provider Health

**Key Indicators:**
- **Status**: Should always show "✓ OK"
- **Latency**: Typically <2s, alert at >10s
- **Failure Rate**: Should be <5%, alert at >10%
- **Traffic**: Verify expected message volume

**Troubleshooting High Latency:**
1. Check provider's API status page
2. Review network latency to provider
3. Check if API rate limits are being hit
4. Verify request payload size

**Troubleshooting High Failure Rate:**
1. Check provider error responses in logs
2. Verify API credentials haven't expired
3. Check account balance/credits
4. Review recent API changes from provider

#### 5. Log Analysis with Loki

**Useful Queries:**

```logql
# All errors in last hour
{job="aegisbox-logs"} | json | level="error" 

# HTTP provider errors
{job="aegisbox-logs"} | json | provider=~"TERMII|AfricaTalking" | level="error"

# Circuit breaker events
{job="aegisbox-logs"} | json | msg=~".*circuit.*"

# Slow requests (>1s)
{job="aegisbox-logs"} | json | json.duration > 1000

# Failed message submissions
{job="aegisbox-logs"} | json | msg=~".*submit.*failed.*"

# DLR webhook issues
{job="aegisbox-logs"} | json | msg=~".*webhook.*" | level="error"
```

### Maintenance Windows

To silence alerts during planned maintenance:

```bash
# Silence all alerts for 1 hour
amtool silence add alertname=~".+" --duration=1h --comment="Planned maintenance"

# Silence specific service
amtool silence add job="smpp-gateway" --duration=2h --comment="Gateway upgrade"

# Silence specific provider
amtool silence add provider="TERMII" --duration=30m --comment="Testing"

# List active silences
amtool silence query

# Expire silence early
amtool silence expire <silence-id>
```

### Accessing Monitoring Tools

| Tool | URL | Credentials | Purpose |
|------|-----|-------------|---------|
| Grafana | http://localhost:3001 | admin/admin | Dashboards & visualization |
| Prometheus | http://localhost:9090 | None | Raw metrics & PromQL queries |
| Alertmanager | http://localhost:9093 | None | Alert management & silences |
| Loki (via Grafana) | Grafana → Explore | - | Log queries |

**Security Note:** Change default Grafana password in production via `.env`:
```bash
GRAFANA_USER=admin
GRAFANA_PASSWORD=your-secure-password-here
```

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
│   ├── prometheus.yml              # Prometheus scrape configuration
│   └── alerts.yml                 # Alert rules (NEW)
│
├── alertmanager/
│   ├── alertmanager.yml           # Alertmanager configuration (NEW)
│   └── templates/
│       └── email.tmpl             # Email alert templates (NEW)
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
            ├── aegisbox-overview.json
            └── provider-health.json  # Provider health dashboard (NEW)
```

## Image References

| Service | Image | Tag |
|---------|-------|-----|
| Portal | `ghcr.io/thrillee/aegisbox-portal` | `docker-prod` |
| Manager API | `ghcr.io/thrillee/aegisbox/manager-api` | `docker-prod` |
| SMPP Gateway | `ghcr.io/thrillee/aegisbox/smpp-gateway` | `docker-prod` |
| Migration | `ghcr.io/thrillee/aegisbox/migration` | `docker-prod` |
| PostgreSQL | `postgres` | `16-alpine` |
| Grafana | `grafana/grafana` | `latest` |
| Prometheus | `prom/prometheus` | `latest` |
| Alertmanager | `prom/alertmanager` | `latest` |
| Loki | `grafana/loki` | `latest` |
| Promtail | `grafana/promtail` | `latest` |

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
