# Aegisbox SMS Gateway

*Automatically synced with your [v0.app](https://v0.app) deployments*

[![Deployed on Vercel](https://img.shields.io/badge/Deployed%20on-Vercel-black?style=for-the-badge&logo=vercel)](https://vercel.com/thrillees-projects/v0-aegisbox-sms-gateway)
[![Built with v0](https://img.shields.io/badge/Built%20with-v0.app-black?style=for-the-badge)](https://v0.app/chat/0CuEiqxl4rn)

## Overview

This repository will stay in sync with your deployed chats on [v0.app](https://v0.app).
Any changes you make to your deployed app will be automatically pushed to this repository from [v0.app](https://v0.app).

## Deployment

### Vercel (Recommended)

Your project is live at:

**[https://vercel.com/thrillees-projects/v0-aegisbox-sms-gateway](https://vercel.com/thrillees-projects/v0-aegisbox-sms-gateway)**

### Docker

The application can also be deployed using Docker.

**Prerequisites:**
- Docker installed
- `.env.local` file with required environment variables

**Environment Variables Required:**
```
DATABASE_URL          # Neon PostgreSQL connection string
CLERK_PUBLISHABLE_KEY # Clerk publishable key
CLERK_SECRET_KEY      # Clerk secret key
CLERK_ADMIN_ORG_ID    # Clerk admin organization ID
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
```

**Quick Start:**

```bash
# Build and run with docker compose
docker compose up -d

# Or build and run manually
docker build -t aegisbox-portal .
docker run -d -p 3000:3000 --env-file .env.local aegisbox-portal
```

The application will be available at `http://localhost:3000`

**Useful Commands:**

```bash
# View logs
docker compose logs -f

# Stop containers
docker compose down

# Rebuild after code changes
docker compose build --no-cache && docker compose up -d
```

### GitHub Container Registry (GHCR)

The Docker image is automatically built and pushed to GHCR when code is pushed to the `docker-prod` branch.

**Image Location:**
```
ghcr.io/<github-owner>/<repo-name>:docker-prod
```

**Supported Platforms:**
- `linux/amd64` (x86_64)
- `linux/arm64` (aarch64)

Docker will automatically pull the correct platform image based on your system.

**Pull and Run:**

```bash
# Authenticate with GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <github-username> --password-stdin

# Pull the latest image
docker pull ghcr.io/<github-owner>/<repo-name>:docker-prod

# Run the container
docker run -d -p 3000:3000 --env-file .env.local ghcr.io/<github-owner>/<repo-name>:docker-prod
```

**Get your GitHub Token:**
1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate new token with `read:packages` scope
3. Use the token for authentication

**Note:** The image is tagged as `docker-prod` on each push to the `docker-prod` branch.

## Build your app

Continue building your app on:

**[https://v0.app/chat/0CuEiqxl4rn](https://v0.app/chat/0CuEiqxl4rn)**

## How It Works

1. Create and modify your project using [v0.app](https://v0.app)
2. Deploy your chats from the v0 interface
3. Changes are automatically pushed to this repository
4. Vercel deploys the latest version from this repository