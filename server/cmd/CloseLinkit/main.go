package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)


func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Could not create connection pool: %v\n", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Could not connect to database (ping failed): %v\n", err)
}

	const port = "8080"

	// Repository
	queries := repository.New(pool)
	
	// Generator
	gen, err := generator.NewShortCodeGenerator(7)
	if err != nil {
		log.Fatalf("Could not initialize generator: %v", err)
	}

	// Services
	urlsSvc := service.NewURLService(queries, gen)

	// Handlers
	urlsHandler := api.NewURLHandler(urlsSvc)


	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("POST /api/v1/shorten", urlsHandler.HandlerCreateURL)
	mux.HandleFunc("GET /api/v1/{shortCode}", urlsHandler.HandlerGetURL)
	mux.HandleFunc("GET /api/v1/{shortCode}/stats", urlsHandler.HandlerGetURLStats)

	srv := &http.Server {
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Server running on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}