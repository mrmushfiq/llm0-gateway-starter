package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/models"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetAPIKey retrieves an API key by its raw key value
func (db *DB) GetAPIKey(ctx context.Context, rawKey string) (*models.APIKey, error) {
	// Hash the key
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT id, key_hash, key_prefix, name, rate_limit_per_minute, cache_enabled, 
		       cache_ttl_seconds, is_active, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	var apiKey models.APIKey
	err := db.conn.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&apiKey.Name,
		&apiKey.RateLimitPerMinute,
		&apiKey.CacheEnabled,
		&apiKey.CacheTTLSeconds,
		&apiKey.IsActive,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &apiKey, nil
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp
func (db *DB) UpdateAPIKeyLastUsed(ctx context.Context, apiKeyID string) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := db.conn.ExecContext(ctx, query, apiKeyID)
	return err
}

// GetModelPricing retrieves pricing for a model
func (db *DB) GetModelPricing(ctx context.Context, provider, model string) (*models.ModelPricing, error) {
	query := `
		SELECT id, provider, model, input_per_1k_tokens, output_per_1k_tokens,
		       context_window, supports_streaming, created_at, updated_at
		FROM model_pricing
		WHERE provider = $1 AND model = $2
	`

	var pricing models.ModelPricing
	err := db.conn.QueryRowContext(ctx, query, provider, model).Scan(
		&pricing.ID,
		&pricing.Provider,
		&pricing.Model,
		&pricing.InputPer1kTokens,
		&pricing.OutputPer1kTokens,
		&pricing.ContextWindow,
		&pricing.SupportsStreaming,
		&pricing.CreatedAt,
		&pricing.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pricing not found for %s/%s", provider, model)
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &pricing, nil
}

// LogRequest logs a gateway request
func (db *DB) LogRequest(ctx context.Context, log *models.GatewayLog) error {
	query := `
		INSERT INTO gateway_logs (
			api_key_id, method, endpoint, model, provider, cost_usd, latency_ms,
			prompt_tokens, completion_tokens, total_tokens, cache_hit, failover_used,
			original_provider, status_code, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := db.conn.ExecContext(ctx,
		query,
		log.APIKeyID,
		log.Method,
		log.Endpoint,
		log.Model,
		log.Provider,
		log.CostUSD,
		log.LatencyMs,
		log.PromptTokens,
		log.CompletionTokens,
		log.TotalTokens,
		log.CacheHit,
		log.FailoverUsed,
		log.OriginalProvider,
		log.StatusCode,
		log.ErrorMessage,
	)

	return err
}
