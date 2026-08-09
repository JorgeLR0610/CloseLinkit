package web

import (
	_ "embed"
	"net/http"
)

//go:embed 404.html
var NotFoundPage []byte

func ServeNotFoundPage(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(NotFoundPage); err != nil {
		return err
	}

	return nil
}
