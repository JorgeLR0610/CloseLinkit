package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/docs"
	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/swaggest/swgui/v5emb"
)

func main() {

	// Text logger during development
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load allowed origins env var
	corsOriginRaw := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	if corsOriginRaw == "" {
		logger.Error(
			"could not load ALLOWED_ORIGINS environment variable",
		)
		os.Exit(1)
	}
	allowedOrigins = strings.Split(corsOriginRaw, ",")

	// Load JWT secret env var
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Error(
			"could not load JWT_SECRET environment variable",
		)
		os.Exit(1)
	}

	// shortenRateLimiter middleware
	shortenRateLimiter := middleware.NewIPRateLimiter(1, 5, 15*time.Minute, 10*time.Minute)
	statsRateLimiter := middleware.NewIPRateLimiter(5, 10, 5*time.Minute, 2*time.Minute)

	ctx := context.Background()

	// Create pool connection
	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		logger.Error(
			"could not create connection pool",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error(
			"database ping failed",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	// Repository
	queries := repository.New(pool)

	// Generator
	gen, err := generator.NewShortCodeGenerator(7)
	if err != nil {
		logger.Error(
			"could not initiate generator",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	// Services
	urlsSvc := service.NewURLService(queries, gen)
	authSvc := service.NewAuthService(queries, service.AuthConfig{
		JWTSecret: []byte(jwtSecret),
	})

	// Handlers
	urlsHandler := api.NewURLHandler(urlsSvc, logger, os.Getenv("BASE_URL"))
	authHandler := api.NewAuthHandler(authSvc, logger)

	// Background goroutines to clean up inactive IPs
	go shortenRateLimiter.CleanInactiveIPs()
	go statsRateLimiter.CleanInactiveIPs()

	mux := http.NewServeMux()

	// Endpoints
	mux.Handle(
		"POST /api/v1/shorten",
		middleware.RequestLogging(logger)(
			middleware.RateLimiting(shortenRateLimiter, logger)(
				middleware.OptionalAuth(authSvc, logger)(
					http.HandlerFunc(urlsHandler.HandlerCreateURL),
				),
			),
		),
	)

	mux.Handle(
		"GET /api/v1/{shortCode}/stats",
		middleware.RequestLogging(logger)(
			middleware.RateLimiting(statsRateLimiter, logger)(
				http.HandlerFunc(urlsHandler.HandlerGetURLStats),
			),
		),
	)

	mux.Handle(
		"GET /{shortCode}",
		middleware.RequestLogging(logger)(
			http.HandlerFunc(urlsHandler.HandlerResolveShortURL),
		),
	)

	// Auth endpoints
	mux.Handle(
		"POST /api/v1/auth/register",
		middleware.RequestLogging(logger)(
			http.HandlerFunc(authHandler.HandlerRegister),
		),
	)

	mux.Handle(
		"POST /api/v1/auth/login",
		middleware.RequestLogging(logger)(
			http.HandlerFunc(authHandler.HandlerLogin),
		),
	)

	mux.Handle(
		"POST /api/v1/auth/refresh",
		middleware.RequestLogging(logger)(
			http.HandlerFunc(authHandler.HandlerRefreshToken),
		),
	)

	mux.Handle(
		"POST /api/v1/auth/logout",
		middleware.RequestLogging(logger)(
			http.HandlerFunc(authHandler.HandlerLogout),
		),
	)

	// Swagger UI
	mux.Handle(
		"GET /openapi.yaml",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			if _, err = w.Write(docs.OpenAPISpec); err != nil {
				logger.Error(
					"could not write openapi spec",
					slog.Any("error", err),
				)
			}
		}),
	)

	mux.Handle(
		"/docs/",
		v5emb.New(
			"CloseLinkit API",
			"/openapi.yaml",
			"/docs/",
		),
	)

	srv := &http.Server{
		Addr: ":8080",
		Handler: middleware.CORSMiddleware(allowedOrigins)(
			middleware.Recover(logger)(
				middleware.RequestID(mux),
			),
		),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	logger.Info(
		"server running",
		slog.String("port", srv.Addr),
	)

	if err := srv.ListenAndServe(); err != nil {
		logger.Error(
			"server stopped",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
