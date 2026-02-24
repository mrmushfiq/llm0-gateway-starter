package providers

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string                         `json:"model"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	Temperature *float32                       `json:"temperature,omitempty"`
	MaxTokens   *int                           `json:"max_tokens,omitempty"`
	TopP        *float32                       `json:"top_p,omitempty"`
	Stream      bool                           `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID                string                        `json:"id"`
	Object            string                        `json:"object"`
	Created           int64                         `json:"created"`
	Model             string                        `json:"model"`
	Choices           []openai.ChatCompletionChoice `json:"choices"`
	Usage             openai.Usage                  `json:"usage"`
	SystemFingerprint string                        `json:"system_fingerprint,omitempty"`
	LatencyMs         int                           `json:"latency_ms,omitempty"`
	CostUSD           float64                       `json:"cost_usd,omitempty"`
}

// StreamReader is an interface for streaming responses
type StreamReader interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close() error
}

// Provider is the interface all LLM providers must implement
type Provider interface {
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatCompletionStream(ctx context.Context, req ChatRequest) (StreamReader, error)
	ValidateModel(model string) bool
	GetProviderName() string
}
