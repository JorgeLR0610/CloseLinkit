package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Text logger during development
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := godotenv.Load("../.env"); err != nil {
		logger.Error(
			"could not load .env file",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	ctx := context.Background()

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

	const port = "8080"

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

	// Handlers
	urlsHandler := api.NewURLHandler(urlsSvc, logger)

	mux := http.NewServeMux()

	// Endpoints
	mux.Handle("POST /api/v1/shorten", middleware.RequestLogging(logger)(http.HandlerFunc(urlsHandler.HandlerCreateURL)))
	mux.Handle("GET /api/v1/{shortCode}", middleware.RequestLogging(logger)(http.HandlerFunc(urlsHandler.HandlerGetURL)))
	mux.Handle("GET /api/v1/{shortCode}/stats", middleware.RequestLogging(logger)(http.HandlerFunc(urlsHandler.HandlerGetURLStats)))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logger.Info(
		"server running",
		slog.String("port", port),
	)

	if err := srv.ListenAndServe(); err != nil {
		logger.Error(
			"server stopped",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
