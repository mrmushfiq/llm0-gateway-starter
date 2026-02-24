# Architecture - LLM0 Gateway Starter

## System Design

Production-ready LLM gateway built with Go, designed for high performance, reliability, and horizontal scalability.

```
llm0-gateway-starter/
├── cmd/
│   └── gateway/           # Main entry point
│       └── main.go
├── internal/
│   ├── gateway/           # Gateway-specific logic
│   │   ├── cache/         # Caching layer
│   │   ├── handlers/      # HTTP handlers
│   │   └── providers/     # LLM provider clients
│   └── shared/            # Shared utilities
│       ├── config/        # Configuration
│       ├── database/      # PostgreSQL client
│       ├── models/        # Data models
│       └── redis/         # Redis client
├── migrations/            # SQL migrations
├── scripts/               # Helper scripts
├── Dockerfile             # Container definition
└── docker-compose.yml     # Local development setup
```

## Request Flow

1. **HTTP Request** → Client sends POST to `/v1/chat/completions`
2. **Middleware Chain:**
   - CORS handling
   - Authentication (validate API key)
   - Rate limiting (check Redis)
3. **Handler:**
   - Parse request body
   - Check cache (Redis) if enabled
   - If cache miss:
     - Route to appropriate provider (OpenAI, Anthropic, or Gemini)
     - Handle failover if primary provider fails
     - Calculate cost based on token usage
     - Store in cache
   - Return response with cost/latency headers
4. **Logging:**
   - Asynchronously log to PostgreSQL
   - Update API key last_used_at

## Key Components

### Provider Manager

Handles multi-provider support and automatic failover:

```go
// Detect provider by model name
gpt-4o-mini → OpenAI
claude-sonnet → Anthropic  
gemini-2.5-flash → Google

// Failover chains (preset)
gpt-4o-mini → claude-haiku → gemini-2.5-flash
```

### Cache Layer

Simple exact-match caching using Redis:

- **Key:** SHA-256 hash of `{model, messages, temperature, max_tokens, top_p}`
- **Value:** Serialized ChatResponse
- **TTL:** Configurable per API key (default: 1 hour)

### Cost Tracking

Calculates cost per request:

```
cost = (prompt_tokens / 1000 * input_price) + 
       (completion_tokens / 1000 * output_price)
```

Pricing stored in PostgreSQL (`model_pricing` table).

### Rate Limiting

Token bucket algorithm using Redis:

- **Window:** 1 minute
- **Limit:** Configurable per API key (default: 100 req/min)
- **Implementation:** Redis INCR + EXPIRE

### Streaming (SSE)

Server-Sent Events for real-time responses:

- All providers converted to OpenAI-compatible format
- Chunks sent as `data: {json}\n\n`
- Terminates with `data: [DONE]\n\n`

## Database Schema

### api_keys

Stores gateway API keys for authentication.

```sql
- id (UUID)
- key_hash (SHA-256)
- key_prefix (for display)
- rate_limit_per_minute
- cache_enabled
- cache_ttl_seconds
```

### model_pricing

Stores per-model pricing for cost calculation.

```sql
- provider (openai, anthropic, google)
- model
- input_per_1k_tokens
- output_per_1k_tokens
```

### gateway_logs

Stores request logs for analytics.

```sql
- model, provider
- cost_usd, latency_ms
- prompt_tokens, completion_tokens
- cache_hit, failover_used
```

## Deployment

### Local Development

```bash
docker-compose up -d  # PostgreSQL + Redis
go run cmd/gateway/main.go
```

### Docker

```bash
docker build -t llm-gateway .
docker run -p 8080:8080 --env-file .env llm-gateway
```

### Cloud Deployment

1. Build Docker image
2. Set environment variables
3. Connect to managed PostgreSQL (e.g., Neon) and Redis
4. Deploy to your cloud provider of choice

**For production:** [LLM0.ai](https://llm0.ai) *(Coming Soon)* handles deployment and scaling for you.

## Performance

- **Cache hit:** ~1-3ms (Redis lookup)
- **Rate limit check:** ~2ms (Redis INCR)
- **Database logging:** Async (non-blocking)
- **Concurrency:** Go's goroutines handle 1000s of concurrent requests

## Security

- API keys stored as SHA-256 hashes
- Rate limiting per key
- SQL injection protection (parameterized queries)
- CORS configured

## Advanced Features

This project provides core gateway functionality. For production-scale deployments with high traffic, [LLM0.ai](https://llm0.ai) *(Coming Soon)* offers:

- **Semantic caching** (36-40% hit rate vs. 12-15%)
- **Self-hosted models** (vLLM: Llama, Mistral, Qwen with managed GPU infrastructure)
- **Cost-based rate limiting** ($5/day per customer caps)
- **Customer attribution** (multi-dimensional tracking)
- **Scheduled maintenance** (automatic cache cleanup, reconciliation)
- **Budget alerts** (70%, 85%, 100% thresholds) *(Coming Soon)*
- **Multi-channel notifications** (email, webhook, Slack, PagerDuty) *(Coming Soon)*
- **Spend forecasting** (predict monthly costs) *(Coming Soon)*
- **Anomaly detection** (unusual spend patterns) *(Coming Soon)*
- **Managed infrastructure** (global deployment, auto-scaling)
