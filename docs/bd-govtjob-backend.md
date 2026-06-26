# BD Govt Job Circular Portal — Backend Documentation

> **Stack:** Go (Golang) · PostgreSQL · Redis · Docker  
> **Repo:** `job-circular-api`  
> **Version:** 1.0.0

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Project Structure](#2-project-structure)
3. [Tech Stack](#3-tech-stack)
4. [Database Schema](#4-database-schema)
5. [API Endpoints](#5-api-endpoints)
6. [Authentication System](#6-authentication-system)
7. [Web Scraper / Crawler](#7-web-scraper--crawler)
8. [Configuration & Environment](#8-configuration--environment)
9. [Running the Project](#9-running-the-project)
10. [Deployment](#10-deployment)

---

## 1. Project Overview

The `job-circular-api` is a Go backend service that:

- **Aggregates** government job circulars from multiple Bangladeshi sources (BDJobs API, Teletalk, ministry sites)
- **Exposes** a REST API consumed by the `job-circular-client` (Next.js frontend — separate repo)
- **Manages** user authentication, bookmarks, and email alert preferences
- **Runs** scheduled scrape jobs to keep circular data fresh

The frontend (`job-circular-client`) is a **completely separate codebase** and communicates with this service exclusively via the REST API over HTTP.

---

## 2. Project Structure

```
job-circular-api/
├── cmd/
│   └── server/
│       └── main.go               # Entry point
├── internal/
│   ├── config/
│   │   └── config.go             # Load env vars, app config struct
│   ├── database/
│   │   ├── postgres.go           # DB connection pool (pgx)
│   │   ├── redis.go              # Redis client
│   │   └── migrations/           # SQL migration files
│   │       ├── 001_create_users.sql
│   │       ├── 002_create_categories.sql
│   │       ├── 003_create_organizations.sql
│   │       ├── 004_create_circulars.sql
│   │       ├── 005_create_bookmarks.sql
│   │       ├── 006_create_alerts.sql
│   │       └── 007_create_scrape_logs.sql
│   ├── handler/
│   │   ├── auth.go               # Register, login, logout, refresh
│   │   ├── circular.go           # CRUD + search + filter
│   │   ├── user.go               # Profile, bookmarks, alerts
│   │   ├── category.go           # List categories
│   │   ├── admin.go              # Admin panel endpoints
│   │   └── health.go             # Health check
│   ├── middleware/
│   │   ├── auth.go               # JWT validation middleware
│   │   ├── role.go               # Admin role guard
│   │   ├── ratelimit.go          # Redis-backed rate limiter
│   │   └── cors.go               # CORS headers
│   ├── model/
│   │   ├── user.go
│   │   ├── circular.go
│   │   ├── category.go
│   │   ├── organization.go
│   │   ├── bookmark.go
│   │   └── scrape_log.go
│   ├── repository/
│   │   ├── circular_repo.go      # DB queries for circulars
│   │   ├── user_repo.go
│   │   ├── bookmark_repo.go
│   │   └── alert_repo.go
│   ├── service/
│   │   ├── auth_service.go       # Business logic: JWT, bcrypt
│   │   ├── circular_service.go   # Filtering, pagination, upsert
│   │   ├── email_service.go      # Send verification / alert emails
│   │   └── scrape_service.go     # Orchestrate crawlers
│   └── scraper/
│       ├── scheduler.go          # Cron job scheduler
│       ├── bdjobs_fetcher.go     # BDJobs internal JSON API fetcher
│       ├── teletalk_scraper.go   # Teletalk HTML scraper
│       ├── normalizer.go         # Normalize raw data → model
│       └── deduplicator.go       # SHA-256 hash dedup logic
├── pkg/
│   └── response/
│       └── response.go           # Standard JSON response helpers
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── .env.example
└── go.mod
```

---

## 3. Tech Stack

| Layer | Technology | Package / Notes |
|---|---|---|
| Language | Go 1.22+ | — |
| HTTP Router | **Chi** | `github.com/go-chi/chi/v5` — lightweight, idiomatic |
| Database | **PostgreSQL 15** | `github.com/jackc/pgx/v5` — native Go PG driver |
| Migrations | **golang-migrate** | `github.com/golang-migrate/migrate/v4` |
| Redis | **Redis 7** | `github.com/redis/go-redis/v9` — cache + rate limit + session |
| Auth | **JWT** | `github.com/golang-jwt/jwt/v5` |
| Password | **bcrypt** | `golang.org/x/crypto/bcrypt` |
| Validation | **validator** | `github.com/go-playground/validator/v10` |
| Scheduler | **gocron** | `github.com/go-co-op/gocron/v2` — scrape job scheduling |
| HTTP Client | **resty** | `github.com/go-resty/resty/v2` — for BDJobs API calls |
| HTML Scraper | **colly** | `github.com/gocolly/colly/v2` — for Teletalk / HTML sources |
| Email | **gomail** | `gopkg.in/gomail.v2` — SMTP email sending |
| Config | **godotenv** | `github.com/joho/godotenv` |
| Logging | **zerolog** | `github.com/rs/zerolog` — structured JSON logs |
| Container | **Docker** | Multi-stage build |

---

## 4. Database Schema

### Extensions

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

---

### Table: `users`

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(120) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'user',  -- 'user' | 'admin'
    is_verified     BOOLEAN DEFAULT FALSE,
    verify_token    TEXT,
    reset_token     TEXT,
    reset_token_exp TIMESTAMPTZ,
    phone           VARCHAR(20),
    district        VARCHAR(60),
    education_level VARCHAR(60),
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

---

### Table: `categories`

```sql
CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    name_bn     VARCHAR(100),
    slug        VARCHAR(100) NOT NULL UNIQUE,
    icon        VARCHAR(50),
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO categories (name, name_bn, slug, icon, sort_order) VALUES
  ('BCS',        'বিসিএস',      'bcs',        '⭐', 1),
  ('Bank Jobs',  'ব্যাংক চাকরি', 'bank-jobs',  '🏦', 2),
  ('Defense',    'সেনাবাহিনী',  'defense',    '🛡', 3),
  ('Police',     'পুলিশ',        'police',     '👮', 4),
  ('Education',  'শিক্ষা',       'education',  '🎓', 5),
  ('Health',     'স্বাস্থ্য',    'health',     '⚕',  6),
  ('Ministry',   'মন্ত্রণালয়', 'ministry',   '🏛', 7),
  ('Engineering','প্রকৌশল',     'engineering','⚙',  8),
  ('Railway',    'রেলওয়ে',      'railway',    '🚆', 9),
  ('Others',     'অন্যান্য',    'others',     '📋', 10);
```

---

### Table: `organizations`

```sql
CREATE TABLE organizations (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(255) NOT NULL UNIQUE,
    name_bn        VARCHAR(255),
    type           VARCHAR(60),   -- 'ministry' | 'autonomous' | 'university' | 'defense' | 'bank'
    website        TEXT,
    logo_url       TEXT,
    apply_base_url TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_org_name ON organizations USING gin(name gin_trgm_ops);
```

---

### Table: `circulars` *(main table)*

```sql
CREATE TABLE circulars (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id          VARCHAR(100),
    source               VARCHAR(60) NOT NULL,   -- 'bdjobs' | 'teletalk' | 'manual'
    source_url           TEXT,

    -- Core
    title                VARCHAR(500) NOT NULL,
    title_bn             VARCHAR(500),
    organization_id      INTEGER REFERENCES organizations(id) ON DELETE SET NULL,
    organization_name    VARCHAR(255) NOT NULL,
    category_id          INTEGER REFERENCES categories(id) ON DELETE SET NULL,

    -- Job Details
    vacancy              INTEGER,
    job_type             VARCHAR(50) DEFAULT 'permanent',
    gender               VARCHAR(20),
    age_min              INTEGER,
    age_max              INTEGER,
    age_note             TEXT,
    education_level      VARCHAR(100),
    education_detail     TEXT,
    experience_years     INTEGER,
    experience_note      TEXT,

    -- Salary
    salary_min           DECIMAL(12,2),
    salary_max           DECIMAL(12,2),
    salary_grade         VARCHAR(20),
    salary_display       VARCHAR(200),

    -- Location
    location             VARCHAR(255) DEFAULT 'Bangladesh',
    district             VARCHAR(60),
    division             VARCHAR(60),

    -- Dates
    published_date       DATE NOT NULL,
    application_deadline DATE,
    exam_date            DATE,

    -- Application
    apply_url            TEXT,
    apply_via            VARCHAR(60),   -- 'teletalk' | 'online' | 'physical'
    teletalk_code        VARCHAR(50),

    -- Content
    description          TEXT,
    requirements         TEXT,
    circular_image_url   TEXT,
    circular_pdf_url     TEXT,

    -- Status
    status               VARCHAR(30) DEFAULT 'active',  -- active | expired | closed
    is_featured          BOOLEAN DEFAULT FALSE,
    is_verified          BOOLEAN DEFAULT FALSE,
    view_count           INTEGER DEFAULT 0,
    content_hash         VARCHAR(64),

    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_circulars_status    ON circulars(status);
CREATE INDEX idx_circulars_deadline  ON circulars(application_deadline);
CREATE INDEX idx_circulars_published ON circulars(published_date DESC);
CREATE INDEX idx_circulars_category  ON circulars(category_id);
CREATE INDEX idx_circulars_org       ON circulars(organization_id);
CREATE INDEX idx_circulars_hash      ON circulars(content_hash);
CREATE UNIQUE INDEX uq_circular_hash ON circulars(content_hash);

-- Full-text search
CREATE INDEX idx_circulars_fts ON circulars
  USING gin(to_tsvector('english',
    coalesce(title,'') || ' ' || coalesce(organization_name,'')));

CREATE INDEX idx_circulars_title_trgm ON circulars USING gin(title gin_trgm_ops);
```

---

### Table: `bookmarks`

```sql
CREATE TABLE bookmarks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    circular_id UUID NOT NULL REFERENCES circulars(id) ON DELETE CASCADE,
    note        TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, circular_id)
);

CREATE INDEX idx_bookmarks_user     ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_circular ON bookmarks(circular_id);
```

---

### Table: `alerts`

```sql
CREATE TABLE alerts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    keyword         VARCHAR(200),
    category_id     INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE SET NULL,
    education_level VARCHAR(100),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_alerts_user   ON alerts(user_id);
CREATE INDEX idx_alerts_active ON alerts(is_active);
```

---

### Table: `scrape_logs`

```sql
CREATE TABLE scrape_logs (
    id            SERIAL PRIMARY KEY,
    source        VARCHAR(60) NOT NULL,
    started_at    TIMESTAMPTZ DEFAULT NOW(),
    finished_at   TIMESTAMPTZ,
    status        VARCHAR(20) DEFAULT 'running',  -- running | success | failed
    total_fetched INTEGER DEFAULT 0,
    new_inserted  INTEGER DEFAULT 0,
    updated       INTEGER DEFAULT 0,
    skipped       INTEGER DEFAULT 0,
    error_message TEXT,
    meta          JSONB
);
```

---

### Auto-update `updated_at` Trigger

```sql
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp BEFORE UPDATE ON circulars
  FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
```

---

## 5. API Endpoints

Base URL: `http://localhost:8080/api/v1`

### Auth

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `POST` | `/auth/register` | Public | Register new user |
| `POST` | `/auth/login` | Public | Login, returns JWT |
| `POST` | `/auth/logout` | JWT | Invalidate refresh token |
| `GET` | `/auth/verify-email` | Public | `?token=` Email verification |
| `POST` | `/auth/forgot-password` | Public | Send reset email |
| `POST` | `/auth/reset-password` | Public | Reset with token |
| `POST` | `/auth/refresh` | Cookie | Issue new access token |
| `GET` | `/auth/me` | JWT | Get own profile |

---

### Circulars

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/circulars` | Public | List circulars (paginated + filtered) |
| `GET` | `/circulars/:id` | Public | Single circular detail |
| `GET` | `/circulars/search` | Public | `?q=` Full-text search |
| `GET` | `/circulars/featured` | Public | Featured circulars |
| `POST` | `/circulars` | Admin | Create circular manually |
| `PUT` | `/circulars/:id` | Admin | Update circular |
| `DELETE` | `/circulars/:id` | Admin | Delete circular |
| `PATCH` | `/circulars/:id/feature` | Admin | Toggle featured flag |

#### Query Parameters for `GET /circulars`

```
page          int     default: 1
limit         int     default: 20, max: 100
category      string  category slug (e.g. "bcs", "bank-jobs")
status        string  active | expired | all   (default: active)
search        string  keyword search
deadline_from date    YYYY-MM-DD
deadline_to   date    YYYY-MM-DD
education     string  ssc | hsc | degree | masters
gender        string  male | female | both
sort          string  published_desc | deadline_asc | views_desc
```

#### Response shape (`GET /circulars`)

```json
{
  "success": true,
  "data": {
    "circulars": [
      {
        "id": "uuid",
        "title": "Assistant Director",
        "organization_name": "Bangladesh Bank",
        "category": { "id": 2, "name": "Bank Jobs", "slug": "bank-jobs" },
        "vacancy": 50,
        "salary_display": "Tk. 35,500 (Grade-6)",
        "published_date": "2026-06-01",
        "application_deadline": "2026-06-30",
        "apply_via": "teletalk",
        "status": "active",
        "is_featured": false
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 345,
      "total_pages": 18
    }
  }
}
```

---

### Users

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/users/me` | JWT | Get profile |
| `PUT` | `/users/me` | JWT | Update profile |
| `GET` | `/users/me/bookmarks` | JWT | List bookmarks |
| `POST` | `/users/me/bookmarks/:id` | JWT | Add bookmark |
| `DELETE` | `/users/me/bookmarks/:id` | JWT | Remove bookmark |
| `GET` | `/users/me/alerts` | JWT | List alert rules |
| `POST` | `/users/me/alerts` | JWT | Create alert |
| `DELETE` | `/users/me/alerts/:id` | JWT | Delete alert |

---

### Categories & Organizations

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/categories` | Public | List all categories |
| `GET` | `/organizations` | Public | List all organizations |

---

### Admin

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/admin/stats` | Admin | Dashboard stats |
| `GET` | `/admin/users` | Admin | List all users |
| `POST` | `/admin/scrape/run` | Admin | Trigger manual scrape |
| `GET` | `/admin/scrape/logs` | Admin | View scrape history |

---

## 6. Authentication System

### JWT Flow

```
Register → bcrypt hash password → save user → send verify email
Login    → check password → issue access token (15m) + refresh token (7d)
Access   → Bearer token in Authorization header
Refresh  → POST /auth/refresh with refresh token cookie → new access token
Logout   → delete refresh token from Redis
```

### Go Implementation Sketch

```go
// internal/service/auth_service.go

package service

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateAccessToken(userID, role, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID, secret string) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

### Middleware

```go
// internal/middleware/auth.go

func AuthRequired(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                response.Error(w, http.StatusUnauthorized, "missing token")
                return
            }
            tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
            claims, err := parseToken(tokenStr, jwtSecret)
            if err != nil {
                response.Error(w, http.StatusUnauthorized, "invalid token")
                return
            }
            ctx := context.WithValue(r.Context(), "claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## 7. Web Scraper / Crawler

### Data Sources

| Source | Method | Notes |
|---|---|---|
| `gateway.bdjobs.com` | JSON API | Internal REST API, no key needed, returns clean JSON |
| `alljobs.teletalk.com.bd` | HTML scrape | Colly-based scraper |
| `bdgovtjobs.com` | HTML scrape | Backup aggregator source |
| Ministry `.gov.bd` sites | HTML scrape | Original circulars (may be PDF/image) |

### BDJobs API Fetcher (Go)

```go
// internal/scraper/bdjobs_fetcher.go

package scraper

import (
    "fmt"
    "github.com/go-resty/resty/v2"
)

const bdjobsBaseURL = "https://gateway.bdjobs.com/ActtivejobsTest/api/JobSubsystem"

type BDJobsFetcher struct {
    client *resty.Client
}

func NewBDJobsFetcher() *BDJobsFetcher {
    client := resty.New().
        SetHeader("User-Agent", "Mozilla/5.0 (compatible; BDGovtJobBot/1.0)").
        SetHeader("Accept", "application/json").
        SetRetryCount(3).
        SetRetryWaitTime(2 * time.Second)
    return &BDJobsFetcher{client: client}
}

func (f *BDJobsFetcher) FetchJobList(page, pageSize int) ([]RawJob, error) {
    var result JobListResponse
    _, err := f.client.R().
        SetQueryParams(map[string]string{
            "jobType":  "government",
            "page":     fmt.Sprintf("%d", page),
            "pageSize": fmt.Sprintf("%d", pageSize),
        }).
        SetResult(&result).
        Get(bdjobsBaseURL + "/jobList")
    if err != nil {
        return nil, err
    }
    return result.Data, nil
}

func (f *BDJobsFetcher) FetchJobDetail(jobID int) (*RawJobDetail, error) {
    var result RawJobDetail
    _, err := f.client.R().
        SetQueryParam("jobId", fmt.Sprintf("%d", jobID)).
        SetResult(&result).
        Get(bdjobsBaseURL + "/jobDetails")
    return &result, err
}
```

### Scheduler (gocron)

```go
// internal/scraper/scheduler.go

package scraper

import (
    "github.com/go-co-op/gocron/v2"
    "time"
)

func StartScheduler(svc *ScrapeService) {
    s, _ := gocron.NewScheduler()

    // BDJobs every 6 hours
    s.NewJob(gocron.DurationJob(6*time.Hour),
        gocron.NewTask(svc.RunBDJobsScrape),
    )

    // Teletalk every 12 hours
    s.NewJob(gocron.DurationJob(12*time.Hour),
        gocron.NewTask(svc.RunTeletalkScrape),
    )

    // Expire old circulars daily at 1am
    s.NewJob(gocron.CronJob("0 1 * * *", false),
        gocron.NewTask(svc.ExpireOldCirculars),
    )

    s.Start()
}
```

### Deduplication

```go
// internal/scraper/deduplicator.go

package scraper

import (
    "crypto/sha256"
    "fmt"
)

func GenerateHash(title, orgName, publishedDate string) string {
    raw := fmt.Sprintf("%s|%s|%s", title, orgName, publishedDate)
    hash := sha256.Sum256([]byte(raw))
    return fmt.Sprintf("%x", hash)
}
```

### Upsert Query

```sql
INSERT INTO circulars (
  external_id, source, source_url, title, organization_name,
  vacancy, salary_display, salary_min, salary_max,
  published_date, application_deadline, apply_url,
  description, content_hash, status
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
ON CONFLICT (content_hash) DO UPDATE SET
  vacancy              = EXCLUDED.vacancy,
  application_deadline = EXCLUDED.application_deadline,
  status               = EXCLUDED.status,
  updated_at           = NOW();
```

---

## 8. Configuration & Environment

```bash
# .env.example

# Server
APP_PORT=8080
APP_ENV=development          # development | production

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=bdgovtjobs
DB_USER=bduser
DB_PASSWORD=your_strong_password
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_SECRET=your_256bit_random_secret_here
JWT_REFRESH_SECRET=another_256bit_random_secret_here
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h         # 7 days

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASS=your_app_password
FROM_EMAIL=noreply@yourdomain.com

# Frontend (for CORS + email links)
FRONTEND_URL=http://localhost:3000

# File Storage (MinIO — S3-compatible, self-hosted)
MINIO_BUCKET=bd-govt-jobs-assets
MINIO_REGION=us-east-1
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_ENDPOINT=http://localhost:9000

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

---

## 9. Running the Project

### Prerequisites

- Go 1.22+
- Docker + Docker Compose
- Make

### Quick Start (Development)

```bash
# 1. Clone the repo
git clone https://github.com/your-org/job-circular-api.git
cd job-circular-api

# 2. Copy env file
cp .env.example .env
# Edit .env with your values

# 3. Start PostgreSQL + Redis
docker compose up -d postgres redis

# 4. Run DB migrations
make migrate-up

# 5. Run the server
make run

# Server starts at http://localhost:8080
```

### Makefile

```makefile
.PHONY: run build test migrate-up migrate-down docker-up

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go

test:
	go test ./... -v

migrate-up:
	migrate -path internal/database/migrations \
	        -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	migrate -path internal/database/migrations \
	        -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

docker-up:
	docker compose up --build

scrape:
	curl -X POST http://localhost:8080/api/v1/admin/scrape/run \
	  -H "Authorization: Bearer $(ADMIN_TOKEN)"
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: bdgovtjobs
      POSTGRES_USER: bduser
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pg_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bduser"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  api:
    build: .
    env_file: .env
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    restart: unless-stopped

volumes:
  pg_data:
```

### Dockerfile

```dockerfile
# Multi-stage build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/server ./cmd/server/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/server .
EXPOSE 8080
CMD ["./server"]
```

---

## 10. Deployment

### Production Server Setup (Ubuntu 22.04)

```bash
# Install Go
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz

# Install Docker
curl -fsSL https://get.docker.com | sh

# Clone and build
git clone https://github.com/your-org/job-circular-api.git
cd job-circular-api
docker compose -f docker-compose.prod.yml up -d

# Nginx reverse proxy (listens on 80/443, proxies to :8080)
# certbot for SSL
```

### CORS Configuration

The API allows requests from the `FRONTEND_URL` env variable only. Update this when deploying the `job-circular-client` frontend:

```bash
FRONTEND_URL=https://your-nextjs-domain.com
```

### Health Check

```
GET /health
→ 200 { "status": "ok", "db": "ok", "redis": "ok" }
```

---

*Documentation last updated: June 2026*
