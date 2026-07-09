package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/joho/godotenv"
)


func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	const port = "8080"

	// Repository
	dbQueries := repository.New(db)
	
	// Generator
	gen, err := generator.NewShortCodeGenerator(7)
	if err != nil {
		log.Fatalf("Could not initiate generator: %w", err)
	}

	// Services
	urlsSvc := service.NewURLService(dbQueries, gen)


	mux := http.NewServeMux()

	// Handlers
	//mux.HandleFunc("GET /api/shorten", )

	srv := &http.Server {
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Server running on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}