# BD Govt Job Circular API

> **Stack:** Go · PostgreSQL · Redis · MinIO · Docker · Traefik  
> **Version:** 1.0.0

Backend REST API for aggregating Bangladeshi government job circulars. Handles user authentication, bookmarking, email alerts, and scheduled scraping.

---

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Next.js FE  │────▶│  Go API :8080│────▶│ PostgreSQL   │
│  (separate)  │     │  (chi)       │     │ Redis        │
└──────────────┘     └──────────────┘     │ MinIO        │
                                          └──────────────┘
```

| Service | Port | Auth |
|---|---|---|
| **Go API** | `8080` | JWT (access + refresh) |
| **PostgreSQL** | `5432` | user/pass |
| **Redis** | `6379` | password |
| **MinIO API** | `9000` | access key + secret |
| **MinIO Console** | `9001` | access key + secret |

---

## Quick Start

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- Make

### 1. Clone & configure

```bash
git clone https://github.com/fuad71/job-circular-api.git
cd job-circular-api

# Copy and edit the env file
cp .env.example .env
```

### 2. Generate JWT secrets

```bash
# Run this and paste the output into your .env
openssl rand -hex 32   # → JWT_SECRET
openssl rand -hex 32   # → JWT_REFRESH_SECRET
```

### 3. Start services, run migrations, start API

```bash
# Start Postgres, Redis, MinIO
docker compose up -d postgres redis minio

# Run migrations
make migrate-up

# Start the Go server
make run
```

Server starts at **http://localhost:8080**

### 4. Verify

```bash
curl http://localhost:8080/api/v1
# → {"success":true,"data":{"message":"BD Govt Job Circular API v1.0.0"}}

curl http://localhost:8080/api/v1/health
# → {"success":true,"data":{"status":"ok","db":"ok","redis":"ok"}}
```

---

## Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```env
# ── Server ──────────────────────────────────────────
APP_PORT=8080
APP_ENV=development             # development | production

# ── PostgreSQL ──────────────────────────────────────
DB_HOST=localhost
DB_PORT=5432
DB_NAME=bdgovtjobs
DB_USER=bduser
DB_PASSWORD=your_strong_password
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# ── Redis ───────────────────────────────────────────
REDIS_PASSWORD=your_strong_redis_password
REDIS_URL=redis://:your_strong_redis_password@localhost:6379/0

# ── JWT ─────────────────────────────────────────────
JWT_SECRET=run_openssl_rand_hex_32
JWT_REFRESH_SECRET=run_openssl_rand_hex_32_again
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h            # 7 days

# ── SMTP (email) ────────────────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASS="your app password"
FROM_EMAIL=noreply@yourdomain.com

# ── Frontend ────────────────────────────────────────
FRONTEND_URL=http://localhost:3000

# ── MinIO ───────────────────────────────────────────
MINIO_USER=minioadmin
MINIO_PASSWORD=minioadmin
MINIO_BUCKET=bd-govt-jobs-assets
MINIO_REGION=us-east-1
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_ENDPOINT=http://localhost:9000

# ── Rate Limiting ───────────────────────────────────
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# ── Production (DigitalOcean) ───────────────────────
DO_REGISTRY=your-registry-name
ACME_EMAIL=your@email.com
```

### Tip: Generate secure random values

```bash
openssl rand -hex 32                                  # JWT secrets
openssl rand -base64 24 | tr -d '+/=' | cut -c1-24   # DB/Redis passwords
```

---

## Run Locally (Go on host, services in Docker)

Docker Compose uses 3 files that merge automatically:

| File | Purpose |
|---|---|
| `docker-compose.yml` | Service definitions (no ports, no build) |
| `docker-compose.override.yml` | **Local dev** — exposes ports, builds from source |
| `docker-compose.prod.yml` | **Production** — Traefik, pulls image from registry |

### Option A: Everything in Docker

```bash
make docker-up
# same as: docker compose up --build
```

### Option B: Services in Docker, Go on host (faster dev loop)

```bash
# Start services only
docker compose up -d postgres redis minio

# Run migrations
make migrate-up

