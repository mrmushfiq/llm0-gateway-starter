# Features - LLM0 Gateway Starter

## Core Gateway Features

### 🌐 Multi-Provider Support

Unified API for three major LLM providers:

- **OpenAI:** GPT-4, GPT-4o, GPT-4o-mini, GPT-3.5-turbo
- **Anthropic:** Claude Opus 4.5, Sonnet 4.5, Haiku 4.5, Claude 3.5
- **Google Gemini:** Gemini 2.5 Pro/Flash, Gemini 2.0 Flash

**OpenAI-compatible endpoint:**
```bash
POST /v1/chat/completions
```

Works with existing SDKs (OpenAI, LangChain, LlamaIndex) - no code changes needed.

---

### 🔄 Automatic Failover

**Preset failover chains** handle provider failures automatically:

```
gpt-4o-mini fails (429 rate limit)
  ↓ Automatic retry
claude-haiku-4-5-20251001
  ↓ If also fails
gemini-2.5-flash
  ↓ Success
```

**Triggers:**
- 429 (rate limit exceeded)
- 5xx (server errors)
- Timeouts
- Connection failures

**Response headers:**
```http
X-Failover: true
X-Original-Provider: openai
X-Provider: anthropic
```

**Zero configuration required** - chains are preset for equivalent models.

---

### 📡 Streaming (SSE)

Real-time Server-Sent Events streaming for all providers:

**Unified format:**
```
data: {"choices":[{"delta":{"role":"assistant"}}]}
data: {"choices":[{"delta":{"content":"Hello"}}]}
data: {"choices":[{"delta":{"content":" world"}}]}
data: [DONE]
```

**All providers normalized to OpenAI format:**
- OpenAI → Native SSE
- Anthropic → Convert from Anthropic SSE
- Gemini → Convert from Gemini chunked JSON

**Compatible with:**
- OpenAI SDK streaming
- LangChain streaming
- Custom SSE clients
- EventSource API (browsers)

---

### 💾 Intelligent Caching

Redis-backed exact-match caching with automatic cache key generation:

**Cache key:** SHA-256 hash of `{model, messages, temperature, max_tokens, top_p}`

**Performance:**
- **Cache hit:** ~1-3ms latency
- **Cache miss:** Full provider call (~1-3s)
- **Hit rate:** 12-15% (identical queries only)
- **Cost savings:** $0 for cache hits

**Response headers:**
```http
X-Cache-Hit: exact
X-Cost-USD: 0.000000
X-Latency-Ms: 2
```

**Configurable per API key:**
- Enable/disable caching
- Custom TTL (default: 1 hour)

---

### 🚦 Rate Limiting

Token bucket algorithm using Redis atomic operations:

**Features:**
- Per-API-key limits (e.g., 100 requests/minute)
- Rolling window (1 minute)
- Atomic increment (Lua scripts)
- Standard headers

**Response headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
Retry-After: 60  (if exceeded)
```

**Status:** 429 Too Many Requests when exceeded

---

### 💰 Cost Tracking

Real-time cost calculation for every request:

**Calculation:**
```
cost = (prompt_tokens / 1000 * input_price) + 
       (completion_tokens / 1000 * output_price)
