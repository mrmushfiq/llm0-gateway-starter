package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider handles OpenAI API requests
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
	}
}

// ChatCompletion makes a chat completion request to OpenAI
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Build OpenAI request
	openaiReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
	}

	if req.Temperature != nil {
		openaiReq.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		openaiReq.MaxTokens = *req.MaxTokens
	}
	if req.TopP != nil {
		openaiReq.TopP = *req.TopP
	}

	// Make request
	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	// Build response
	return &ChatResponse{
		ID:                resp.ID,
		Object:            resp.Object,
		Created:           resp.Created,
		Model:             resp.Model,
		Choices:           resp.Choices,
		Usage:             resp.Usage,
		SystemFingerprint: resp.SystemFingerprint,
		LatencyMs:         latencyMs,
	}, nil
}

// ChatCompletionStream creates a streaming chat completion request
func (p *OpenAIProvider) ChatCompletionStream(ctx context.Context, req ChatRequest) (StreamReader, error) {
	openaiReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   true,
	}

	if req.Temperature != nil {
		openaiReq.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		openaiReq.MaxTokens = *req.MaxTokens
	}
	if req.TopP != nil {
		openaiReq.TopP = *req.TopP
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI streaming API error: %w", err)
	}

	return &OpenAIStreamReader{stream: stream}, nil
}

// OpenAIStreamReader wraps OpenAI's stream
type OpenAIStreamReader struct {
	stream *openai.ChatCompletionStream
}

// Recv reads the next chunk
func (r *OpenAIStreamReader) Recv() (openai.ChatCompletionStreamResponse, error) {
	return r.stream.Recv()
}

// Close closes the stream
func (r *OpenAIStreamReader) Close() error {
	r.stream.Close()
	return nil
}

// ValidateModel checks if a model is valid for chat completions
func (p *OpenAIProvider) ValidateModel(model string) bool {
	validModels := map[string]bool{
		"gpt-4":             true,
		"gpt-4-turbo":       true,
		"gpt-4o":            true,
		"gpt-4o-mini":       true,
		"gpt-3.5-turbo":     true,
		"gpt-3.5-turbo-16k": true,
	}
	return validModels[model]
}

// GetProviderName returns the provider name
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}
