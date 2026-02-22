package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

// New creates a new Redis client
func New(ctx context.Context, redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis ping failed: %w", err)
	}

	return &Client{client: client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a value with TTL
func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Incr increments a counter
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Expire sets a TTL on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

// CheckRateLimit checks if the rate limit has been exceeded
// Uses a simple token bucket algorithm
func (c *Client) CheckRateLimit(ctx context.Context, apiKeyID string, limit int) (bool, int, error) {
	key := fmt.Sprintf("ratelimit:%s", apiKeyID)

	// Get current count
	count, err := c.client.Get(ctx, key).Int()
	if err == redis.Nil {
		// First request in this window - set count to 1
		if err := c.client.Set(ctx, key, 1, time.Minute).Err(); err != nil {
			return false, 0, err
		}
		return false, limit - 1, nil
	}
	if err != nil {
		return false, 0, err
	}

	// Check if limit exceeded
	if count >= limit {
		return true, 0, nil
	}

	// Increment counter
	newCount, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	// Set expiry if this is the first request
	if newCount == 1 {
		c.client.Expire(ctx, key, time.Minute)
	}

	remaining := limit - int(newCount)
	if remaining < 0 {
		remaining = 0
	}

	return false, remaining, nil
}
