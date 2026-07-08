package api

import (
	"log"
	"net/http"
)


func main() {
	mux := http.NewServeMux()

	// Handlers
	//mux.HandleFunc("GET /api/shorten", )

	srv := &http.Server {
		Addr: ":8080",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}