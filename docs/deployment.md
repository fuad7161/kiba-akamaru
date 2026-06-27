# BD Govt Job Circular API — Production Deployment Guide

> **Stack:** Go · PostgreSQL · Redis · MinIO · Docker Compose · Traefik · DigitalOcean Container Registry  
> **Target:** DigitalOcean Droplet (Ubuntu)

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [DigitalOcean Container Registry Setup](#2-digitalocean-container-registry-setup)
3. [Build & Push the Docker Image](#3-build--push-the-docker-image)
4. [Server Setup](#4-server-setup)
5. [Configure the Production `.env`](#5-configure-the-production-env)
6. [Required Changes to `docker-compose.prod.yml`](#6-required-changes-to-docker-composeprodYml)
7. [Deploy the Stack](#7-deploy-the-stack)
8. [Run Database Migrations](#8-run-database-migrations)
9. [Create the MinIO Bucket](#9-create-the-minio-bucket)
10. [Verify the Deployment](#10-verify-the-deployment)
11. [Ongoing Deployments (CI/CD)](#11-ongoing-deployments-cicd)
12. [Useful Makefile Commands](#12-useful-makefile-commands)
13. [Troubleshooting](#13-troubleshooting)

---

## 1. Prerequisites

### Local machine
- Docker + Docker Compose v2 installed
- `doctl` (DigitalOcean CLI) installed and authenticated
- A DigitalOcean account with a **Container Registry** created
- A domain name with DNS pointing to your Droplet IP

### DigitalOcean Droplet (Ubuntu 22.04+ recommended)
- Minimum: **2 vCPU / 2 GB RAM** (4 GB recommended for production)
- Docker and Docker Compose v2 installed (see [Section 4](#4-server-setup))
- Ports **80** and **443** open in the Droplet firewall

---

## 2. DigitalOcean Container Registry Setup

### 2a. Install & authenticate `doctl` (on your local machine)

```bash
# Install doctl (Linux)
sudo snap install doctl

# Authenticate with your DigitalOcean API token
doctl auth init

# Authenticate Docker to the DigitalOcean registry
doctl registry login
```

### 2b. Note your registry name

Your registry URL follows this pattern:
```
registry.digitalocean.com/<YOUR_REGISTRY_NAME>
```

You will use `<YOUR_REGISTRY_NAME>` as the value for `DO_REGISTRY` in your `.env`.

---

## 3. Build & Push the Docker Image

Run these commands **on your local machine** from the project root:

```bash
# Set your registry name
export DO_REGISTRY=your-registry-name

# Build the image
docker build -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest .

# Push to DigitalOcean Container Registry
docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

> **Tip:** Tag with a version alongside `latest` for rollback capability:
> ```bash
> docker build -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:v1.0.0 \
>              -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest .
> docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:v1.0.0
> docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
> ```

---

## 4. Server Setup

SSH into your Droplet and run:

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install Docker
sudo apt install -y docker.io

# Install Docker Compose v2 plugin
sudo apt install -y docker-compose-plugin

# Allow current user to run Docker without sudo
sudo usermod -aG docker $USER
newgrp docker

# Install Make
sudo apt install -y make

# Install doctl to authenticate registry pulls on the server
sudo snap install doctl
doctl auth init     # paste your DigitalOcean API token when prompted
doctl registry login
```

### Clone the repository on the server

```bash
git clone https://github.com/fuad71/job-circular-api.git /opt/job-circuler
cd /opt/job-circuler
```

---

## 5. Configure the Production `.env`

On the **server**, create the `.env` file in `/opt/job-circuler/`:

```bash
cp .env.example .env
nano .env
```

Fill in **all** values below. Do **not** leave any placeholder unchanged.

```env
# Server
APP_PORT=8080
APP_ENV=production

# PostgreSQL
DB_HOST=postgres
DB_PORT=5432
DB_NAME=bdgovtjobs
DB_USER=bduser
DB_PASSWORD=<generate: openssl rand -base64 24>
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis
REDIS_PASSWORD=<generate: openssl rand -base64 24>
REDIS_URL=redis://:<same_redis_password>@redis:6379/0

# JWT
JWT_SECRET=<generate: openssl rand -hex 32>
JWT_REFRESH_SECRET=<generate: openssl rand -hex 32>
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASS=your_gmail_app_password
FROM_EMAIL=noreply@yourdomain.com

# Frontend
FRONTEND_URL=https://yourdomain.com

# MinIO
MINIO_USER=<choose a non-default admin username>
MINIO_PASSWORD=<generate: openssl rand -base64 24>
MINIO_BUCKET=bd-govt-jobs-assets
MINIO_REGION=us-east-1
MINIO_ACCESS_KEY=<same as MINIO_USER>
MINIO_SECRET_KEY=<same as MINIO_PASSWORD>
MINIO_ENDPOINT=http://minio:9000

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Production (DigitalOcean)
DO_REGISTRY=your-registry-name
ACME_EMAIL=admin@yourdomain.com
```

### Generate secure values quickly

```bash
openssl rand -hex 32                                  # for JWT secrets
openssl rand -base64 24 | tr -d '+/=' | cut -c1-24   # for passwords
```

> **Warning:** Never commit `.env` to Git. It is already listed in `.gitignore`.

---

## 6. Required Changes to `docker-compose.prod.yml`

> **Important:** Before deploying, you **must** replace all `yourdomain.com` placeholders
> in `docker-compose.prod.yml` with your actual domain names.

Open the file and update these lines:

| Placeholder | Replace with example |
|---|---|
| `yourdomain.com` (api router) | `api.yourdomain.com` |
| `s3.yourdomain.com` (minio-api router) | `s3.yourdomain.com` |
| `minio.yourdomain.com` (minio-console router) | `minio.yourdomain.com` |

### Updated `docker-compose.prod.yml` (with your actual domains)

```yaml
version: "3.9"

services:
  traefik:
    image: traefik:v3.0
    command:
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.le.acme.tlschallenge=true"
      - "--certificatesresolvers.le.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.le.acme.storage=/certs/acme.json"
      - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_certs:/certs
    restart: unless-stopped

  minio:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.minio-api.rule=Host(`s3.yourdomain.com`)"        # <-- CHANGE
      - "traefik.http.routers.minio-api.entrypoints=websecure"
      - "traefik.http.routers.minio-api.tls.certresolver=le"
      - "traefik.http.services.minio-api.loadbalancer.server.port=9000"
      - "traefik.http.routers.minio-console.rule=Host(`minio.yourdomain.com`)" # <-- CHANGE
      - "traefik.http.routers.minio-console.entrypoints=websecure"
      - "traefik.http.routers.minio-console.tls.certresolver=le"
      - "traefik.http.services.minio-console.loadbalancer.server.port=9001"

  api:
    image: registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
    env_file: .env
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`api.yourdomain.com`)"             # <-- CHANGE
      - "traefik.http.routers.api.entrypoints=websecure"
      - "traefik.http.routers.api.tls.certresolver=le"
      - "traefik.http.services.api.loadbalancer.server.port=8080"

volumes:
  traefik_certs:
```

### DNS Records to Create

Add these **A records** in your DNS provider pointing to your Droplet's public IP:

| Hostname | Type | Value |
|---|---|---|
| `api.yourdomain.com` | A | `<Droplet IP>` |
| `s3.yourdomain.com` | A | `<Droplet IP>` |
| `minio.yourdomain.com` | A | `<Droplet IP>` |

> DNS propagation can take up to 24 hours. Traefik's Let's Encrypt TLS challenge will fail
> if DNS is not yet propagated. You can check with: `dig api.yourdomain.com`

---

## 7. Deploy the Stack

On the **server**, from `/opt/job-circuler`:

```bash
# Pull the latest API image from DigitalOcean registry
docker pull registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest

# Start the full production stack
make docker-prod-up
```

This runs internally:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

All services (Traefik, PostgreSQL, Redis, MinIO, API) will start in detached mode.

### Verify all containers are running

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
```

All services should show status `Up` or `running`.

---

## 8. Run Database Migrations

After the stack is up and PostgreSQL is healthy, run all migrations:

```bash
make migrate-up
```

This uses the `migrate/migrate` Docker image to apply all SQL files from
`internal/database/migrations/` (001 → 007) against the production database.

### Verify migrations were applied

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec postgres \
  psql -U bduser -d bdgovtjobs -c "\dt"
```

Expected tables: `users`, `categories`, `organizations`, `circulars`,
`bookmarks`, `alerts`, `scrape_logs`.

---

## 9. Create the MinIO Bucket

The storage bucket must be created manually after the first deployment:

```bash
# Source the env vars so the shell can use them
export $(grep -v '^#' .env | xargs)

# Create the bucket using MinIO CLI inside the container
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec minio \
  mc alias set local http://localhost:9000 ${MINIO_USER} ${MINIO_PASSWORD}

docker compose -f docker-compose.yml -f docker-compose.prod.yml exec minio \
  mc mb local/bd-govt-jobs-assets
```

Alternatively, visit `https://minio.yourdomain.com`, log in with your `MINIO_USER` /
`MINIO_PASSWORD`, and create the bucket `bd-govt-jobs-assets` via the UI.

---

## 10. Verify the Deployment

```bash
# Health check
curl https://api.yourdomain.com/api/v1/health
# Expected: {"success":true,"data":{"status":"ok","db":"ok","redis":"ok"}}

# Root endpoint
curl https://api.yourdomain.com/api/v1
# Expected: {"success":true,"data":{"message":"BD Govt Job Circular API v1.0.0"}}
```

---

## 11. Ongoing Deployments (CI/CD)

### Deploy a new version

**On your local machine** — build and push the updated image:

```bash
docker build -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest .
docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

**On the server:**

```bash
cd /opt/job-circuler
git pull          # pull any updated compose files or .env.example
make docker-prod-deploy
```

`docker-prod-deploy` runs:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml pull api
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d api
```

Only the `api` container is restarted — Traefik, Postgres, Redis, and MinIO keep running.

### Stream live logs

```bash
make docker-prod-logs
```

---

## 12. Useful Makefile Commands

| Command | Description |
|---|---|
| `make docker-prod-up` | Start the full production stack |
| `make docker-prod-deploy` | Pull latest API image + restart API only |
| `make docker-prod-logs` | Stream live API logs |
| `make docker-prod-down` | Stop and remove all production containers |
| `make migrate-up` | Apply all pending SQL migrations |
| `make migrate-down` | Roll back the last SQL migration |

---

## 13. Troubleshooting

### Traefik not issuing TLS certificate

- Confirm DNS is propagated: `dig api.yourdomain.com`
- Confirm ports 80 and 443 are open in the Droplet Cloud Firewall (DigitalOcean control panel → Networking → Firewalls)
- Check Traefik logs: `docker logs $(docker ps -qf name=traefik)`

### API container exits immediately after start

```bash
make docker-prod-logs
```

Common causes:
- Missing or invalid `.env` values (`JWT_SECRET`, `DB_PASSWORD`, `REDIS_URL`)
- `DB_HOST` is `localhost` instead of `postgres`
- `REDIS_URL` format is wrong — must be `redis://:<password>@redis:6379/0`
- `MINIO_ENDPOINT` is `localhost` instead of `http://minio:9000`

### Database connection refused

- Ensure `DB_HOST=postgres` in production `.env` (the Docker service name, not `localhost`)
- Check that Postgres passed its healthcheck:
  `docker compose -f docker-compose.yml -f docker-compose.prod.yml ps`

### MinIO access denied

- Confirm `MINIO_ACCESS_KEY` matches `MINIO_USER` and `MINIO_SECRET_KEY` matches `MINIO_PASSWORD`
- Confirm `MINIO_ENDPOINT=http://minio:9000` in `.env`

### Image pull fails on server

```bash
doctl registry login
docker pull registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

Ensure `doctl` is authenticated and that your Droplet has outbound internet access to the registry.

### Complete reset (data will be lost)

```bash
make docker-prod-down
docker volume prune -f   # WARNING: destroys Postgres and MinIO data
make docker-prod-up
make migrate-up
# Re-create the MinIO bucket (see Section 9)
```

---

## Deployment Checklist

- [ ] DigitalOcean Container Registry created and `doctl` authenticated locally
- [ ] Docker image built and pushed: `registry.digitalocean.com/<DO_REGISTRY>/bdgovtjobs-api:latest`
- [ ] Droplet has Docker, Docker Compose v2 plugin, Make, and `doctl` installed
- [ ] Repository cloned to `/opt/job-circuler` on the server
- [ ] Production `.env` created — all placeholders replaced with real values
- [ ] `APP_ENV=production` set in `.env`
- [ ] `DB_HOST=postgres` and `MINIO_ENDPOINT=http://minio:9000` in `.env`
- [ ] `REDIS_URL` uses `redis` hostname, not `localhost`
- [ ] `docker-compose.prod.yml` updated — all `yourdomain.com` replaced with real domains
- [ ] DNS A records created for `api.*`, `s3.*`, `minio.*` subdomains pointing to Droplet IP
- [ ] DNS propagated (verify with `dig`)
- [ ] Ports 80 and 443 open in DigitalOcean Cloud Firewall
- [ ] Stack started: `make docker-prod-up`
- [ ] All containers running: `docker compose -f docker-compose.yml -f docker-compose.prod.yml ps`
- [ ] Migrations applied: `make migrate-up`
- [ ] MinIO bucket `bd-govt-jobs-assets` created
- [ ] Health check passes: `curl https://api.yourdomain.com/api/v1/health`
