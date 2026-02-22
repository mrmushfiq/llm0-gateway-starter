# Getting Started - LLM0 Gateway Starter

## Overview

**Open-source LLM gateway** with multi-provider support, automatic failover, streaming, and intelligent caching.

**Use this for:**
- 🏢 Self-hosted deployments
- 🚀 Small-scale production apps
- 🔧 Custom integrations
- 💼 Enterprise on-premise

---

## 5-Minute Setup

### Option A: Automated (Easiest)

```bash
# One command setup - handles everything
./scripts/setup.sh

# Edit .env and add your API key
# OPENAI_API_KEY=sk-...

# Start the gateway
go run cmd/gateway/main.go
```

### Option B: Manual Setup

**Step 1: Start Infrastructure**

```bash
# Start PostgreSQL and Redis with Docker
docker-compose up -d
# Migrations run automatically on first start via /docker-entrypoint-initdb.d
```

**Step 2: Setup Environment**

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add at least one API key:
# OPENAI_API_KEY=sk-...
# or ANTHROPIC_API_KEY=sk-ant-...
# or GEMINI_API_KEY=...
```

**Step 3: Verify Database (Optional)**

```bash
# Check that migrations ran successfully
psql postgresql://gateway:gateway@localhost:5432/gateway -c "\dt"
# Should show: api_keys, model_pricing, gateway_logs

# If migrations didn't run, run manually:
psql postgresql://gateway:gateway@localhost:5432/gateway -f migrations/001_initial_schema.sql
```

**Step 4: Start Gateway**

```bash
go run cmd/gateway/main.go
```

You should see:
```
✓ Connected to PostgreSQL
✓ Connected to Redis
✓ Initialized LLM providers
🚀 Server listening on http://localhost:8080
```

**Test API key created:** `gw_test_abc123`

---

## Your First Request

### Test with cURL

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Say hello in one word"}]
  }'
```

**Response:**
```json
{
  "id": "chatcmpl-...",
  "object": "chat.completion",
  "model": "gpt-4o-mini",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 1,
    "total_tokens": 13
  },
  "cost_usd": 0.000002,
  "latency_ms": 234
}
```

**Headers:**
```
X-Cache-Hit: miss
X-Cost-USD: 0.000002
X-Provider: openai
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
```

---

## Test Streaming

```bash
curl -N -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Count to 3"}],
    "stream": true
  }'
```

**Response (Server-Sent Events):**
```
data: {"id":"...","choices":[{"delta":{"role":"assistant"}}]}
data: {"id":"...","choices":[{"delta":{"content":"1"}}]}
data: {"id":"...","choices":[{"delta":{"content":","}}]}
data: {"id":"...","choices":[{"delta":{"content":" 2"}}]}
data: {"id":"...","choices":[{"delta":{"content":","}}]}
data: {"id":"...","choices":[{"delta":{"content":" 3"}}]}
data: [DONE]
```

---

## Test Caching

```bash
# First request (cache miss)
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "What is 2+2?"}]
  }' | jq '.cost_usd, .latency_ms'

# Output: 0.000003, 450

# Second identical request (cache hit)
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "What is 2+2?"}]
  }' | jq '.cost_usd, .latency_ms'

# Output: 0.000000, 2  (instant, free!)
```

**Cache hit:** ~2ms latency, $0 cost

---

## Test Failover

```bash
# Request with a model that will fail (if OpenAI is down)
# Gateway automatically tries Anthropic, then Gemini
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}]
  }' -D - | grep "X-Failover"
```

If failover triggered, you'll see:
```
X-Failover: true
X-Provider: anthropic
```

---

## Using with OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="gw_test_abc123"  # Your gateway API key
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

Works with **any OpenAI-compatible SDK** (LangChain, LlamaIndex, etc.)!

---

## Check Analytics

```bash
# Connect to database
psql postgresql://gateway:gateway@localhost:5432/gateway

# View request logs
SELECT model, COUNT(*), SUM(cost_usd), AVG(latency_ms) 
FROM gateway_logs 
GROUP BY model;

# Check cache hit rate
SELECT 
  COUNT(*) FILTER (WHERE cache_hit = true) * 100.0 / COUNT(*) as hit_rate
FROM gateway_logs;
```

---

## Next Steps

### 1. Try Different Providers

```bash
# Anthropic
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -d '{"model": "claude-haiku-4-5-20251001", ...}'

# Gemini
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -d '{"model": "gemini-2.5-flash", ...}'
```

### 2. Create Your Own API Key

```sql
-- Connect to database
psql postgresql://gateway:gateway@localhost:5432/gateway

-- Create new API key
INSERT INTO api_keys (key_hash, key_prefix, name, rate_limit_per_minute, cache_enabled)
VALUES (
  encode(digest('my_secret_key_abc123', 'sha256'), 'hex'),
  'my_secret_key',
  'My API Key',
  100,
  true
);
```

