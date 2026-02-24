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

// GeminiProvider handles Google Gemini API requests
type GeminiProvider struct {
	apiKey     string
	httpClient *http.Client
}

// GeminiRequest represents a request to Gemini's API
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents content in Gemini format
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation parameters
type GeminiGenerationConfig struct {
	Temperature     *float32 `json:"temperature,omitempty"`
	TopP            *float32 `json:"topP,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

// GeminiResponse represents a response from Gemini API
type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates"`
	UsageMetadata GeminiUsage       `json:"usageMetadata"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

// GeminiUsage represents token usage
type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to Gemini
func (p *GeminiProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	geminiReq := p.convertRequest(req)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		req.Model, p.apiKey)

	reqBody, _ := json.Marshal(geminiReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	return p.convertResponse(geminiResp, req.Model, latencyMs), nil
}

// ChatCompletionStream makes a streaming request
func (p *GeminiProvider) ChatCompletionStream(ctx context.Context, req ChatRequest) (StreamReader, error) {
	geminiReq := p.convertRequest(req)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s&alt=sse",
		req.Model, p.apiKey)

	reqBody, _ := json.Marshal(geminiReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini streaming API error: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Gemini API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	return &GeminiStreamReader{
		reader: bufio.NewReader(httpResp.Body),
		resp:   httpResp,
		model:  req.Model,
	}, nil
}

// GeminiStreamReader wraps the HTTP response for streaming
type GeminiStreamReader struct {
	reader *bufio.Reader
	resp   *http.Response
	model  string
}

// Recv reads the next streaming chunk
func (r *GeminiStreamReader) Recv() (openai.ChatCompletionStreamResponse, error) {
	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return openai.ChatCompletionStreamResponse{}, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			var geminiResp GeminiResponse
			if err := json.Unmarshal([]byte(dataStr), &geminiResp); err != nil {
				continue
			}

			chunk := r.convertChunkToOpenAI(geminiResp)
			return chunk, nil
		}
	}
}

// Close closes the stream
func (r *GeminiStreamReader) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// convertChunkToOpenAI converts Gemini chunk to OpenAI format
func (r *GeminiStreamReader) convertChunkToOpenAI(resp GeminiResponse) openai.ChatCompletionStreamResponse {
	chunk := openai.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("gemini-stream-%d", time.Now().UnixNano()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   r.model,
		Choices: []openai.ChatCompletionStreamChoice{},
	}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		var content string
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}

		choice := openai.ChatCompletionStreamChoice{
			Index: candidate.Index,
			Delta: openai.ChatCompletionStreamChoiceDelta{},
		}

		if candidate.Content.Role != "" {
			choice.Delta.Role = "assistant"
		}

		if content != "" {
			choice.Delta.Content = content
		}

		if candidate.FinishReason != "" {
			choice.FinishReason = openai.FinishReason("stop")
		}

		chunk.Choices = []openai.ChatCompletionStreamChoice{choice}
	}

	if resp.UsageMetadata.TotalTokenCount > 0 {
		chunk.Usage = &openai.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}

	return chunk
}

// convertRequest converts to Gemini format
func (p *GeminiProvider) convertRequest(req ChatRequest) GeminiRequest {
	geminiReq := GeminiRequest{
		Contents: make([]GeminiContent, 0),
	}

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}

		content := GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: msg.Content}},
		}
		geminiReq.Contents = append(geminiReq.Contents, content)
	}

	if req.Temperature != nil || req.MaxTokens != nil || req.TopP != nil {
		geminiReq.GenerationConfig = &GeminiGenerationConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
		}
	}

	return geminiReq
}

// convertResponse converts Gemini response to standard format
func (p *GeminiProvider) convertResponse(resp GeminiResponse, model string, latencyMs int) *ChatResponse {
	var content string
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			content += part.Text
		}
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
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
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
		LatencyMs: latencyMs,
	}
}

// ValidateModel checks if a model is valid
func (p *GeminiProvider) ValidateModel(model string) bool {
	validModels := map[string]bool{
		"gemini-2.5-flash":     true,
		"gemini-2.5-pro":       true,
		"gemini-2.0-flash":     true,
		"gemini-2.0-flash-exp": true,
	}
	return validModels[model]
}

// GetProviderName returns the provider name
func (p *GeminiProvider) GetProviderName() string {
	return "google"
}
