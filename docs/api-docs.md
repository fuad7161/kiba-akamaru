# API Documentation

> **Base URL:** `http://64.227.165.222:8080/api/v1`  
> **Format:** All requests and responses use `application/json`  
> **Auth:** Bearer token in `Authorization` header for JWT-protected endpoints

---

## Response Format

All responses follow a standard envelope:

```json
{
  "success": true,
  "data": { ... },    // present on success
  "error": "..."      // present on error
}
```

---

## Health & Info

### GET /health

Check if the server and database are running.

```bash
curl http://64.227.165.222:8080/health
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "db": "ok"
  }
}
```

---

### GET /api/v1

API version info.

```bash
curl http://64.227.165.222:8080/api/v1
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "BD Govt Job Circular API v1.0.0"
  }
}
```

---

## Auth

### POST /auth/register

Register a new user account.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123",
    "phone": "01712345678",
    "district": "Dhaka",
    "education_level": "Masters"
  }'
```

**Request body:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | User's full name |
| `email` | string | ✅ | Email address |
| `password` | string | ✅ | Min 6 characters |
| `phone` | string | ❌ | Phone number |
| `district` | string | ❌ | District name |
| `education_level` | string | ❌ | SSC, HSC, Degree, Masters, etc. |

**Success (201):**
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "message": "registration successful. check your email for verification"
  }
}
```

**Errors:**

| Status | Condition |
|---|---|
| `400` | Missing required fields or password < 6 chars |
| `409` | Email already registered |

> **Note:** Verification email is log-only. Check server output for the verification link.

---

### POST /auth/login

Authenticate and receive JWT tokens.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Request body:**

| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | ✅ | Registered email |
| `password` | string | ✅ | Account password |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Test User",
      "email": "test@example.com",
      "role": "user",
      "is_verified": false,
      "phone": "01712345678",
      "district": "Dhaka",
      "education_level": "Masters",
      "created_at": "2026-06-28T12:00:00Z"
    }
  }
}
```

Also sets a **`refresh_token` httpOnly cookie** (Path: `/api/v1/auth`, 7 days expiry).

| Token | TTL | Where |
|---|---|---|
| `access_token` | 15m (default) | Response body → store in memory |
| `refresh_token` | 7d (default) | httpOnly cookie → sent automatically |

**Errors:**

| Status | Condition |
|---|---|
| `401` | Invalid email or password |

---

### POST /auth/refresh

Get a new access token using the refresh token cookie.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/refresh \
  -b "refresh_token=eyJhbGciOiJIUzI1NiIs..."
```

| Auth | Source |
|---|---|
| None (cookie) | `refresh_token` cookie set by `/auth/login` |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "a1b2c3d4-...",
      "name": "Test User",
      "email": "test@example.com",
      "role": "user",
      "is_verified": true,
      "phone": "01712345678",
      "district": "Dhaka",
      "education_level": "Masters",
      "created_at": "2026-06-28T12:00:00Z"
    }
  }
}
```

Also sets a **new** `refresh_token` cookie (token rotation).

**Errors:**

| Status | Condition |
|---|---|
| `401` | Missing, invalid, or expired refresh token |

---

### POST /auth/logout

Invalidate the current refresh token.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

| Auth | Source |
|---|---|
| JWT | `Authorization: Bearer <access_token>` |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "message": "logged out"
  }
}
```

Also clears the `refresh_token` cookie.

---

### GET /auth/verify-email

Verify email address using the token from the verification email.

```bash
curl "http://64.227.165.222:8080/api/v1/auth/verify-email?token=abc123def456..."
```

**Query parameter:**

| Param | Required | Description |
|---|---|---|
| `token` | ✅ | 64-char hex token from verification email |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "message": "email verified successfully"
  }
}
```

**Errors:**

| Status | Condition |
|---|---|
| `400` | Missing or invalid/expired token |

---

### POST /auth/forgot-password

Request a password reset email.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

**Request body:**

| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | ✅ | Registered email address |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "message": "if an account with that email exists, a reset link has been sent"
  }
}
```

> Always returns 200 even if email doesn't exist (prevents email enumeration).

> **Note:** Reset email is log-only. Check server output for the reset link.

---

### POST /auth/reset-password

Reset password using the token from the reset email.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123def456...",
    "new_password": "newpassword456"
  }'
```

**Request body:**

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | ✅ | 64-char hex token from reset email |
| `new_password` | string | ✅ | Min 6 characters |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "message": "password reset successfully"
  }
}
```

**Errors:**

| Status | Condition |
|---|---|
| `400` | Missing fields, password < 6 chars, or invalid/expired token |

---

### GET /auth/me

Get the authenticated user's profile.

```bash
curl http://64.227.165.222:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

| Auth | Source |
|---|---|
| JWT | `Authorization: Bearer <access_token>` |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Test User",
    "email": "test@example.com",
    "role": "user",
    "is_verified": true,
    "phone": "01712345678",
    "district": "Dhaka",
    "education_level": "Masters",
    "created_at": "2026-06-28T12:00:00Z"
  }
}
```

**Errors:**

| Status | Condition |
|---|---|
| `401` | Missing or invalid access token |

---

## Circulars

### GET /circulars

List circulars with pagination and filters.

```bash
curl "http://64.227.165.222:8080/api/v1/circulars?page=1&limit=20&status=active&sort=published_desc"
```

**Query parameters:**

| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page (max 100) |
| `category` | string | — | Category slug (e.g. `bcs`, `bank-jobs`) |
| `status` | string | `active` | `active`, `expired`, or `all` |
| `search` | string | — | Search in title and organization name |
| `deadline_from` | string | — | YYYY-MM-DD |
| `deadline_to` | string | — | YYYY-MM-DD |
| `education` | string | — | SSC, HSC, Degree, Masters |
| `gender` | string | — | `male`, `female`, `both` |
| `sort` | string | `published_desc` | `published_desc`, `deadline_asc`, `views_desc` |

**Success (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "title": "Assistant Director",
        "organization_name": "Bangladesh Bank",
        "category": { "id": 2, "name": "Bank Jobs", "slug": "bank-jobs" },
        "vacancy": 50,
        "salary_display": "Tk. 35,500 (Grade-6)",
        "published_date": "2026-06-01T00:00:00Z",
        "application_deadline": "2026-06-30T00:00:00Z",
        "apply_via": "teletalk",
        "location": "Dhaka",
        "district": "Dhaka",
        "job_type": "permanent",
        "status": "active",
        "is_featured": false
      }
    ],
    "pagination": { "page": 1, "limit": 20, "total": 345, "total_pages": 18 }
  }
}
```

---

### GET /circulars/featured

Get featured circulars for the homepage.

```bash
curl http://64.227.165.222:8080/api/v1/circulars/featured
```

**Success (200):** Returns an array of `CircularListItem` objects.

---

### GET /circulars/:id

Get full detail of a single circular.

```bash
curl http://64.227.165.222:8080/api/v1/circulars/abc-123
```

**Success (200):** Returns the full `Circular` object with nested `category` and `organization`.

---

### POST /circulars — Admin

Create a new circular manually.

```bash
curl -X POST http://64.227.165.222:8080/api/v1/circulars \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Job","organization_name":"Test Org","published_date":"2026-07-01"}'
```

| Auth | Source |
|---|---|
| JWT (admin) | `Authorization: Bearer <access_token>` |

**Required:** `title`, `organization_name`, `published_date` (YYYY-MM-DD)

---

### PUT /circulars/:id — Admin

Update a circular.

```bash
curl -X PUT http://64.227.165.222:8080/api/v1/circulars/abc-123 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","status":"expired"}'
```

---

### DELETE /circulars/:id — Admin

Delete a circular.

```bash
curl -X DELETE http://64.227.165.222:8080/api/v1/circulars/abc-123 \
  -H "Authorization: Bearer <access_token>"
```

---

### PATCH /circulars/:id/feature — Admin

Toggle featured flag.

```bash
curl -X PATCH http://64.227.165.222:8080/api/v1/circulars/abc-123/feature \
  -H "Authorization: Bearer <access_token>"
```

**Success (200):** `{ "is_featured": true }`

---

## Categories & Organizations

### GET /categories

```bash
curl http://64.227.165.222:8080/api/v1/categories
```

**Success (200):** Returns array of categories (id, name, name_bn, slug, icon, sort_order).

---

### GET /organizations

```bash
curl http://64.227.165.222:8080/api/v1/organizations
```

**Success (200):** Returns array of organizations (id, name, name_bn, type, website, logo_url).

---

## Users

All require JWT.

### GET /users/me

### PUT /users/me

```bash
curl -X PUT http://64.227.165.222:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"phone":"01712345678","district":"Dhaka"}'
```

---

## Bookmarks

All require JWT.

### GET /users/me/bookmarks
### POST /users/me/bookmarks/:id
### DELETE /users/me/bookmarks/:id

---

## Alerts

All require JWT.

### GET /users/me/alerts
### POST /users/me/alerts

```bash
curl -X POST http://64.227.165.222:8080/api/v1/users/me/alerts \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"keyword":"bank","category_id":2}'
```

### DELETE /users/me/alerts/:id
### PATCH /users/me/alerts/:id/toggle

---

## Admin

All require JWT + admin role.

### GET /admin/stats

```bash
curl http://64.227.165.222:8080/api/v1/admin/stats \
  -H "Authorization: Bearer <admin_access_token>"
```

**Success (200):**
```json
{
  "success": true,
  "data": {
    "active_circulars": 345,
    "total_circulars": 1250,
    "total_users": 89
  }
}
```

### GET /admin/users
### POST /admin/scrape/run
### GET /admin/scrape/logs

---

## Error Reference

### Standard error response

```json
{
  "success": false,
  "error": "descriptive error message"
}
```

### Common HTTP status codes

| Code | Meaning |
|---|---|
| `200` | Success |
| `201` | Created (registration) |
| `400` | Bad request (missing/invalid fields) |
| `401` | Unauthorized (invalid credentials or expired token) |
| `404` | Not found |
| `409` | Conflict (email already registered) |
| `429` | Too many requests (rate limited) |
| `500` | Internal server error |

---

## Authentication Flow Summary

```
Register
  │
  ▼
┌──────────────────────────────────────────┐
│  POST /auth/register                     │
│  → 201 + verify email sent (log-only)    │
└──────────────────────────────────────────┘
  │
  ▼
┌──────────────────────────────────────────┐
│  GET /auth/verify-email?token=...         │
│  → 200  (manually get token from logs)   │
└──────────────────────────────────────────┘
  │
  ▼
┌──────────────────────────────────────────┐
│  POST /auth/login                        │
│  → { access_token } + refresh_token cookie│
└──────────────────────────────────────────┘
  │
  ├── Access resource: Authorization: Bearer <access_token>
  │
  ├── Token expired? → POST /auth/refresh (cookie sent automatically)
  │
  └── POST /auth/logout → invalidates refresh token
```
