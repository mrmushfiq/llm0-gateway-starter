package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/database"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/models"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/redis"
)

type Middleware struct {
	db    *database.DB
	redis *redis.Client
}

func NewMiddleware(db *database.DB, redis *redis.Client) *Middleware {
	return &Middleware{
		db:    db,
		redis: redis,
	}
}

// AuthMiddleware validates API keys
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		apiKeyValue := parts[1]

		// Validate API key
		apiKey, err := m.db.GetAPIKey(r.Context(), apiKeyValue)
		if err != nil {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		// Add API key to context
		ctx := context.WithValue(r.Context(), "api_key", apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware enforces rate limits
func (m *Middleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, ok := r.Context().Value("api_key").(*models.APIKey)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		limit := apiKey.RateLimitPerMinute
		if limit <= 0 {
			limit = 100 // fallback default
		}

		exceeded, remaining, err := m.redis.CheckRateLimit(r.Context(), apiKey.ID, limit)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if exceeded {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS
func (m *Middleware) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
