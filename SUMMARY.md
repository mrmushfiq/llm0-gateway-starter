# LLM0 Gateway Starter - Project Summary

## What This Is

**Open-source, production-ready LLM gateway** with multi-provider support, automatic failover, and intelligent caching.

**Use cases:**
- 🏢 **Self-hosted deployments** (full control over infrastructure)
- 🚀 **Small-scale production** (LLM spend < $500/month)
- 💼 **Enterprise on-premise** (air-gapped environments)
- 🔧 **Custom integrations** (modify as needed)

## What's Included

### Core Features ✅

1. **Multi-Provider Support**
   - OpenAI (GPT-4, GPT-4o, GPT-4o-mini)
   - Anthropic (Claude 4, Claude 3.5)
   - Google Gemini (Gemini 2.0, 2.5)
   - Unified OpenAI-compatible API

2. **Automatic Failover**
   - Preset failover chains
   - Triggers on 429, 5xx, timeouts
   - Zero configuration

3. **Streaming (SSE)**
   - Real-time responses
   - All providers → OpenAI format
   - Works with existing SDKs

4. **Basic Caching**
   - Redis exact-match
   - Configurable TTL
   - ~12-15% hit rate

5. **Rate Limiting**
   - Token bucket algorithm
   - Per-API-key limits
   - Standard headers

6. **Cost Tracking**
   - Per-request calculation
   - Token counting
   - Database-backed pricing

7. **Request Logging**
   - PostgreSQL storage
   - Cost, latency, tokens
   - Queryable analytics

### Tech Stack

- **Language:** Go 1.21+
- **Database:** PostgreSQL 15+ (or Neon free tier)
- **Cache:** Redis 7+
- **Deployment:** Docker, local development

## Advanced Features (LLM0.ai *Coming Soon*)

