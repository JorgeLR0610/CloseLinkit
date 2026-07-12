package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	WriteJSON(w, code, errorResponse{
		Error: msg,
	})
}

func WriteJSON(w http.ResponseWriter, code int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
