# LLM0 Gateway Starter

**Open-source, production-ready LLM gateway** with multi-provider support, automatic failover, and intelligent caching.

> **Self-hosted alternative to LiteLLM and Portkey.** Built in Go for performance and reliability. For advanced features like semantic caching and cost-based rate limiting, see [LLM0.ai](https://llm0.ai) *(Coming Soon)*.

---

## ✨ Features

### What's Included ✅

- **🌐 Multi-Provider Support**
  - OpenAI (GPT-4, GPT-4o, GPT-4o-mini)
  - Anthropic (Claude 4, Claude 3.5)
  - Google Gemini (Gemini 2.0, 2.5)
  - Unified OpenAI-compatible API

- **🔄 Automatic Failover**
  - Preset failover chains (e.g., OpenAI → Anthropic → Gemini)
  - Triggers: 429 rate limits, 5xx errors, timeouts
  - Zero configuration needed

- **📡 Streaming Support (SSE)**
  - Server-Sent Events for real-time responses
  - Works with all three providers
  - Unified format (OpenAI-compatible)

- **💾 Basic Caching**
  - Redis-backed exact-match caching
  - Configurable TTL
  - Cache hit headers

- **🚦 Token Bucket Rate Limiting**
  - Per-API-key limits
  - Redis Lua scripts (atomic operations)
  - Standard X-RateLimit-* headers

- **💰 Cost Tracking**
  - Per-request cost calculation
  - Token counting
  - Database-driven model pricing

- **📊 Request Logging**
  - PostgreSQL-backed logs
  - Cost, latency, tokens tracked
  - Queryable for analytics

### Coming Soon 🚧

- **🤖 Self-Hosted Models (vLLM + K8s)**
  - **Llama 3.3 8B / Llama 3.1 8B** — Sweet spot for K8s self-hosting
    - Fits on a single A100 40GB or T4 16GB (quantized)
    - Strong general performance for SaaS use cases
    - Your "cheap alternative to GPT-4o mini"
  - **Mistral Nemo 12B / Mistral 7B** — Apache 2.0, punches above its weight
    - Excellent for coding tasks
    - No usage restrictions
  - **Qwen 2.5 Coder 7B/14B** — Best coding-specialized model at small sizes
    - Perfect for SaaS devs building AI features
    - Optimized for developer tooling

- **🔔 Budget Alerts & Notifications** *(Coming Soon)*
  - Multi-threshold alerts (70%, 85%, 100%)
  - Spend forecasting & anomaly detection
  - Multi-channel notifications (email, webhook, Slack, PagerDuty)

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ (or [Neon](https://neon.tech) free tier)
- Redis 7+
- API keys for OpenAI, Anthropic, and/or Gemini

### 1. Clone & Setup

```bash
git clone https://github.com/yourusername/llm0-gateway-starter
cd llm0-gateway-starter

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env with your credentials
```

### 2. Database Setup

**Option A: Docker (Quick)**
```bash
docker-compose up -d postgres redis
```

**Option B: Neon (Free)**
```bash
# 1. Sign up at https://neon.tech
# 2. Create a database
# 3. Copy connection string to .env
```

### 3. Run Migrations

```bash
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
```

This creates tables and inserts a test API key: `gw_test_abc123`

### 4. Start the Gateway

```bash
go run cmd/gateway/main.go
```

Gateway runs at `http://localhost:8080`

---

## 📖 Usage

### Basic Request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming Request

```bash
curl -N -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Count to 5"}],
    "stream": true
  }'
```

### Response Headers

```http
HTTP/1.1 200 OK
X-Cache-Hit: miss
X-Cost-USD: 0.000009
X-Provider: openai
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-Latency-Ms: 234
```

**Note:** Use the test API key `gw_test_abc123` for testing (created by migration).

---

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP POST
       ▼
┌─────────────────────────────────────────────────┐
│           LLM Gateway (Go)                      │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  1. Auth & Rate Limiting                 │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  2. Cache Check (Redis)                  │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  3. Provider Detection                   │  │
│  │     (OpenAI / Anthropic / Gemini)        │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  4. Failover Logic                       │  │
│  │     (Retry with next provider if fails)  │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  5. Streaming or Standard Response       │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  6. Cost Tracking & Logging              │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
       │
       ▼ (One of three)
┌──────────┬──────────────┬──────────────┐
│  OpenAI  │  Anthropic   │    Gemini    │
└──────────┴──────────────┴──────────────┘
```

---

## 🏗️ Technical Stack

- **Language:** Go 1.21+ (performance, concurrency, single binary deployment)
- **Database:** PostgreSQL 15+ (request logs, cost tracking, analytics)
- **Cache:** Redis 7+ (exact-match caching, rate limiting)
- **Streaming:** Server-Sent Events (SSE) - OpenAI-compatible format
- **Deployment:** Docker, single binary, cloud-agnostic

---

## 🆚 Feature Comparison

### Basic vs. Advanced Features

### 🧠 **Semantic Caching**

**What this project includes:**
- ✅ Exact-match caching (Redis)
- ✅ ~12-15% cache hit rate
- ✅ Sub-millisecond cache responses

**What [LLM0.ai](https://llm0.ai) *(Coming Soon)* adds:**
- ✅ Semantic similarity matching
- ✅ 36-40% cache hit rate (3x better)
- ✅ 60-89% cost reduction
- ✅ Vector-based similarity search

**Why it matters:**
- "What is AI?" matches "Explain artificial intelligence"
- Users rephrase questions constantly
- Exact-match alone misses 75% of cacheable queries

[Learn more →](https://llm0.ai/features/semantic-caching)

---

### 💰 **Cost-Based Rate Limiting**

**What this project includes:**
- ✅ Token bucket rate limiting (requests/minute)
- ✅ Per-API-key limits
- ✅ Redis atomic operations

**What [LLM0.ai](https://llm0.ai) *(Coming Soon)* adds:**
- ✅ Cost-based limits ($X/day per customer)
- ✅ Multi-dimensional tracking (cost + requests + model + labels)
- ✅ Soft degradation (downgrade model instead of blocking)
- ✅ Real-time spend tracking

**Why it matters:**
- 1000 GPT-4 requests ≠ 1000 GPT-4o-mini requests (cost difference: 50x)
- Standard rate limiting doesn't protect profit margins
- Power users can destroy your economics

**This project includes:**
- ✅ Token bucket rate limiting (requests/minute)
- ✅ Per-API-key limits
- ✅ Redis atomic operations

**[LLM0.ai](https://llm0.ai) *(Coming Soon)* adds:**
- ✅ **Cost-based limits** ($X/day per customer)
- ✅ **Per-customer tracking** (who costs you what)
- ✅ **Label attribution** (track by feature, team, client)
- ✅ **Real-time dashboards** (spend by customer, model, label)

[Learn more about LLM0's cost-based rate limiting →](https://llm0.ai/features/cost-based-limits)

---

### 📊 **Customer Attribution & Analytics**

**What this project includes:**
- ✅ Request logging (PostgreSQL)
- ✅ Cost and latency tracking
- ✅ Per-model analytics

**What [LLM0.ai](https://llm0.ai) *(Coming Soon)* adds:**
- ✅ Per-customer spend tracking (`X-Customer-ID`)
- ✅ Label-based attribution (`X-LLM0-Feature`, `X-LLM0-Client`)
- ✅ Real-time dashboards ("who costs me money")
- ✅ Budget alerts (70%, 85%, 100% thresholds) *(Coming Soon)*
- ✅ Spend forecasting (predict monthly costs) *(Coming Soon)*
- ✅ Multi-channel notifications (email, webhook, Slack, PagerDuty) *(Coming Soon)*

**Why it matters:**
- SaaS: Track costs per end-user
- Agencies: Track costs per client
- Multi-tenant: Prevent one user from blowing your budget

**This project includes:**
- ✅ Request logging (PostgreSQL)
- ✅ Cost, latency, token tracking
- ✅ Queryable analytics

**[LLM0.ai](https://llm0.ai) adds:**
- ✅ Per-customer spend tracking
- ✅ Multi-dimensional attribution (feature, team, client)
- ✅ Real-time headers (`X-Customer-Spend-Today`)
- ✅ Spend forecasting & anomaly detection
- ✅ Multi-channel alerts (email, webhook, Slack, PagerDuty)

[Learn more about LLM0's analytics →](https://llm0.ai/features/analytics)

---

## 🎯 Comparison: Starter vs. LLM0

| Feature | This Starter | [LLM0.ai](https://llm0.ai) |
|---------|--------------|---------------------------|
| **Multi-provider** | ✅ 3 providers | ✅ 3 providers |
| **Failover** | ✅ Basic | ✅ Advanced + configurable |
| **Streaming** | ✅ SSE | ✅ SSE |
| **Caching** | ✅ Exact-match (Redis) | ✅ Exact + **Semantic** |
| **Rate Limiting** | ✅ Token bucket | ✅ Token bucket + **Cost-based** |
| **Cost Tracking** | ✅ Per-request | ✅ Per-request + **Per-customer** |
| **Customer Attribution** | ❌ | ✅ **Multi-dimensional** |
| **Soft Limits** | ❌ | ✅ **Model downgrade** |
| **Label Tracking** | ❌ | ✅ **Feature/team/client** |
| **Budget Alerts** | ❌ | ✅ **Email/webhook/Slack/PagerDuty** |
| **Spend Forecasting** | ❌ | ✅ **Predictive alerts** |
| **Shadow Mode** | ❌ | ✅ **A/B testing** |
| **Scheduled Jobs** | ❌ | ✅ **Automatic maintenance** |
| **Dashboard** | ❌ | ✅ **Real-time analytics** |
| **Support** | Community | ✅ **Priority support** |

---

## 🤝 Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**We'd love help with:**
- New providers (Cohere, Mistral, Together AI)
- Failover chain improvements
- Performance optimizations
- Additional test coverage
- Documentation enhancements

---

## 📜 License

MIT License - Free to use, modify, and distribute.

**Attribution appreciated but not required.**

---

## 🙏 Acknowledgments

Created by Mushfiq Rahman [@mrmushfiq](https://github.com/mrmushfiq).

---

## 🔗 Links

- **Production Version:** [LLM0.ai](https://llm0.ai)
- [Twitter](https://twitter.com/mushfiq_dev)

---

**⭐ If this helped you, please star the repo!**

**Questions? Open an issue or check out the [full version at LLM0.ai](https://llm0.ai)** (Coming soon)
