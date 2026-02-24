package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/cache"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/providers"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/database"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/models"
	"github.com/sashabaranov/go-openai"
)

type ChatHandler struct {
	providerMgr *providers.Manager
	cache       *cache.Cache
	db          *database.DB
}

func NewChatHandler(providerMgr *providers.Manager, cache *cache.Cache, db *database.DB) *ChatHandler {
	return &ChatHandler{
		providerMgr: providerMgr,
		cache:       cache,
		db:          db,
	}
}

// HandleChatCompletion handles POST /v1/chat/completions
func (h *ChatHandler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Get API key from context (set by auth middleware)
	apiKey, ok := ctx.Value("api_key").(*models.APIKey)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req providers.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Handle streaming separately
	if req.Stream {
		h.handleStreamingChat(w, r, apiKey, req)
		return
	}

	// Check cache if enabled
	var cacheHit bool
	var resp *providers.ChatResponse
	if apiKey.CacheEnabled {
		cachedResp, err := h.cache.Get(ctx, req)
		if err == nil {
			resp = cachedResp
			resp.CostUSD = 0 // Cache hits are free
			cacheHit = true
		}
	}

	// If not cached, call provider
	var providerName string
	var failoverUsed bool
	if !cacheHit {
		var err error
		resp, providerName, failoverUsed, err = h.providerMgr.ChatCompletion(ctx, req)
		if err != nil {
			http.Error(w, fmt.Sprintf("provider error: %v", err), http.StatusInternalServerError)
			h.logRequest(ctx, apiKey, req, nil, providerName, time.Since(startTime), false, failoverUsed, err)
			return
		}

		// Calculate cost
		cost, _ := h.calculateCost(ctx, providerName, req.Model, resp.Usage)
		resp.CostUSD = cost

		// Cache the response if enabled
		if apiKey.CacheEnabled {
			ttl := time.Duration(apiKey.CacheTTLSeconds) * time.Second
			h.cache.Set(ctx, req, resp, ttl)
		}
	}

	totalLatency := int(time.Since(startTime).Milliseconds())
	resp.LatencyMs = totalLatency

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Hit", fmt.Sprintf("%v", cacheHit))
	w.Header().Set("X-Cost-USD", fmt.Sprintf("%.6f", resp.CostUSD))
	w.Header().Set("X-Provider", providerName)
	w.Header().Set("X-Latency-Ms", fmt.Sprintf("%d", totalLatency))
	if failoverUsed {
		w.Header().Set("X-Failover", "true")
	}

	// Log request
	h.logRequest(ctx, apiKey, req, resp, providerName, time.Since(startTime), cacheHit, failoverUsed, nil)

	// Return response
	json.NewEncoder(w).Encode(resp)
}

// handleStreamingChat handles streaming chat completions
func (h *ChatHandler) handleStreamingChat(w http.ResponseWriter, r *http.Request, apiKey *models.APIKey, req providers.ChatRequest) {
	ctx := r.Context()
	startTime := time.Now()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get provider
	provider, providerName, err := h.providerMgr.GetProvider(req.Model)
	if err != nil {
		http.Error(w, fmt.Sprintf("provider error: %v", err), http.StatusInternalServerError)
		return
	}

	// Create stream
	stream, err := provider.ChatCompletionStream(ctx, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("streaming error: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Stream chunks
	var totalTokens int
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
			flusher.Flush()
			return
		}

		// Track usage
		if chunk.Usage != nil {
			totalTokens = chunk.Usage.TotalTokens
		}

		// Send chunk
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}

	// Send [DONE]
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// Log request
	usage := openai.Usage{TotalTokens: totalTokens}
	resp := &providers.ChatResponse{Usage: usage}
	cost, _ := h.calculateCost(ctx, providerName, req.Model, usage)
	resp.CostUSD = cost

	h.logRequest(ctx, apiKey, req, resp, providerName, time.Since(startTime), false, false, nil)
}

// calculateCost calculates the cost of a request
func (h *ChatHandler) calculateCost(ctx context.Context, provider, model string, usage openai.Usage) (float64, error) {
	pricing, err := h.db.GetModelPricing(ctx, provider, model)
	if err != nil {
		return 0, err
	}

	inputCost := float64(usage.PromptTokens) / 1000.0 * pricing.InputPer1kTokens
	outputCost := float64(usage.CompletionTokens) / 1000.0 * pricing.OutputPer1kTokens

	return inputCost + outputCost, nil
}

// logRequest logs the request to the database
func (h *ChatHandler) logRequest(ctx context.Context, apiKey *models.APIKey, req providers.ChatRequest, resp *providers.ChatResponse, provider string, duration time.Duration, cacheHit bool, failoverUsed bool, err error) {
	log := &models.GatewayLog{
		APIKeyID:     &apiKey.ID,
		Method:       "POST",
		Endpoint:     "/v1/chat/completions",
		Model:        req.Model,
		Provider:     provider,
		LatencyMs:    int(duration.Milliseconds()),
		CacheHit:     cacheHit,
		FailoverUsed: failoverUsed,
		StatusCode:   200,
	}

	if resp != nil {
		log.CostUSD = resp.CostUSD
		log.PromptTokens = resp.Usage.PromptTokens
		log.CompletionTokens = resp.Usage.CompletionTokens
		log.TotalTokens = resp.Usage.TotalTokens
	}

	if err != nil {
		log.StatusCode = 500
		errMsg := err.Error()
		log.ErrorMessage = &errMsg
	}

	// Log asynchronously to avoid blocking
	go h.db.LogRequest(context.Background(), log)

	// Update API key last used
	go h.db.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
}