```

**Pricing stored in PostgreSQL:**
- Per-model pricing (input/output)
- Updates via migration
- Queryable for analytics

**Response:**
```json
{
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 8,
    "total_tokens": 20
  },
  "cost_usd": 0.000009,
  "latency_ms": 234
}
```

**Headers:**
```http
X-Cost-USD: 0.000009
X-Latency-Ms: 234
```

---

### 📊 Request Logging

Comprehensive PostgreSQL logging for analytics:

**Tracked per request:**
- Model, provider, endpoint
- Cost (USD)
- Latency (milliseconds)
- Token usage (prompt, completion, total)
- Cache hit (yes/no)
- Failover used (yes/no)
- Status code
- Error messages

**Queryable for:**
- Daily/monthly spend analysis
- Cache hit rate calculations
- Provider performance comparison
- Cost breakdown by model
- Latency percentiles

**Example query:**
```sql
SELECT model, COUNT(*), SUM(cost_usd), AVG(latency_ms)
FROM gateway_logs
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY model;
```

---

### 🔐 Authentication

**API key authentication** with SHA-256 hashing:

**Format:** `Authorization: Bearer gw_xxx...`

**Database storage:**
- Key hash (SHA-256)
- Key prefix (for display)
- Per-key rate limits
- Per-key cache settings
- Last used timestamp

**Security:**
- Keys never stored in plaintext
- Fast lookup (indexed hash)
- Automatic last_used tracking

---

### 🏗️ Architecture

**Clean separation of concerns:**

```
handlers/     → HTTP request handling, middleware
providers/    → LLM provider clients (OpenAI, Anthropic, Gemini)
cache/        → Caching layer (Redis)
database/     → PostgreSQL queries
redis/        → Redis operations
config/       → Configuration management
models/       → Data models
```

**Patterns used:**
- Provider interface (easy to add new providers)
- Manager pattern (failover orchestration)
- Middleware chain (auth, rate limiting, CORS)
- Repository pattern (database access)

---

## Performance Benchmarks

| Operation | Latency | Throughput |
|-----------|---------|------------|
| Cache hit | 1-3ms | 50K+ req/s |
| Rate limit check | 2-5ms | 20K+ req/s |
| Database log (async) | Non-blocking | N/A |
| Provider API call | 1-3s | Provider-limited |
| Streaming (first token) | 200-400ms | Provider-limited |

**End-to-end latency (optimized cloud deployment):**
- Exact-match cache hit: ~28ms
- Cache miss: ~230-3030ms (28ms gateway + 200-3000ms provider API)

**LLM0.ai with semantic caching:**
- Semantic cache hit: ~52ms (+24ms for semantic matching)
- Exact-match cache hit: ~28ms
- 3x better hit rates (36-40% vs 12-15%)

*Measured on optimized production infrastructure*

**Tested on:** MacBook Pro M1, 8GB RAM, local PostgreSQL + Redis

---

## What's NOT Included

### Advanced Features (Available in [LLM0.ai](https://llm0.ai) *Coming Soon*)

#### Semantic Caching
- **This project:** 12-15% hit rate (exact-match), ~28ms end-to-end
- **LLM0 *(Coming Soon)*:** 36-40% hit rate (exact + semantic), ~52ms end-to-end for semantic matches
- **Difference:** 3x better hit rates, up to 40% reduction in API costs (varies with query repetition patterns in your workload)

#### Open-Source Models via vLLM (LLM0 Managed)
- **This project:** Not included (requires GPU infrastructure)
- **LLM0 *(Coming Soon)*:** Llama 3.3 8B, Mistral Nemo 12B, Qwen 2.5 Coder with managed GPU deployment
- **Difference:** Run inference at ~$0.10/1M tokens, no rate limits, full control

#### Cost-Based Rate Limiting
- **This project:** Token bucket (requests/minute)
- **LLM0 *(Coming Soon)*:** Multi-dimensional cost limits (configurable per-customer spend caps, per-model caps, label attribution)

#### Customer Attribution
- **This project:** Per-API-key logging
- **LLM0 *(Coming Soon)*:** Per-customer, per-feature, per-client tracking with real-time dashboards

#### Production Operations
- **This project:** Manual maintenance
- **LLM0 *(Coming Soon)*:** Scheduled jobs, data reconciliation, budget alerts (70%/85%/100%) *(Coming Soon)*, anomaly detection *(Coming Soon)*, multi-channel notifications (email/webhook/Slack/PagerDuty) *(Coming Soon)*, managed infrastructure

[Compare features →](https://llm0.ai/pricing)

---

## Feature Comparison

| Feature | This Project | [LLM0.ai](https://llm0.ai) *(Coming Soon)* |
|---------|--------------|---------------------------|
| **Multi-provider** | ✅ 3 providers | ✅ 3 providers + Self-hosted (vLLM) |
| **Open-source models** | ❌ | ✅ Llama, Mistral, Qwen on LLM0 managed GPUs |
| **Failover** | ✅ Preset chains | ✅ Preset + custom |
| **Streaming** | ✅ SSE | ✅ SSE |
| **Caching** | ✅ Exact-match (12-15%) | ✅ Exact + Semantic (36-40%) |
| **Rate Limiting** | ✅ Token bucket | ✅ Token bucket + Cost-based |
| **Cost Tracking** | ✅ Per-request | ✅ Per-request + Per-customer |
| **Analytics** | ✅ PostgreSQL logs | ✅ Real-time dashboards |
| **Customer Attribution** | ❌ | ✅ Multi-dimensional |
| **Budget Alerts** | 🚧 Coming Soon | ✅ Multi-threshold (70%/85%/100%) |
| **Notifications** | 🚧 Coming Soon | ✅ Email/Webhook/Slack/PagerDuty |
| **Spend Forecasting** | ❌ | ✅ Predictive analytics |
| **Anomaly Detection** | 🚧 Coming Soon | ✅ Unusual spend patterns |
| **Shadow Mode** | ❌ | ✅ A/B testing |
| **Scheduled Jobs** | ❌ | ✅ Automatic maintenance |
| **Deployment** | Self-hosted | ✅ Managed + self-hosted |
| **Support** | Community | ✅ Priority support |

---

## Technical Highlights

### Provider Abstraction

```go
type Provider interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    ChatCompletionStream(ctx context.Context, req ChatRequest) (StreamReader, error)
    ValidateModel(model string) bool
    GetProviderName() string
}
```

Adding a new provider = implement 4 methods.

### Failover Manager

```go
// Automatic retries with equivalent models
resp, provider, failoverUsed, err := manager.ChatCompletion(ctx, req)

