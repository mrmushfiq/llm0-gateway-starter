# LLM0 Gateway Starter

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Production-ready LLM gateway with multi-provider support, automatic failover, streaming, and intelligent caching. Self-hosted alternative to LiteLLM and Portkey.

> Built in Go for performance and reliability. For advanced features (semantic caching, cost-based rate limiting, customer attribution), see [LLM0.ai](https://llm0.ai).

---

## Features

### Core Gateway
- **Multi-Provider Support** — OpenAI, Anthropic, Google Gemini with unified API
- **Automatic Failover** — Preset chains for 429s, 5xx errors, timeouts
- **Streaming (SSE)** — Real-time responses in OpenAI-compatible format
- **Exact-Match Caching** — Redis-backed with 12-15% hit rate
- **Token Bucket Rate Limiting** — Per-API-key with atomic Redis operations
- **Cost Tracking** — Per-request cost calculation and token counting
- **Request Logging** — PostgreSQL analytics for cost, latency, tokens

### Coming Soon
- Self-hosted open-source models via vLLM (Llama, Mistral, Qwen)
- Budget alerts & notifications

---

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 15+ (or [Neon](https://neon.tech) free tier)
- Redis 7+
- At least one LLM provider API key

### Automated Setup (Recommended)

```bash
git clone https://github.com/mrmushfiq/llm0-gateway-starter
cd llm0-gateway-starter

# Run setup script (handles everything)
./scripts/setup.sh

# Add your API key to .env
nano .env

# Start the gateway
go run cmd/gateway/main.go
```

Gateway runs at `http://localhost:8080`

### Manual Setup

```bash
# Clone and install dependencies
git clone https://github.com/mrmushfiq/llm0-gateway-starter
cd llm0-gateway-starter
go mod download

# Start infrastructure (auto-runs migrations)
docker-compose up -d

# Configure environment
cp .env.example .env
# Edit .env with your API keys

# Start the gateway
go run cmd/gateway/main.go
```

### Environment Variables

Required variables (see [.env.example](.env.example)):

```bash
# Server
PORT=8080

# Database (PostgreSQL 15+)
DATABASE_URL=postgresql://gateway:gateway@localhost:5432/gateway

# Redis
REDIS_URL=redis://localhost:6379

# LLM Provider API Keys (at least one required)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...

# Optional: Rate Limiting & Caching
DEFAULT_RATE_LIMIT=100
CACHE_TTL_SECONDS=3600
```

---

## Usage

### Basic Request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**Test API key:** `gw_test_abc123` (created by migration)

### Streaming

```bash
curl -N -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer gw_test_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Count to 5"}],
    "stream": true
  }'
```

### Response Headers

```http
X-Cache-Hit: miss
X-Cost-USD: 0.000009
X-Provider: openai
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-Latency-Ms: 28
```

---

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP POST /v1/chat/completions
       ▼
┌─────────────────────────────────────────┐
│        LLM Gateway (Go)                 │
│  1. Auth & Rate Limiting (Redis)        │
│  2. Cache Check (Redis)                 │
│  3. Provider Detection                  │
│  4. Failover Logic                      │
│  5. Streaming or Standard Response      │
│  6. Cost Tracking & Logging (Postgres)  │
└──────────┬──────────────────────────────┘
           │
           ▼ (Automatic failover)
┌──────────┬──────────────┬──────────────┐
│  OpenAI  │  Anthropic   │    Gemini    │
└──────────┴──────────────┴──────────────┘
```

**Tech Stack:**
- Go 1.21+ (performance, concurrency, single binary)
- PostgreSQL 15+ (logs, cost tracking, analytics)
- Redis 7+ (caching, rate limiting)
- Server-Sent Events (SSE) for streaming

---

## Performance

| Operation | Latency | Notes |
|-----------|---------|-------|
| Cache hit | 1-3ms | Redis lookup |
| Rate limit check | 2ms | Redis INCR |
| Provider API call | 200-3000ms | Provider-dependent |
| Streaming (first token) | 200-400ms | Provider-dependent |
| Database log | Non-blocking | Async via goroutine |

**Gateway overhead:** ~28ms (auth + rate limit + cache check + logging)

---

## Database Migrations

### How It Works

**With Docker (Automatic):**
```bash
docker-compose up -d
# Migrations run automatically on first start via /docker-entrypoint-initdb.d
```

**With setup.sh:**
```bash
./scripts/setup.sh
# Handles Docker startup + migrations automatically
```

**Manual:**
```bash
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
```

### Making Schema Changes

For production schema changes, we recommend [Atlas](https://atlasgo.io):

```bash
brew install ariga/tap/atlas

# Edit migrations/001_initial_schema.sql, then:
atlas migrate diff --to "file://migrations/001_initial_schema.sql" --dev-url "docker://postgres/15"
atlas schema apply --url "$DATABASE_URL" --to "file://migrations/001_initial_schema.sql"
```

Alternative: [golang-migrate](https://github.com/golang-migrate/migrate)

---

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test basic functionality (requires running gateway)
./scripts/test_basic.sh
```

---

## Feature Comparison

| Feature | This Starter | [LLM0.ai](https://llm0.ai) |
|---------|-------------|---------------------------|
| **Multi-provider** | ✅ OpenAI, Anthropic, Gemini | ✅ Same |
| **Automatic Failover** | ✅ Preset chains | ✅ Preset + custom |
| **Streaming (SSE)** | ✅ OpenAI-compatible | ✅ Same |
| **Caching** | ✅ Exact-match (12-15% hit rate) | ✅ Exact + Semantic (36-40% hit rate) |
| **Rate Limiting** | ✅ Token bucket (requests/min) | ✅ Token bucket + Cost-based ($/day) |
| **Cost Tracking** | ✅ Per-request | ✅ Per-request + Per-customer |
| **Request Logging** | ✅ PostgreSQL | ✅ PostgreSQL + real-time dashboards |
| **Customer Attribution** | ❌ | ✅ Multi-dimensional (customer/feature/team) |
| **Budget Alerts** | ❌ | ✅ Multi-threshold (70%/85%/100%) |
| **Notifications** | ❌ | ✅ Email, webhook, Slack, PagerDuty |
| **Spend Forecasting** | ❌ | ✅ Predictive analytics |
| **Anomaly Detection** | ❌ | ✅ Unusual spend patterns |
| **Shadow Mode** | ❌ | ✅ A/B testing |
| **Scheduled Jobs** | ❌ | ✅ Automatic maintenance |
| **Support** | Community | ✅ Priority support |
| **Deployment** | Self-hosted | ✅ Managed + self-hosted |

---

## Documentation

- [FEATURES.md](FEATURES.md) — Detailed feature list
- [ARCHITECTURE.md](ARCHITECTURE.md) — Technical design deep-dive
- [DEPLOYMENT.md](DEPLOYMENT.md) — Deployment options (local, Neon, production)
- [CONTRIBUTING.md](CONTRIBUTING.md) — Contribution guidelines
- [GETTING_STARTED.md](GETTING_STARTED.md) — Step-by-step setup guide

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

**Looking for help with:**
- New providers (Cohere, Mistral, Together AI)
- Failover chain improvements
- Performance optimizations
- Test coverage
- Documentation

---

## License

MIT License — Free to use, modify, and distribute. See [LICENSE](LICENSE).

---

## Links

- **Managed Version:** [LLM0.ai](https://llm0.ai) (Coming Soon)
- **Author:** [@mrmushfiq](https://github.com/mrmushfiq)
- **Twitter:** [@mushfiq_dev](https://twitter.com/mushfiq_dev)

---

**⭐ Star this repo if it helped you!**
