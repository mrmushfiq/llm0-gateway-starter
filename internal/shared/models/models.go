package models

import "time"

// APIKey represents a gateway API key
type APIKey struct {
	ID                 string
	KeyHash            string
	KeyPrefix          string
	Name               string
	RateLimitPerMinute int
	CacheEnabled       bool
	CacheTTLSeconds    int
	IsActive           bool
	LastUsedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ModelPricing represents pricing for an LLM model
type ModelPricing struct {
	ID                  string
	Provider            string
	Model               string
	InputPer1kTokens    float64
	OutputPer1kTokens   float64
	ContextWindow       int
	SupportsStreaming   bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GatewayLog represents a request log entry
type GatewayLog struct {
	ID               string
	APIKeyID         *string
	Method           string
	Endpoint         string
	Model            string
	Provider         string
	CostUSD          float64
	LatencyMs        int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CacheHit         bool
	FailoverUsed     bool
	OriginalProvider *string
	StatusCode       int
	ErrorMessage     *string
	CreatedAt        time.Time
}
