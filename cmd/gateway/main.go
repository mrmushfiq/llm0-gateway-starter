package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/cache"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/handlers"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/gateway/providers"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/config"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/database"
	"github.com/mrmushfiq/llm0-gateway-starter/internal/shared/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting LLM Gateway Starter on port %s (env: %s)", cfg.Port, cfg.Env)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Connected to PostgreSQL")

	// Initialize Redis
	redisClient, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Connected to Redis")

	// Initialize provider manager
	providerMgr := providers.NewManager(cfg)
	log.Println("✓ Initialized LLM providers")

	// Initialize cache
	cacheService := cache.New(redisClient)
	log.Println("✓ Initialized cache")

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(providerMgr, cacheService, db)
	middleware := handlers.NewMiddleware(db, redisClient)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(middleware.CORSMiddleware)

	// Health check (no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes (with auth and rate limiting)
	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RateLimitMiddleware)

		r.Post("/chat/completions", chatHandler.HandleChatCompletion)
	})

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server listening on http://localhost:%s", cfg.Port)
		log.Println("   POST /v1/chat/completions - Chat completions (OpenAI-compatible)")
		log.Println("   GET  /health              - Health check")
		log.Println("")
		log.Println("Ready to accept requests!")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
