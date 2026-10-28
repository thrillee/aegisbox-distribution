# Aegisbox Distribution

Pre-built Docker Compose deployment for Aegisbox SMPP Gateway.

## Images

| Service | Image |
|---------|-------|
| Manager API | `ghcr.io/thrillee/aegisbox/manager-api:docker-prod` |
| SMPP Gateway | `ghcr.io/thrillee/aegisbox/smpp-gateway:docker-prod` |
| Migration | `ghcr.io/thrillee/aegisbox/migration:docker-prod` |
| Portal | `ghcr.io/thrillee/aegisbox-portal:docker-prod` |

## Deploy on Dokploy (no GHCR)

`docker-compose.dokploy.yml` runs the same stack but builds the Aegisbox apps from source (two images: `aegisbox` and `portal`), so the server needs no GHCR login. The app source is copied into this repository under `src/`, so Dokploy needs access to this repository only (no submodules):

| Path | Copied from |
|------|-------------|
| `src/aegisbox` | `thrillee/aegisbox`: one `aegisbox` image that runs migrations, the manager API and the SMPP gateway |
| `src/aegisbox-portal` | `thrillee/aegisbox-portal` (portal) |

`SOURCE_VERSIONS` records the commit each copy came from.

In Dokploy, create a Compose service from this repository and set:

* Branch: `main`
* Compose Path: `./docker-compose.dokploy.yml`
* Environment: the variables in `.env.example`. At minimum `DATABASE_URL`, `MANAGER_API_KEY` (protects the manager API; the portal sends it automatically), the Clerk keys, and the `NEXT_PUBLIC_*` values (built into the portal, so redeploy after changing them). Ports and internal URLs have working defaults.
* Domains: point them at the `aegisbox` service (container port 8000 for the SP HTTP API, 9001 for MNO DLRs) or `portal` (port 3000). There is no `smpp-gateway` or `manager-api` service any more.

On start the `aegisbox` container runs pending migrations, then the manager API and SMPP gateway. It keeps the old `manager-api` and `smpp-gateway` hostnames on the internal network, so the portal and Prometheus need no changes.

### Deploy newer app code

Do not edit `src/` by hand. Change the app repositories, then copy their latest `main` here:

```bash
scripts/sync-sources.sh
git add -A src SOURCE_VERSIONS
git commit -m "Sync app source"
git push
```

Pass refs to deploy something other than `main`: `scripts/sync-sources.sh <aegisbox-ref> <portal-ref>`.

## Prerequisites

- Docker Engine 24+
- Docker Compose V2
- 4GB+ RAM
- 10GB+ Disk space

## Quick Start

### 1. Clone This Repository

```bash
git clone https://github.com/thrillee/aegisbox-distribution.git
cd aegisbox-distribution
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

Required changes in `.env`:
- `POSTGRES_PASSWORD` - Set a secure password
- `DATABASE_URL` - Update if using external PostgreSQL
- `SMPP_SYSTEM_ID` / `SMPP_PASSWORD` - Your SMPP credentials
- `GRAFANA_PASSWORD` - Change admin password

### 3. Login to GHCR

```bash
# Replace <username> with your GitHub username
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

Generate a GitHub PAT with `read:packages` scope at:
https://github.com/settings/tokens

### 4. Start Services

```bash
docker compose up -d
```

### 5. Verify Deployment

```bash
# Check service status
docker compose ps

# View logs
docker compose logs -f
```

## Services

| Service | Port | URL |
|---------|------|-----|
| Manager API | 8001 | http://localhost:8001 |
| SMPP Gateway | 2775 | tcp://localhost:2775 |
| Portal | 3001 | http://localhost:3001 |
| Grafana | 3000 | http://localhost:3000 |
| Prometheus | 9090 | http://localhost:9090 |
| Loki | 3100 | http://localhost:3100 |
| PostgreSQL | 5432 | localhost:5432 |

**Default Grafana credentials:** admin/admin

## Stopping Services

```bash
# Stop without removing data
docker compose stop

# Stop and remove containers (data preserved)
docker compose down

# Stop and remove everything including volumes
docker compose down -v
```

## Updating Images

```bash
# Pull latest images
docker compose pull

# Restart with new images
docker compose up -d
```

## Customizing Tags

To use a different image tag:

```bash
TAG=sha-abcd1234 docker compose up -d
```

Or modify the `TAG` value in `.env`.

## Troubleshooting

### Services not starting

```bash
# Check logs
docker compose logs <service_name>

# Check PostgreSQL connectivity
docker compose exec postgres pg_isready -U aegisbox
```

### Grafana dashboards empty

Check Loki data source is configured correctly in Grafana → Connections → Data Sources.

### SMPP connection failing

Verify `SMPP_SYSTEM_ID` and `SMPP_PASSWORD` in `.env` match your provider's credentials.

## Monitoring

- **Grafana Dashboards:** http://localhost:3000 (Dashboard: Aegisbox Overview)
- **Prometheus Metrics:** http://localhost:9090
- **Log Explorer:** http://localhost:3000 → Explore → Loki

## Production Recommendations

1. **Change all default passwords**
2. **Use external PostgreSQL** for better reliability
3. **Configure SSL/TLS** for production traffic
4. **Set up backup** for PostgreSQL volume
5. **Use secrets management** instead of `.env` file
6. **Configure resource limits** in docker-compose.yml

## License

MIT
