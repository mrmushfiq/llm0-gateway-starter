package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the gateway
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Provider API Keys
	OpenAIAPIKey    string
	AnthropicAPIKey string
	GeminiAPIKey    string

	// Rate Limiting
	DefaultRateLimit int

	// Caching
	CacheTTLSeconds int
	CacheEnabled    bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		DefaultRateLimit: getEnvInt("DEFAULT_RATE_LIMIT", 100),
		CacheTTLSeconds:  getEnvInt("CACHE_TTL_SECONDS", 3600),
		CacheEnabled:     getEnvBool("CACHE_ENABLED", true),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// At least one provider API key is required
	if cfg.OpenAIAPIKey == "" && cfg.AnthropicAPIKey == "" && cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("at least one provider API key is required (OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY)")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
