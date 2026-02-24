package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/config"
)

// Manager manages multiple LLM providers and handles failover
type Manager struct {
	providers map[string]Provider
	failover  map[string][]string // model -> [fallback models]
}

// NewManager creates a new provider manager
func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		providers: make(map[string]Provider),
		failover:  make(map[string][]string),
	}

	// Initialize providers based on available API keys
	if cfg.OpenAIAPIKey != "" {
		m.providers["openai"] = NewOpenAIProvider(cfg.OpenAIAPIKey)
	}
	if cfg.AnthropicAPIKey != "" {
		m.providers["anthropic"] = NewAnthropicProvider(cfg.AnthropicAPIKey)
	}
	if cfg.GeminiAPIKey != "" {
		m.providers["google"] = NewGeminiProvider(cfg.GeminiAPIKey)
	}

	// Setup failover chains
	m.setupFailoverChains()

	return m
}

// setupFailoverChains defines which models to fall back to
func (m *Manager) setupFailoverChains() {
	// OpenAI failover chains
	m.failover["gpt-4o"] = []string{"claude-sonnet-4-5-20250929", "gemini-2.5-pro"}
	m.failover["gpt-4o-mini"] = []string{"claude-haiku-4-5-20251001", "gemini-2.5-flash"}
	m.failover["gpt-4"] = []string{"claude-opus-4-5-20251101", "gemini-2.5-pro"}

	// Anthropic failover chains
	m.failover["claude-sonnet-4-5-20250929"] = []string{"gpt-4o", "gemini-2.5-pro"}
	m.failover["claude-haiku-4-5-20251001"] = []string{"gpt-4o-mini", "gemini-2.5-flash"}

	// Gemini failover chains
	m.failover["gemini-2.5-flash"] = []string{"gpt-4o-mini", "claude-haiku-4-5-20251001"}
	m.failover["gemini-2.5-pro"] = []string{"gpt-4o", "claude-sonnet-4-5-20250929"}
}

// GetProvider returns the provider for a given model
func (m *Manager) GetProvider(model string) (Provider, string, error) {
	// Determine provider by model name
	providerName := m.detectProvider(model)
	if providerName == "" {
		return nil, "", fmt.Errorf("unknown model: %s", model)
	}

	provider, ok := m.providers[providerName]
	if !ok {
		return nil, "", fmt.Errorf("provider %s not configured (check API key)", providerName)
	}

	return provider, providerName, nil
}

// detectProvider determines which provider a model belongs to
func (m *Manager) detectProvider(model string) string {
	if strings.HasPrefix(model, "gpt-") {
		return "openai"
	}
	if strings.HasPrefix(model, "claude-") {
		return "anthropic"
	}
	if strings.HasPrefix(model, "gemini-") {
		return "google"
	}
	return ""
}

// GetFailoverChain returns the failover models for a given model
func (m *Manager) GetFailoverChain(model string) []string {
	chain, ok := m.failover[model]
	if !ok {
		return []string{}
	}

	// Filter out models whose providers aren't configured
	var available []string
	for _, fallbackModel := range chain {
		providerName := m.detectProvider(fallbackModel)
		if _, ok := m.providers[providerName]; ok {
			available = append(available, fallbackModel)
		}
	}

	return available
}

// ChatCompletion makes a chat completion request with automatic failover
func (m *Manager) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, string, bool, error) {
	originalModel := req.Model
	originalProvider := m.detectProvider(originalModel)
	failoverUsed := false

	// Try the primary model first
	provider, providerName, err := m.GetProvider(req.Model)
	if err != nil {
		return nil, "", false, err
	}

	resp, err := provider.ChatCompletion(ctx, req)
	if err == nil {
		return resp, providerName, failoverUsed, nil
	}

	// Check if error is retryable (rate limit, timeout, server error)
	if !isRetryableError(err) {
		return nil, providerName, false, err
	}

	// Try failover chain
	failoverChain := m.GetFailoverChain(originalModel)
	for _, fallbackModel := range failoverChain {
		req.Model = fallbackModel
		provider, providerName, err := m.GetProvider(fallbackModel)
		if err != nil {
			continue
		}

		resp, err := provider.ChatCompletion(ctx, req)
		if err == nil {
			failoverUsed = true
			return resp, providerName, failoverUsed, nil
		}
	}

	return nil, originalProvider, false, fmt.Errorf("all providers failed for model %s", originalModel)
}

// isRetryableError checks if an error should trigger failover
func isRetryableError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "status 5")
}