Now use `my_secret_key_abc123` as your API key.

### 3. Customize Failover Chains

Edit `internal/gateway/providers/manager.go`:

```go
func (m *Manager) setupFailoverChains() {
    // Add your own failover chains
    m.failover["gpt-4o"] = []string{"claude-sonnet-4-5-20250929", "gemini-2.5-pro"}
    // ...
}
```

### 4. Adjust Rate Limits

Update in database:

```sql
UPDATE api_keys 
SET rate_limit_per_minute = 200 
WHERE key_prefix = 'gw_test_abc';
```

### 5. Deploy to Production

See [DEPLOYMENT.md](DEPLOYMENT.md) for deployment guides.

---

## Troubleshooting

### "Invalid API key"

Make sure you're using `gw_test_abc123` (created by migration) or create your own.

### "Provider not configured"

Check that you've added the API key for the provider in `.env`:
- `OPENAI_API_KEY` for GPT models
- `ANTHROPIC_API_KEY` for Claude models  
- `GEMINI_API_KEY` for Gemini models

### "Database error"

Verify migrations ran:
```bash
psql "$DATABASE_URL" -c "\dt"
```

Should show: `api_keys`, `model_pricing`, `gateway_logs`

### "Redis connection failed"

Check Redis is running:
```bash
redis-cli -u "$REDIS_URL" PING
```

Should return: `PONG`

### "Need to run migrations again"

```bash
# Migrations are idempotent (safe to run multiple times)
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
```

---

## Database Schema & Migrations

### How Migrations Work

**With Docker:**
- Migrations run **automatically on first start** via `/docker-entrypoint-initdb.d`
- Only runs once when the `postgres_data` volume is created
- If you need to re-run: `docker-compose down -v` (deletes data!) then `docker-compose up -d`

**With setup.sh:**
- Script automatically runs migrations after starting Docker

**Manually:**
```bash
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
```

### Making Schema Changes

If you need to modify the database schema, we recommend [**Atlas**](https://atlasgo.io):

```bash
# Install on MacOS
brew install ariga/tap/atlas

# Edit migrations/001_initial_schema.sql with your changes

# Generate migration
atlas migrate diff \
  --to "file://migrations/001_initial_schema.sql" \
  --dev-url "docker://postgres/15/dev"

# Apply changes
atlas schema apply \
  --url "$DATABASE_URL" \
  --to "file://migrations/001_initial_schema.sql"
```

**Why Atlas?**
- Declarative schema management
- Automatic migration generation
- Free for open source projects

**Alternative:** [golang-migrate](https://github.com/golang-migrate/migrate) for version-based migrations.

---

## Documentation

- [README.md](README.md) - Main documentation & features
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical design
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment options
- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute

### Key Files

1. `cmd/gateway/main.go` - HTTP server entry point
2. `internal/gateway/handlers/chat.go` - Request handler with caching & failover
3. `internal/gateway/providers/manager.go` - Multi-provider abstraction & failover logic
4. `internal/gateway/cache/cache.go` - Redis caching layer

---

## Feature Comparison

### This Project (Open Source)
- ✅ Exact-match caching (12-15% hit rate)
- ✅ Token bucket rate limiting
- ✅ Per-request cost tracking
- ✅ Multi-provider failover
- ✅ Full source code access
- 🚧 **Self-hosted models (vLLM + K8s)** — Coming Soon
  - Llama 3.3 8B / Llama 3.1 8B
  - Mistral Nemo 12B / Mistral 7B
  - Qwen 2.5 Coder 7B/14B
- 🚧 **Budget alerts & notifications** — Coming Soon

### [LLM0.ai](https://llm0.ai) *(Coming Soon)* (Managed Service)
- ✅ Everything in open source, plus:
- ✅ **Semantic caching** (36-40% hit rate, 60-89% cost savings)
- ✅ **Cost-based rate limiting** ($5/day per customer caps)
- ✅ **Customer attribution** (track per-user costs)
- ✅ **Budget alerts** (70%, 85%, 100% thresholds) *(Coming Soon)*
- ✅ **Multi-channel notifications** (email, webhook, Slack, PagerDuty) *(Coming Soon)*
- ✅ **Spend forecasting** (predict monthly costs) *(Coming Soon)*
- ✅ **Anomaly detection** (unusual spend patterns) *(Coming Soon)*
- ✅ **Managed infrastructure** (zero ops overhead)

---

**Questions?** Open an issue on GitHub.

**Production deployment?** Check out [LLM0.ai](https://llm0.ai) *(Coming Soon)*.
