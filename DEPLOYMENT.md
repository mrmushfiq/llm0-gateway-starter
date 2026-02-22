# Deployment Guide

Guide for deploying LLM0 Gateway Starter locally and with managed databases.

## Prerequisites

- PostgreSQL 15+ database
- Redis 7+ instance
- At least one LLM provider API key (OpenAI, Anthropic, or Gemini)

---

## Local Development (Recommended)

### Quick Start with Docker

```bash
# 1. Start PostgreSQL and Redis
docker-compose up -d

# 2. Copy and edit environment file
cp .env.example .env
# Edit .env and add your API keys

# 3. Run migrations
psql postgresql://gateway:gateway@localhost:5432/gateway -f migrations/001_initial_schema.sql

# 4. Start the gateway
go run cmd/gateway/main.go
```

**Test:** Visit `http://localhost:8080/health`

### Manual Setup (Without Docker)

If you have PostgreSQL and Redis installed locally:

```bash
# 1. Create database
createdb gateway

# 2. Run migrations
psql gateway -f migrations/001_initial_schema.sql

# 3. Create .env file
cp .env.example .env

# Edit .env:
DATABASE_URL=postgresql://localhost:5432/gateway?sslmode=disable
REDIS_URL=redis://localhost:6379
OPENAI_API_KEY=sk-...
# ... add other API keys

# 4. Start gateway
go run cmd/gateway/main.go
```

---

## Using Neon (Free PostgreSQL)

