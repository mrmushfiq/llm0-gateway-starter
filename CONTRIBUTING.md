# Contributing

Thank you for your interest in contributing to LLM0 Gateway Starter!

This project provides core gateway functionality. For advanced features (semantic caching, cost-based limits, customer attribution), see [LLM0.ai](https://llm0.ai) *(Coming Soon)*.

## Good First Issues

Looking to contribute? Here are some beginner-friendly tasks:

### 🐛 Bug Fixes
- Fix error messages to be more descriptive
- Improve error handling in edge cases
- Fix typos in documentation

### 📚 Documentation
- Improve code documentation
- Add deployment examples
- Create architecture diagrams

### ✨ New Providers
- Add support for Cohere
- Add support for Mistral
- Add support for Together AI

### ⚡ Performance
- Optimize cache key generation
- Reduce memory allocations in hot paths
- Add connection pooling optimizations

### 🧪 Testing
- Add unit tests for providers
- Add integration tests
- Add load testing scripts

## Scope

This project focuses on **core gateway functionality**. Advanced features are available in [LLM0.ai](https://llm0.ai) *(Coming Soon)*:

**Not in scope for this repo:**
- Semantic caching (vector similarity search)
- Open-source models via vLLM (available in LLM0.ai managed platform)
- Cost-based rate limiting (per-customer spend caps)
- Customer attribution (multi-dimensional tracking)
- Scheduled maintenance jobs
- Budget alerts & notifications (email, webhook, Slack, PagerDuty)
- Spend forecasting & anomaly detection

These features are part of LLM0's managed service.

## How to Contribute

### 1. Fork the Repository

Click "Fork" in the top right of the GitHub page.

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/llm0-gateway-starter
cd llm0-gateway-starter
```

### 3. Create a Branch

```bash
git checkout -b feature/my-new-feature
```

### 4. Make Changes

- Follow Go best practices
- Keep code simple and readable
- Add comments for complex logic
- Update documentation if needed

### 5. Test Your Changes

```bash
# Run tests
go test ./...

# Test manually
./scripts/setup.sh
go run cmd/gateway/main.go
./scripts/test_basic.sh
```

### 6. Commit and Push

```bash
git add .
git commit -m "Add: brief description of changes"
git push origin feature/my-new-feature
```

### 7. Open a Pull Request

Go to the original repository and click "New Pull Request".

**PR Template:**

```markdown
## What does this PR do?

Brief description of the changes.

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement

## Checklist

- [ ] Code follows Go best practices
- [ ] Comments added for complex logic
- [ ] Documentation updated (if needed)
- [ ] Tested locally
- [ ] No breaking changes

## Screenshots (if applicable)

Add screenshots for UI changes.
```

## Code Style

### Go

- Use `gofmt` to format code
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Keep functions small and focused
- Use meaningful variable names

### Example:

```go
// Good
func calculateCost(tokens int, pricePerToken float64) float64 {
    return float64(tokens) * pricePerToken / 1000.0
}

// Bad
func calc(t int, p float64) float64 {
    return float64(t) * p / 1000.0
}
```

## Questions?

- Open an issue with the `question` label
- Check existing issues first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing!**
