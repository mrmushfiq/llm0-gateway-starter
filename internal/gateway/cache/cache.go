package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/providers"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/redis"
)

type Cache struct {
	redis *redis.Client
}

// New creates a new cache instance
func New(redisClient *redis.Client) *Cache {
	return &Cache{redis: redisClient}
}

// generateCacheKey generates a hash of the request for caching
func (c *Cache) generateCacheKey(req providers.ChatRequest) string {
	// Create a deterministic key from the request
	keyData := fmt.Sprintf("%s:%v:%v:%v:%v",
		req.Model,
		req.Messages,
		req.Temperature,
		req.MaxTokens,
		req.TopP,
	)

	hash := sha256.Sum256([]byte(keyData))
	return "cache:exact:" + hex.EncodeToString(hash[:])
}

// Get retrieves a cached response
func (c *Cache) Get(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	key := c.generateCacheKey(req)

	// Get from Redis
	val, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Deserialize
	var cachedResp providers.ChatResponse
	if err := json.Unmarshal([]byte(val), &cachedResp); err != nil {
		return nil, fmt.Errorf("failed to deserialize cached response: %w", err)
	}

	return &cachedResp, nil
}

// Set stores a response in cache
func (c *Cache) Set(ctx context.Context, req providers.ChatRequest, resp *providers.ChatResponse, ttl time.Duration) error {
	key := c.generateCacheKey(req)

	// Serialize response
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	// Store in Redis
	return c.redis.Set(ctx, key, string(data), ttl)
}