// If OpenAI fails, tries Anthropic, then Gemini
// Returns response from first successful provider
```

### Streaming Abstraction

```go
type StreamReader interface {
    Recv() (openai.ChatCompletionStreamResponse, error)
    Close() error
}
```

All providers implement the same interface - unified streaming.

### Atomic Rate Limiting

```go
// Redis atomic increment with TTL
exceeded, remaining, err := redis.CheckRateLimit(ctx, apiKeyID, limit)
// Uses INCR + EXPIRE in a single operation
```

No race conditions, perfect for high concurrency.

---

## Deployment Characteristics

### Stateless Design

- No local state (all in Redis/PostgreSQL)
- Horizontal scaling ready
- Load balancer compatible
- Rolling updates safe

### Resource Usage

- **Memory:** ~50MB baseline + in-flight requests
- **CPU:** Minimal (I/O bound, most time in provider APIs)
- **Disk:** None (logs to PostgreSQL)
- **Network:** Provider API calls + Redis + PostgreSQL

### Connection Pooling

- **PostgreSQL:** Max 25 connections, 10 idle, 5min lifetime
- **Redis:** Single connection per instance (pipelining supported)
- **HTTP:** Keep-alive enabled, 60s timeout

---

## API Compatibility

### Works With

- ✅ OpenAI Python SDK
- ✅ OpenAI Node.js SDK
- ✅ LangChain (OpenAI integration)
- ✅ LlamaIndex
- ✅ CrewAI
- ✅ AutoGen
- ✅ Any OpenAI-compatible client

### Migration Example

**Before:**
```python
client = OpenAI(api_key="sk-...")
```

**After:**
```python
client = OpenAI(
    base_url="https://your-gateway.com/v1",
    api_key="gw_..."  # Your gateway API key
)
```

**That's it!** No other code changes needed.

---

## Security Features

- ✅ API keys hashed with SHA-256
- ✅ Rate limiting per key (DDoS protection)
- ✅ SQL injection prevention (parameterized queries)
- ✅ CORS configuration
- ✅ Request logging (audit trail)
- ✅ Provider API keys never logged

---

## When to Use This vs. LLM0

### Use This Project:
- Self-hosting requirements (full control)
- LLM spend < $500/month
- Basic caching sufficient (12-15% hit rate)
- Standard rate limiting works
- Want full source code access

### Use [LLM0.ai](https://llm0.ai):
- Multi-tenant SaaS (need per-customer tracking)
- LLM spend > $500/month (semantic caching reduces API costs — savings scale with query volume and repetition)
- Need cost-based limits (protect margins)
- Want managed infrastructure (zero ops)
- Require advanced analytics

---

**Questions?** Open an issue on [GitHub](https://github.com/mrmushfiq/llm0-gateway-starter).

**Production deployment?** Check out [LLM0.ai](https://llm0.ai).
