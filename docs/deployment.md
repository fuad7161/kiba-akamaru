# BD Govt Job Circular API — Production Deployment Guide

> **Stack:** Go · PostgreSQL · Redis · MinIO · Docker Compose · DigitalOcean Container Registry
> **Target:** Single DigitalOcean Droplet (Ubuntu), IP-only (no domain)

All services — API, PostgreSQL, Redis, MinIO — run on the **same Droplet** via Docker Compose.
The API is accessible at `http://<your-droplet-ip>:8080`.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Step 1 — DigitalOcean Container Registry](#2-step-1--digitalocean-container-registry)
3. [Step 2 — Build & Push Docker Image (local machine)](#3-step-2--build--push-docker-image-local-machine)
4. [Step 3 — Server Setup (Droplet)](#4-step-3--server-setup-droplet)
5. [Step 4 — Configure Production `.env`](#5-step-4--configure-production-env)
6. [Step 5 — Deploy the Stack](#6-step-5--deploy-the-stack)
7. [Step 6 — Run Database Migrations](#7-step-6--run-database-migrations)
8. [Step 7 — Create MinIO Bucket](#8-step-7--create-minio-bucket)
9. [Step 8 — Verify the Deployment](#9-step-8--verify-the-deployment)
10. [Updating the App (Re-deploy)](#10-updating-the-app-re-deploy)
11. [Useful Commands](#11-useful-commands)
12. [Troubleshooting](#12-troubleshooting)

---

## 1. Prerequisites

### Local machine
- Docker installed
- `doctl` (DigitalOcean CLI) installed and authenticated
- A DigitalOcean account with a **Container Registry** created

### DigitalOcean Droplet
- Ubuntu 22.04 LTS
- Minimum **2 vCPU / 2 GB RAM** (4 GB recommended)
- Port **8080** open in the Droplet firewall (for API access)
- Port **9001** open if you want to access the MinIO web console remotely

> Open ports via: DigitalOcean Control Panel → Networking → Firewalls → Inbound Rules

---

## 2. Step 1 — DigitalOcean Container Registry

This is where your compiled Docker image will be stored and pulled from on the server.

### On your local machine

```bash
# Install doctl
sudo snap install doctl

# Authenticate doctl with your DigitalOcean API token
doctl auth init

# Log Docker into the DigitalOcean registry
doctl registry login
```

Your registry name can be found in the DigitalOcean dashboard under **Container Registry**.
It will be used as the `DO_REGISTRY` value in your `.env`.

```
registry.digitalocean.com/<your-registry-name>
```

---

## 3. Step 2 — Build & Push Docker Image (local machine)

Run from the project root on your **local machine**:

```bash
export DO_REGISTRY=your-registry-name   # replace with your actual registry name

# Build the image
docker build -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest .

# Push to DigitalOcean registry
docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

> You must do this before deploying — the server will pull this image.

---

## 4. Step 3 — Server Setup (Droplet)

SSH into your Droplet and run the following once:

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker
sudo apt install -y docker.io

# Install Docker Compose v2 plugin
sudo apt install -y docker-compose-plugin

# Install Make
sudo apt install -y make

# Allow your user to run Docker without sudo
sudo usermod -aG docker $USER

newgrp docker

# Install doctl so the server can pull images from the registry
sudo snap install doctl
doctl auth init        # enter your DigitalOcean API token
doctl registry login   # authenticate Docker on the server to your registry
```

### Clone the project onto the server

```bash
git clone https://github.com/fuad71/job-circular-api.git /opt/job-circuler
cd /opt/job-circuler
```

---

## 5. Step 4 — Configure Production `.env`

On the **server**, inside `/opt/job-circuler`:

```bash
cp .env.example .env
nano .env
```

Fill in every value. The complete production `.env` should look like this:

```env
# Server
APP_PORT=8080
APP_ENV=production

# PostgreSQL
# DB_HOST must be "postgres" — the Docker service name, not localhost
DB_HOST=postgres
DB_PORT=5432
DB_NAME=bdgovtjobs
DB_USER=bduser
DB_PASSWORD=your_strong_db_password
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis
# REDIS_URL must use "redis" hostname — the Docker service name, not localhost
REDIS_PASSWORD=your_strong_redis_password
REDIS_URL=redis://:your_strong_redis_password@redis:6379/0

# JWT — generate with: openssl rand -hex 32
JWT_SECRET=your_64char_hex_secret
JWT_REFRESH_SECRET=another_64char_hex_secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASS=your_gmail_app_password
FROM_EMAIL=noreply@youremail.com

# Frontend URL (used for CORS and email links)
# Use your Droplet IP since there is no domain yet
FRONTEND_URL=http://<your-droplet-ip>:3000

# MinIO
# MINIO_ENDPOINT must use "minio" hostname — the Docker service name, not localhost
MINIO_USER=your_minio_admin_username
MINIO_PASSWORD=your_strong_minio_password
MINIO_BUCKET=bd-govt-jobs-assets
MINIO_REGION=us-east-1
MINIO_ACCESS_KEY=your_minio_admin_username
MINIO_SECRET_KEY=your_strong_minio_password
MINIO_ENDPOINT=http://minio:9000

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# DigitalOcean Registry
DO_REGISTRY=your-registry-name
```

### Generate secure random values

```bash
openssl rand -hex 32                                 # for JWT_SECRET / JWT_REFRESH_SECRET
openssl rand -base64 24 | tr -d '+/=' | cut -c1-24  # for DB / Redis / MinIO passwords
```

> **Note:** `DB_HOST`, `REDIS_URL`, and `MINIO_ENDPOINT` use Docker service names (`postgres`, `redis`, `minio`),
> not `localhost`. This is required because each service runs in its own container on the shared Docker network,
> even though they are all on the same physical machine.

> **Never commit `.env` to Git** — it is already in `.gitignore`.

---

## 6. Step 5 — Deploy the Stack

On the **server**, from `/opt/job-circuler`:

```bash
make docker-prod-up
```

This runs:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

It starts all four services in the background:
- `postgres` — database
- `redis` — cache / token store
- `minio` — file storage
- `api` — your Go backend (pulls image from DigitalOcean registry)

### Confirm all containers are running

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
```

All four should show `running`. Wait ~15 seconds for Postgres to become healthy before proceeding.

---

## 7. Step 6 — Run Database Migrations

```bash
make migrate-up
```

This applies all 7 SQL migration files from `internal/database/migrations/` to the PostgreSQL database.

### Confirm tables were created

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec postgres \
  psql -U bduser -d bdgovtjobs -c "\dt"
```

Expected tables: `users`, `categories`, `organizations`, `circulars`, `bookmarks`, `alerts`, `scrape_logs`

---

## 8. Step 7 — Create MinIO Bucket

The bucket must be created once after the first deployment.

### Option A — via CLI (recommended)

```bash
# Load env vars into your shell
export $(grep -v '^#' .env | xargs)

# Set up alias inside the minio container and create the bucket
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec minio \
  mc alias set local http://localhost:9000 ${MINIO_USER} ${MINIO_PASSWORD}

docker compose -f docker-compose.yml -f docker-compose.prod.yml exec minio \
  mc mb local/bd-govt-jobs-assets
```

### Option B — via web UI

Open `http://<your-droplet-ip>:9001` in a browser, log in with your `MINIO_USER` and `MINIO_PASSWORD`,
then create a bucket named `bd-govt-jobs-assets`.

---

## 9. Step 8 — Verify the Deployment

```bash
# Health check — confirms API, DB, and Redis are all connected
curl http://<your-droplet-ip>:8080/api/v1/health
# Expected: {"success":true,"data":{"status":"ok","db":"ok","redis":"ok"}}

# Root endpoint
curl http://<your-droplet-ip>:8080/api/v1
# Expected: {"success":true,"data":{"message":"BD Govt Job Circular API v1.0.0"}}
```

---

## 10. Updating the App (Re-deploy)

When you push new code:

### On your local machine
```bash
export DO_REGISTRY=your-registry-name

docker build -t registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest .
docker push registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

### On the server
```bash
cd /opt/job-circuler
git pull                   # pull any compose file or config changes
make docker-prod-deploy    # pulls new image, restarts only the api container
```

Only the `api` container restarts. PostgreSQL, Redis, and MinIO keep running uninterrupted.

---

## 11. Useful Commands

| Command | What it does |
|---|---|
| `make docker-prod-up` | Start the full production stack |
| `make docker-prod-deploy` | Pull new API image + restart API only |
| `make docker-prod-logs` | Stream live API logs |
| `make docker-prod-down` | Stop and remove all containers |
| `make migrate-up` | Apply all pending SQL migrations |
| `make migrate-down` | Roll back the last SQL migration |

---

## 12. Troubleshooting

### API container exits immediately

```bash
make docker-prod-logs
```

Common causes:
| Problem | Fix |
|---|---|
| `JWT_SECRET` is empty | Generate with `openssl rand -hex 32` and add to `.env` |
| `DB_HOST=localhost` | Must be `DB_HOST=postgres` |
| `REDIS_URL` uses localhost | Must be `redis://:password@redis:6379/0` |
| `MINIO_ENDPOINT` uses localhost | Must be `http://minio:9000` |
| Image not found | Run `doctl registry login` then `docker pull ...` manually |

### Database connection refused

- Wait ~15 seconds for Postgres to finish its healthcheck after `docker-prod-up`
- Check status: `docker compose -f docker-compose.yml -f docker-compose.prod.yml ps`
- Check Postgres logs: `docker compose -f docker-compose.yml -f docker-compose.prod.yml logs postgres`

### MinIO access denied

- Confirm `MINIO_ACCESS_KEY` equals `MINIO_USER` and `MINIO_SECRET_KEY` equals `MINIO_PASSWORD` in `.env`
- Confirm `MINIO_ENDPOINT=http://minio:9000`

### Can't pull image on server

```bash
doctl auth init       # re-authenticate
doctl registry login  # re-authenticate Docker
docker pull registry.digitalocean.com/${DO_REGISTRY}/bdgovtjobs-api:latest
```

### Full reset (⚠️ deletes all data)

```bash
make docker-prod-down
docker volume prune -f    # destroys Postgres data and MinIO files
make docker-prod-up
make migrate-up
# Re-create MinIO bucket (see Step 7)
```

---

## Deployment Checklist

- [ ] DigitalOcean Container Registry exists and `doctl` is authenticated on local machine
- [ ] Docker image built and pushed from local machine
- [ ] Droplet has Docker, Docker Compose plugin, Make, and `doctl` installed
- [ ] `doctl registry login` run on the Droplet
- [ ] Repo cloned to `/opt/job-circuler` on the Droplet
- [ ] `.env` created with all real values — no placeholders remaining
- [ ] `APP_ENV=production` in `.env`
- [ ] `DB_HOST=postgres` in `.env`
- [ ] `REDIS_URL` uses `redis` hostname in `.env`
- [ ] `MINIO_ENDPOINT=http://minio:9000` in `.env`
- [ ] Port `8080` open in Droplet firewall
- [ ] `make docker-prod-up` — all 4 containers running
- [ ] `make migrate-up` — all 7 migrations applied
- [ ] MinIO bucket `bd-govt-jobs-assets` created
- [ ] `curl http://<ip>:8080/api/v1/health` returns `{"status":"ok","db":"ok","redis":"ok"}`
