# BD Govt Job Circular API — Progress Checklist & TODO

> **Stack:** Go · PostgreSQL · Docker  
> **Last checked:** July 2026

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
- [x] **All 7 SQL migrations** (up + down)

### ⚙️ Configuration
- [x] `internal/config/config.go` — env var loading + app config struct

### 🌐 Entry Point
- [x] `cmd/server/main.go` — server bootstrap + all routes wired + CORS

### 📦 Shared Utilities
- [x] `pkg/response/response.go` — standard JSON success/error/paginated response helpers

### 🔐 Auth (complete)
- [x] `internal/model/user.go` — User + UserProfile structs
- [x] `internal/repository/user_repo.go` — Create, GetByEmail, GetByID, GetByVerifyToken, GetByResetToken, MarkVerified, SetResetToken, UpdatePassword, UpdateLastLogin, IsEmailTaken
- [x] `internal/service/auth_service.go` — Register, Login, VerifyEmail, ForgotPassword, ResetPassword, RefreshToken, Logout, GetProfile, JWT + bcrypt helpers
- [x] `internal/middleware/auth.go` — AuthRequired JWT middleware
- [x] `internal/handler/auth_handler.go` — All 8 auth endpoints

### 🏗️ Models (complete)
- [x] `model/user.go`
- [x] `model/circular.go`
- [x] `model/category.go`
- [x] `model/bookmark.go` (includes Alert + ScrapeLog)

### 🗃️ Repositories (complete)
- [x] `repository/user_repo.go`
- [x] `repository/circular_repo.go` — List, GetFeatured, GetByID, Create, Update, Delete, ToggleFeatured, ListCategories, ListOrganizations, GetStats, ListUsers, ListScrapeLogs
- [x] `repository/bookmark_repo.go` — List, Add, Remove + AlertRepo (List, Create, Delete, Toggle)

### 🔧 Services (partial)
- [x] `service/auth_service.go` — JWT generation, bcrypt hash/compare, email verification tokens
- [ ] `service/circular_service.go` — (business logic in repository for now)
- [ ] `service/email_service.go` — SMTP email sending
- [ ] `service/scrape_service.go` — scrape orchestration

### 🧰 Handlers (complete)
- [x] `handler/auth.go` — All 8 auth endpoints
- [x] `handler/circular.go` — List (paginated+filtered), Detail, Featured, Admin CRUD, Toggle featured
- [x] `handler/user.go` — Profile GET/PUT, Bookmarks CRUD, Alerts CRUD + toggle
- [x] `handler/category.go` — List categories, List organizations (in circular_handler)
- [x] `handler/admin.go` — Stats dashboard, User list, Manual scrape trigger, Scrape logs

### 🛡️ Middleware (complete)
- [x] `middleware/auth.go` — JWT validation middleware (`AuthRequired`)
- [x] `middleware/admin.go` — Admin role guard (`AdminOnly`)
- [x] CORS — Inline in main.go

---

## ❌ Not Yet Implemented

### 🕷️ Scraper (`internal/scraper/`) — **EMPTY**
- [ ] `scraper/bdjobs_fetcher.go`
- [ ] `scraper/teletalk_scraper.go`
- [ ] `scraper/normalizer.go`
- [ ] `scraper/deduplicator.go`
- [ ] `scraper/scheduler.go`

### 🔔 Email Alert System
- [ ] Cron job to match new circulars against user alert rules
- [ ] Send digest emails

### 🧪 Testing
- [ ] Unit tests for auth service
- [ ] Integration tests

---

## 📊 API Summary

| Area | Endpoints | Status |
|---|---|---|
| Health | `GET /health` | ✅ |
| Auth | 8 (register, login, logout, verify, forgot, reset, refresh, me) | ✅ |
| Circulars (public) | 4 (list, featured, detail, search) | ✅ |
| Circulars (admin) | 4 (create, update, delete, toggle feature) | ✅ |
| Categories | `GET /categories` | ✅ |
| Organizations | `GET /organizations` | ✅ |
| Users | 2 (get/put profile) | ✅ |
| Bookmarks | 3 (list, add, remove) | ✅ |
| Alerts | 4 (list, create, delete, toggle) | ✅ |
| Admin | 4 (stats, users, scrape trigger, scrape logs) | ✅ |
| **Total** | **32 endpoints** | |