For production-scale deployments, [LLM0.ai](https://llm0.ai) *(Coming Soon)* offers:

### 🧠 Semantic Caching
- **This starter:** 12-15% cache hit rate (exact-match only)
- **LLM0:** 36-40% cache hit rate (exact + semantic)
- **Savings:** 60-89% cost reduction with semantic matching

### 🤖 Self-Hosted Models (vLLM)
- **This starter:** Not included (requires GPU infrastructure)
- **LLM0:** Llama 3.3 8B, Mistral Nemo 12B, Qwen 2.5 Coder with managed GPU deployment
- **Savings:** Run inference at ~$0.10/1M tokens vs. $0.15-$0.60 for cloud APIs

### 💰 Cost-Based Rate Limiting
- **This starter:** Token bucket (requests/min only)
- **LLM0:** Multi-dimensional cost limits ($5/day per customer)
- **Why it matters:** Protect profit margins from power users

### 📊 Customer Attribution
- **This starter:** Basic request logging
- **LLM0:** Per-customer, per-feature, per-client tracking
- **Why it matters:** Know who costs you money

### 🔧 Production Operations
- **This starter:** Manual maintenance
- **LLM0:** Scheduled jobs, budget alerts *(Coming Soon)*, monitoring, reconciliation, anomaly detection *(Coming Soon)*, multi-channel notifications *(Coming Soon)*

## File Structure

```
llm0-gateway-starter/
├── README.md              # Main documentation
├── ARCHITECTURE.md        # Technical deep-dive
├── CONTRIBUTING.md        # Contribution guide
├── LICENSE                # MIT License
├── Dockerfile             # Container definition
├── docker-compose.yml     # Local dev setup
├── go.mod                 # Go dependencies
├── .env.example           # Environment template
├── cmd/
│   └── gateway/           # Main entry point
├── internal/
│   ├── gateway/           # Gateway logic
│   │   ├── cache/         # Caching layer
│   │   ├── handlers/      # HTTP handlers
│   │   └── providers/     # LLM clients
│   └── shared/            # Utilities
│       ├── config/        # Configuration
│       ├── database/      # PostgreSQL
│       ├── models/        # Data models
│       └── redis/         # Redis client
├── migrations/            # SQL migrations
└── scripts/               # Helper scripts
    ├── setup.sh           # Setup automation
    └── test_basic.sh      # Basic tests
```

## Getting Started (3 Steps)

### 1. Setup

```bash
./scripts/setup.sh
```

This will:
- Check prerequisites
- Install dependencies
- Start PostgreSQL + Redis (Docker)
- Run migrations
- Create test API key

### 2. Configure

Edit `.env` and add at least one provider API key:

```env
OPENAI_API_KEY=sk-...
# or
ANTHROPIC_API_KEY=sk-ant-...
# or
GEMINI_API_KEY=...
```

### 3. Run

```bash
go run cmd/gateway/main.go
```

Test with:

```bash
./scripts/test_basic.sh
```

## Use Cases

### Self-Hosted Deployments

- Full control over infrastructure
- No vendor lock-in
- Custom modifications
- Air-gapped environments

### Small-Scale Production

- Startups with < $500/month LLM spend
- Internal tools and prototypes
- MVPs and proof-of-concepts
- Development environments

### Enterprise On-Premise

- Security-sensitive deployments
- Compliance requirements (HIPAA, SOC 2)
- Custom provider integrations
- Private cloud deployments

## Performance

| Operation | Latency | Notes |
|-----------|---------|-------|
| Cache hit | 1-3ms | Redis lookup |
| Rate limit check | 2ms | Redis INCR |
| OpenAI call | 1-3s | Provider-dependent |
| Streaming (first token) | 200-400ms | Provider-dependent |

## Deployment Options

### Local Development
```bash
docker-compose up -d
go run cmd/gateway/main.go
```

### Docker
```bash
docker build -t llm-gateway .
docker run -p 8080:8080 --env-file .env llm-gateway
```

### Cloud Platforms

For production deployment, you'll need to set up:
- Container registry (ECR, GCR, etc.)
- Container orchestration
- Managed databases
- Auto-scaling
- Monitoring

**For hassle-free production deployment:** [LLM0.ai](https://llm0.ai) *(Coming Soon)* handles all infrastructure for you.

## Choosing Between Self-Hosted and Managed

### Use This Project If:
- ✅ Self-hosting requirements
- ✅ Full infrastructure control needed
- ✅ LLM spend < $500/month
- ✅ Basic caching sufficient (12-15% hit rate)
- ✅ Standard rate limiting works for your use case

### Choose [LLM0.ai](https://llm0.ai) *(Coming Soon)* If:
- 🚀 Multi-tenant SaaS application
- 🚀 LLM spend > $500/month
- 🚀 Need semantic caching (36-40% hit rate, 60-89% cost savings)
- 🚀 Cost-based rate limiting required
- 🚀 Per-customer attribution needed
- 🚀 Budget alerts *(Coming Soon)* and anomaly detection *(Coming Soon)* needed
- 🚀 Prefer managed infrastructure (zero ops overhead)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

**Good first issues:**
- Add new provider (Cohere, Mistral)
- Improve error messages
- Add integration tests
- Improve documentation

**Please don't add:**
- Semantic caching (competitive advantage)
- Cost-based rate limiting (competitive advantage)
- Customer attribution (competitive advantage)

## License

MIT License - Free to use, modify, and distribute.

## Support

- **Issues:** Open a GitHub issue
- **Discussions:** Use GitHub Discussions
- **Production version:** [LLM0.ai](https://llm0.ai) *(Coming Soon)*

## Metrics (Lines of Code)

- **Go code:** ~1,500 lines
- **SQL:** ~200 lines
- **Documentation:** ~2,000 lines
- **Total:** ~3,700 lines

**Clean, readable, production-ready code.**

## About

Built by [@mrmushfiq](https://github.com/mrmushfiq).

**Related:** [LLM0.ai](https://llm0.ai) *(Coming Soon)* - Production LLM gateway with semantic caching and cost-based rate limiting.

---

**⭐ Star this repo if you find it useful!**
