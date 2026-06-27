# BD Govt Job Circular API — Progress Checklist & TODO

> **Stack:** Go · PostgreSQL · Docker  
> **Last checked:** June 26, 2026

---

## ✅ Completed Work

### 📁 Project Scaffolding & Configuration
- [x] Full directory structure created (`cmd/`, `internal/`, `pkg/`)
- [x] `go.mod` and `go.sum` initialized with all required dependencies
- [x] `.env.example` with all required env variables documented
- [x] `.env` present (local dev config)
- [x] `.gitignore` present
- [x] `Dockerfile` — multi-stage build (Go builder → Alpine)
- [x] `docker-compose.yml` — services: `postgres`, `api` with healthchecks
- [x] `Makefile` — `run`, `build`, `test`, `migrate-up`, `migrate-down`, `docker-up`, `scrape` targets

### 🗄️ Database Layer
- [x] `internal/database/postgres.go` — pgx connection pool setup
- [x] **All 7 SQL migrations** (up + down):
  - [x] `001_create_users.up.sql` / `.down.sql`
  - [x] `002_create_categories.up.sql` / `.down.sql` *(with seed data for 10 categories)*
  - [x] `003_create_organizations.up.sql` / `.down.sql`
  - [x] `004_create_circulars.up.sql` / `.down.sql` *(with all indexes + FTS + trgm)*
  - [x] `005_create_bookmarks.up.sql` / `.down.sql`
  - [x] `006_create_alerts.up.sql` / `.down.sql`
  - [x] `007_create_scrape_logs.up.sql` / `.down.sql`

### ⚙️ Configuration
- [x] `internal/config/config.go` — env var loading + app config struct

### 🌐 Entry Point
- [x] `cmd/server/main.go` — server bootstrap (DB connect, auth routes wired, server start)

### 📦 Shared Utilities
- [x] `pkg/response/response.go` — standard JSON success/error response helpers

### 🔐 Auth (complete)
- [x] `internal/model/user.go` — User + UserProfile structs
- [x] `internal/repository/user_repo.go` — Create, GetByEmail, GetByID, GetByVerifyToken, GetByResetToken, MarkVerified, SetResetToken, UpdatePassword, UpdateLastLogin, IsEmailTaken
- [x] `internal/service/auth_service.go` — Register, Login, VerifyEmail, ForgotPassword, ResetPassword, RefreshToken, Logout, GetProfile, JWT + bcrypt helpers
- [x] `internal/middleware/auth.go` — AuthRequired JWT middleware
- [x] `internal/handler/auth_handler.go` — All 8 auth endpoints

---

## ❌ Not Yet Implemented

### 🏗️ Models (`internal/model/`) — partial
- [x] `model/user.go`
- [ ] `model/circular.go`
- [ ] `model/category.go`
- [ ] `model/organization.go`
- [ ] `model/bookmark.go`
- [ ] `model/scrape_log.go`

### 🗃️ Repositories (`internal/repository/`) — partial
- [x] `repository/user_repo.go`
- [ ] `repository/circular_repo.go` — DB queries for circulars (list, filter, search, upsert)
- [ ] `repository/bookmark_repo.go` — add/remove/list bookmarks
- [ ] `repository/alert_repo.go` — alert CRUD

### 🔧 Services (`internal/service/`) — partial
- [x] `service/auth_service.go` — JWT generation, bcrypt hash/compare, email verification tokens
- [ ] `service/circular_service.go` — filtering, pagination logic, upsert orchestration
- [ ] `service/email_service.go` — SMTP email sending (verification, alerts)
- [ ] `service/scrape_service.go` — scrape orchestration, `RunBDJobsScrape`, `RunTeletalkScrape`, `ExpireOldCirculars`

### 🧰 Handlers (`internal/handler/`) — partial
- [x] `handler/auth.go` — Register, Login, Logout, Refresh, Verify Email, Forgot/Reset Password, `/auth/me`
- [ ] `handler/health.go` — `GET /health` (inline in main.go, move later)
- [ ] `handler/circular.go` — List (paginated+filtered), Detail, Search, Featured, Admin CRUD, Toggle featured
- [ ] `handler/user.go` — Profile GET/PUT, Bookmarks CRUD, Alerts CRUD
- [ ] `handler/category.go` — List categories, List organizations
- [ ] `handler/admin.go` — Stats dashboard, User list, Manual scrape trigger, Scrape logs

### 🛡️ Middleware (`internal/middleware/`) — partial
- [x] `middleware/auth.go` — JWT validation middleware (`AuthRequired`)
- [ ] `middleware/role.go` — Admin role guard (`AdminOnly`)
- [ ] `middleware/ratelimit.go` — Rate limiter
- [ ] `middleware/cors.go` — CORS headers using `FRONTEND_URL`

### 🕷️ Scraper (`internal/scraper/`) — **EMPTY**
- [ ] `scraper/bdjobs_fetcher.go` — BDJobs internal JSON API fetcher (resty, retry logic)
- [ ] `scraper/teletalk_scraper.go` — Teletalk HTML scraper (colly)
- [ ] `scraper/normalizer.go` — Normalize raw scraped data → `Circular` model
- [ ] `scraper/deduplicator.go` — SHA-256 content hash deduplication
- [ ] `scraper/scheduler.go` — gocron scheduler (BDJobs 6h, Teletalk 12h, Expire 1am daily)

---

## 🔜 Future / Nice-to-Have Features

### 🔔 Email Alert System
- [ ] Cron job to match new circulars against user alert rules
- [ ] Send digest emails to subscribed users when matching circulars found
- [ ] Unsubscribe link in alert emails

### 🔍 Additional Data Sources
- [ ] `bdgovtjobs.com` HTML scraper (backup aggregator)
- [ ] Ministry `.gov.bd` sites scraper (may require PDF/image parsing)

### 🚀 Production & Deployment
- [ ] `docker-compose.prod.yml` — production compose file
- [ ] Nginx reverse proxy config
- [ ] SSL/TLS via Certbot
- [ ] CI/CD pipeline (GitHub Actions: lint → test → build → deploy)

### 🧪 Testing
- [ ] Unit tests for auth service (JWT, bcrypt)
- [ ] Unit tests for deduplicator hash logic
- [ ] Integration tests for circular repository (testcontainers-go)
- [ ] Handler-level HTTP tests

### 📊 Observability
- [ ] Prometheus metrics endpoint (`/metrics`)
- [ ] Structured logging improvements (request ID, latency tracing via zerolog)
- [ ] Admin dashboard stats query (total circulars, active/expired counts, users)

---
