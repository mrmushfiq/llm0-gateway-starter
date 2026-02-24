package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AnthropicProvider handles Anthropic Claude API requests
type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
}

// AnthropicRequest represents a request to Anthropic's Messages API
type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float32           `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents a response from Anthropic's API
type AnthropicResponse struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Role    string                  `json:"role"`
	Content []AnthropicContentBlock `json:"content"`
	Model   string                  `json:"model"`
	Usage   AnthropicUsage          `json:"usage"`
}

// AnthropicContentBlock represents a content block
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to Anthropic
func (p *AnthropicProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Convert to Anthropic format
	anthropicReq, systemPrompt := p.convertRequest(req)
	if systemPrompt != "" {
		anthropicReq.System = systemPrompt
	}

	// Make HTTP request
	reqBody, _ := json.Marshal(anthropicReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Anthropic API error: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	return p.convertResponse(anthropicResp, latencyMs), nil
}

// ChatCompletionStream makes a streaming request
func (p *AnthropicProvider) ChatCompletionStream(ctx context.Context, req ChatRequest) (StreamReader, error) {
	anthropicReq, systemPrompt := p.convertRequest(req)
	anthropicReq.Stream = true
	if systemPrompt != "" {
		anthropicReq.System = systemPrompt
	}

	reqBody, _ := json.Marshal(anthropicReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Anthropic streaming API error: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	return &AnthropicStreamReader{
		reader: bufio.NewReader(httpResp.Body),
		resp:   httpResp,
	}, nil
}

// AnthropicStreamReader wraps the HTTP response for streaming
type AnthropicStreamReader struct {
	reader *bufio.Reader
	resp   *http.Response
}

// Recv reads the next streaming chunk
func (r *AnthropicStreamReader) Recv() (openai.ChatCompletionStreamResponse, error) {
	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return openai.ChatCompletionStreamResponse{}, err
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "event:") {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
				continue
			}

			chunk := openai.ChatCompletionStreamResponse{
				Object:  "chat.completion.chunk",
				Choices: []openai.ChatCompletionStreamChoice{},
			}

			eventType, _ := event["type"].(string)
			if eventType == "content_block_delta" {
				if delta, ok := event["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok && text != "" {
						chunk.Choices = []openai.ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: openai.ChatCompletionStreamChoiceDelta{
									Content: text,
								},
							},
						}
						return chunk, nil
					}
				}
			} else if eventType == "message_start" {
				chunk.Choices = []openai.ChatCompletionStreamChoice{
					{
						Index: 0,
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role: "assistant",
						},
					},
				}
				return chunk, nil
			}
		}
	}
}

// Close closes the stream
func (r *AnthropicStreamReader) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// convertRequest converts to Anthropic format
func (p *AnthropicProvider) convertRequest(req ChatRequest) (AnthropicRequest, string) {
	anthropicReq := AnthropicRequest{
		Model:       req.Model,
		Messages:    []AnthropicMessage{},
		MaxTokens:   4096,
		Temperature: req.Temperature,
	}

	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		anthropicReq.MaxTokens = *req.MaxTokens
	}

	var systemPrompt string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicReq.Messages = append(anthropicReq.Messages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	return anthropicReq, systemPrompt
}

// convertResponse converts Anthropic response to standard format
func (p *AnthropicProvider) convertResponse(resp AnthropicResponse, latencyMs int) *ChatResponse {
	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &ChatResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		LatencyMs: latencyMs,
	}
}

// ValidateModel checks if a model is valid
func (p *AnthropicProvider) ValidateModel(model string) bool {
	validModels := map[string]bool{
		"claude-opus-4-5-20251101":   true,
		"claude-sonnet-4-5-20250929": true,
		"claude-haiku-4-5-20251001":  true,
		"claude-3-5-haiku-20241022":  true,
	}
	return validModels[model]
}

// GetProviderName returns the provider name
func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}