[Neon](https://neon.tech) offers serverless PostgreSQL with a generous free tier.

### 1. Create Neon Database

1. Sign up at [neon.tech](https://neon.tech)
2. Click "Create Project"
3. Choose a name and region
4. Copy the connection string

### 2. Run Migrations

```bash
# Use the connection string from Neon
psql "postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/neondb?sslmode=require" \
  -f migrations/001_initial_schema.sql
```

### 3. Configure Gateway

Update `.env`:

```env
DATABASE_URL=postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/neondb?sslmode=require
REDIS_URL=redis://localhost:6379  # or use Upstash (below)
OPENAI_API_KEY=sk-...
```

### 4. (Optional) Use Upstash for Redis

For a fully serverless setup:

1. Sign up at [upstash.com](https://upstash.com)
2. Create Redis database
3. Copy connection string to `.env`

**Cost:** Both Neon and Upstash are **free** for small projects!

---

## Docker Deployment

### Build and Run

```bash
# Build image
docker build -t llm-gateway:latest .

# Run with Docker Compose (includes PostgreSQL and Redis)
docker-compose up -d

# Or run standalone (requires external DB + Redis)
docker run -d \
  --name llm-gateway \
  -p 8080:8080 \
  --env-file .env \
  llm-gateway:latest
```

---

## Environment Variables

### Required

```env
DATABASE_URL=postgresql://user:pass@host:port/db
REDIS_URL=redis://host:port
```

At least one provider API key:

```env
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...
```

### Optional

```env
PORT=8080
ENV=production
DEFAULT_RATE_LIMIT=100
CACHE_TTL_SECONDS=3600
CACHE_ENABLED=true
```

---

## Database Migrations

### How Migrations Run

**With Docker (Automatic):**
```bash
docker-compose up -d
# Migrations run automatically on first start via /docker-entrypoint-initdb.d
# Only runs once when postgres_data volume is created
```

**With setup.sh (Recommended):**
```bash
./scripts/setup.sh
# Automated script that:
# 1. Starts Docker containers
# 2. Waits for PostgreSQL to be ready
# 3. Runs migrations automatically
```

**Manual Migration (Any PostgreSQL):**
```bash
# For local Docker
psql postgresql://gateway:gateway@localhost:5432/gateway -f migrations/001_initial_schema.sql

# For Neon or remote database
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql

# Inside Docker container
docker exec -i gateway_postgres psql -U gateway -d gateway < migrations/001_initial_schema.sql
```

### Verify Migration

```bash
# Check tables exist
psql "$DATABASE_URL" -c "\dt"

# Check sample data (should return 12 rows for default model pricing)
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM model_pricing;"

# Check test API key exists
psql "$DATABASE_URL" -c "SELECT key_prefix, name FROM api_keys WHERE key_prefix = 'gw_test';"
```

### Making Schema Changes

For production schema changes, we recommend [**Atlas**](https://atlasgo.io):

```bash
# Install Atlas
brew install ariga/tap/atlas  # macOS
# or: curl -sSf https://atlasgo.sh | sh

# Generate migration from schema changes
atlas migrate diff \
  --to "file://migrations/001_initial_schema.sql" \
  --dev-url "docker://postgres/15/dev"

# Apply migrations
atlas schema apply \
  --url "$DATABASE_URL" \
  --to "file://migrations/001_initial_schema.sql"

# Inspect current schema
atlas schema inspect --url "$DATABASE_URL"
```

**Alternative:** Use [golang-migrate](https://github.com/golang-migrate/migrate) for version-based migrations:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path ./migrations -database "$DATABASE_URL" up
```

---

## Health Checks

### Endpoint

```
GET /health
```

Returns `200 OK` if healthy.

### Test

```bash
curl http://localhost:8080/health
```

---

## Monitoring & Analytics

### Check Request Logs

```sql
-- Daily cost breakdown
SELECT 
  DATE(created_at) as date,
  model,
  COUNT(*) as requests,
  SUM(cost_usd) as total_cost,
  AVG(latency_ms) as avg_latency
FROM gateway_logs
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at), model
ORDER BY date DESC, total_cost DESC;

-- Cache hit rate
SELECT 
  DATE(created_at) as date,
  COUNT(*) FILTER (WHERE cache_hit = true) * 100.0 / COUNT(*) as hit_rate
FROM gateway_logs
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Provider usage
SELECT 
  provider,
  COUNT(*) as requests,
  AVG(latency_ms) as avg_latency
FROM gateway_logs
WHERE created_at >= NOW() - INTERVAL '1 day'
GROUP BY provider;
```

---

## Troubleshooting

### Gateway won't start

**Check:** Environment variables

```bash
# Verify DATABASE_URL is accessible
psql "$DATABASE_URL" -c "SELECT 1"

# Verify REDIS_URL is accessible
redis-cli -u "$REDIS_URL" PING

# Check Go version
go version  # Need 1.21+
```

### High latency

**Possible causes:**
- Cache disabled (check `CACHE_ENABLED=true` in .env)
- Database connection pool exhausted
- Redis connection slow
- Provider API slow

### Rate limit issues

**Check Redis keys:**

```bash
redis-cli -u "$REDIS_URL" KEYS "ratelimit:*"
redis-cli -u "$REDIS_URL" GET "ratelimit:your-api-key-id"
```

### Cost calculations wrong

**Check model pricing:**

```sql
SELECT * FROM model_pricing WHERE model = 'gpt-4o-mini';
```

If missing, verify migration ran successfully.

---

## Security Checklist

Before deploying:

- [ ] Change test API key (`gw_test_abc123`)
- [ ] Use environment variables for secrets (not .env file)
- [ ] Enable HTTPS (reverse proxy like Nginx/Caddy)
- [ ] Set up firewall rules (only allow port 443/80)
- [ ] Use managed database with automated backups
- [ ] Review CORS settings in code
- [ ] Set up log retention policies
- [ ] Rotate API keys regularly

---

## Production Deployment

### Self-Hosted at Scale

This gateway is production-ready for small to medium-scale deployments. For large-scale production with:
- Semantic caching (3x better hit rates)
- Cost-based rate limiting
- Customer attribution & analytics
- Auto-scaling infrastructure
- Global edge deployment
- Managed operations

**Check out [LLM0.ai](https://llm0.ai) *(Coming Soon)*** - managed LLM gateway service.