# Start Go server
make run
```

---

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

### Auth (implemented)

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `POST` | `/auth/register` | Public | Register new user |
| `POST` | `/auth/login` | Public | Login → access_token + refresh_token cookie |
| `GET` | `/auth/verify-email` | Public | `?token=` Verify email |
| `POST` | `/auth/forgot-password` | Public | Send reset email |
| `POST` | `/auth/reset-password` | Public | Reset with token |
| `POST` | `/auth/refresh` | Cookie | Issue new access token |
| `POST` | `/auth/logout` | JWT | Invalidate refresh token |
| `GET` | `/auth/me` | JWT | Get own profile |

### Quick test

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Use the returned access_token
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

Or import the Postman collection from `postman/auth_api.json`.

---

## Makefile Reference

```bash
# ── Development ──────────────────────
make run              # Start Go server locally
make build            # Build binary to bin/server
make test             # Run all tests
make docker-up        # Build & start all Docker services
make migrate-up       # Apply SQL migrations
make migrate-down     # Rollback SQL migrations

# ── Production ───────────────────────
make docker-prod-up     # Deploy full production stack
make docker-prod-deploy # Pull latest API image + restart
make docker-prod-logs   # Stream production API logs
make docker-prod-down   # Stop production stack
```

---

## Docker Files Explained

```
docker-compose.yml          # Base: service definitions (postgres, redis, minio, api)
docker-compose.override.yml # Local: adds ports, build context, .env file
docker-compose.prod.yml     # Production: adds Traefik, pulls image from registry
```

- **Local:** `docker compose up` automatically merges `docker-compose.yml` + `docker-compose.override.yml`
- **Production:** Use `-f docker-compose.yml -f docker-compose.prod.yml` (override is ignored)
- **Custom:** Add a `docker-compose.custom.yml` and pass all three with `-f`

---

## MinIO Setup

When using MinIO for the first time, create the bucket:

```bash
# Via CLI
docker compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker compose exec minio mc mb local/bd-govt-jobs-assets
```

Or visit the console at **http://localhost:9001**, login with `minioadmin` / `minioadmin`, and create the bucket via the UI.

---

## Production Deployment

### 1. Set up the server

```bash
# On your Ubuntu server
sudo apt update && sudo apt install -y docker.io docker-compose-v2 make
sudo usermod -aG docker $USER

# Clone the repo
git clone https://github.com/fuad71/job-circular-api.git /opt/job-circuler
cd /opt/job-circuler
```

### 2. Create production `.env`

```env
APP_ENV=production
DB_HOST=postgres
DB_PASSWORD=<strong_prod_password>
REDIS_PASSWORD=<strong_prod_password>
REDIS_URL=redis://:<strong_prod_password>@redis:6379/0
JWT_SECRET=<random_hex_64>
JWT_REFRESH_SECRET=<random_hex_64>
MINIO_USER=admin
MINIO_PASSWORD=<strong_minio_password>
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=<strong_minio_password>
MINIO_ENDPOINT=http://minio:9000
DO_REGISTRY=your-registry-name
ACME_EMAIL=admin@yourdomain.com
```

> Update `docker-compose.prod.yml` — replace `yourdomain.com` with your actual domain.

### 3. Deploy

```bash
make docker-prod-up
```

### 4. Deploy updates

```bash
# Push new image to registry, then:
make docker-prod-deploy
```

---

## Project Structure

```
job-circular-api/
├── cmd/server/main.go              # Entry point, router setup
├── internal/
│   ├── config/config.go            # Env loading, Config struct
│   ├── database/
│   │   ├── postgres.go             # pgx connection pool
│   │   ├── redis.go                # Redis client
│   │   └── migrations/             # SQL migration files (001-007)
│   ├── model/user.go               # User & UserProfile structs
│   ├── repository/user_repo.go     # User DB operations
│   ├── service/auth_service.go     # Auth business logic
│   ├── handler/auth_handler.go     # HTTP handlers
│   ├── middleware/auth.go          # JWT middleware
│   └── scraper/                    # (future) Scraping logic
├── pkg/response/response.go        # JSON response helpers
├── postman/auth_api.json           # Postman collection
├── docker-compose.yml              # Base service definitions
├── docker-compose.override.yml     # Local dev overrides
├── docker-compose.prod.yml         # Production (Traefik)
├── Dockerfile                      # Multi-stage Go build
├── Makefile                        # Task shortcuts
├── go.mod
└── go.sum
```

## Documentation

- [Full Backend Documentation](docs/bd-govtjob-backend.md) — schema, endpoints, auth, scrapers
- [Docker Operations Guide](docs/operate-docker.md) — commands for all services
- [Progress Checklist](docs/progress_checklist.md) — what's done vs. pending

## License

MIT
