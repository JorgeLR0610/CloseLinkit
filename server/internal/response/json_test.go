package response_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/response"
)

func TestWriteJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	payload := map[string]string{"key": "value"}

	err := response.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected value, got %s", result["key"])
	}
}

func TestWriteJSON_Error(t *testing.T) {
	w := httptest.NewRecorder()
	// math.NaN() cannot be serialized to valid JSON
	unmarshalable := math.NaN()

	err := response.WriteJSON(w, http.StatusOK, unmarshalable)
	if err == nil {
		t.Fatal("expected error encoding NaN to JSON, got nil")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	errorMsg := "Resource not found"

	err := response.WriteError(w, http.StatusNotFound, errorMsg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal error body: %v", err)
	}
	if result["error"] != errorMsg {
		t.Errorf("expected error %s, got %s", errorMsg, result["error"])
	}
}

func TestWriteNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	response.WriteNoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %s", w.Body.String())
	}
}
