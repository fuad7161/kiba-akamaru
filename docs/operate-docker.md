# Docker Operations Guide

## 4 Services

| Service    | Container Name           | Port(s)         | Purpose               |
|------------|--------------------------|-----------------|-----------------------|
| **API**    | `job-circuler-api-1`     | `8080`          | Go backend server     |
| **Postgres** | `job-circuler-postgres-1` | `5432`          | Database              |
| **Redis**  | `job-circuler-redis-1`   | `6379`          | Caching / sessions    |
| **MinIO**  | `job-circuler-minio-1`   | `9000`, `9001`  | File storage (API + Console) |

---

## Lifecycle

### Start everything
```bash
docker compose up -d
```

### Start with rebuild (after code changes)
```bash
docker compose up -d --build api
```

### Stop everything (preserves data)
```bash
docker compose stop
```

### Start again (from where you left off)
```bash
docker compose start
```

### Restart a single service
```bash
docker compose restart api
```

### Stop a single service
```bash
docker compose stop redis
```

### Remove containers (keeps volumes — data safe)
```bash
docker compose down
```

### Full cleanup (⚠️ deletes all data)
```bash
docker compose down -v
```

### Check status
```bash
docker compose ps
```

---

## Logs

### Watch live logs (follow)
```bash
docker compose logs -f
```

### Watch a specific service
```bash
docker compose logs -f api
docker compose logs -f postgres
docker compose logs -f redis
docker compose logs -f minio
```

### Last 50 lines
```bash
docker compose logs --tail=50 api
```

---

## PostgreSQL

### Open SQL shell
```bash
docker compose exec postgres psql -U bduser -d bdgovtjobs
```

### Run a single query without entering shell
```bash
docker compose exec -T postgres psql -U bduser -d bdgovtjobs -c "SELECT count(*) FROM users;"
```

### Run all migrations
```bash
cat internal/database/migrations/*.up.sql | docker compose exec -T postgres psql -U bduser -d bdgovtjobs
```

### Backup database
```bash
docker compose exec -T postgres pg_dump -U bduser bdgovtjobs > backup.sql
```

### Restore from backup
```bash
cat backup.sql | docker compose exec -T postgres psql -U bduser -d bdgovtjobs
```

### List tables
```sql
\dt
```

### Describe a table
```sql
\d users
```

### Common troubleshooting
```bash
# Check if postgres is accepting connections
docker compose exec -T postgres pg_isready -U bduser

# Check connection count
docker compose exec -T postgres psql -U bduser -d bdgovtjobs -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## Redis

### Open Redis CLI
```bash
docker compose exec redis redis-cli
```

### Ping test
```bash
docker compose exec redis redis-cli ping
```
Expected: `PONG`

### Key operations
```bash
# List all keys
docker compose exec redis redis-cli KEYS '*'

# Get a value
docker compose exec redis redis-cli GET mykey

# Check key expiry (seconds remaining)
docker compose exec redis redis-cli TTL mykey

# Total key count
docker compose exec redis redis-cli DBSIZE
```

### Monitor all commands in real-time
```bash
docker compose exec redis redis-cli MONITOR
```

### Server info
```bash
docker compose exec redis redis-cli INFO
```

### Flush all keys (clear cache)
```bash
docker compose exec redis redis-cli FLUSHALL
```

---

## MinIO

### Open MinIO Console (Web UI)
Visit **http://localhost:9001** — login with `minioadmin` / `minioadmin`

### Create a bucket (from terminal)
```bash
docker compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker compose exec minio mc mb local/bd-govt-jobs-assets
```

### List buckets
```bash
docker compose exec minio mc ls local
```

### List files in a bucket
```bash
docker compose exec minio mc ls local/bd-govt-jobs-assets
```

### Upload a test file
```bash
docker compose exec minio mc cp /path/to/file local/bd-govt-jobs-assets/
```

### Bucket info / disk usage
```bash
docker compose exec minio mc du local/bd-govt-jobs-assets
```

---

## API

### Health check
```bash
curl http://localhost:8080/api/v1/health
```

### Open a shell inside the container
```bash
docker compose exec api sh
```

### Check environment variables inside container
```bash
docker compose exec api sh -c "env | sort"
```

### View binary info
```bash
docker compose exec api ls -la ./server
```

---

## Quick Reference

```bash
# Full restart cycle after code changes
docker compose down api && docker compose up -d --build api

# Check everything is up
docker compose ps

# Watch all logs
docker compose logs -f

# Stop everything for the day
docker compose stop

# Come back tomorrow
docker compose start
```
