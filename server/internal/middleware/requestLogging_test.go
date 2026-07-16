package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
)

func TestRequestLogging(t *testing.T) {
	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Explicit status code",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created body"))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "created body",
		},
		{
			name: "Implicit status code (200 OK)",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("implicit ok body"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "implicit ok body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

			middlewareFunc := middleware.RequestLogging(logger)
			handler := middlewareFunc(tt.handlerFunc)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected response status %d, got %d", tt.expectedStatus, w.Code)
			}
			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected response body %q, got %q", tt.expectedBody, w.Body.String())
			}

			var logEntry map[string]any
			if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
				t.Fatalf("could not parse log entry: %v", err)
			}

			status, ok := logEntry["status"].(float64)
			if !ok {
				t.Fatalf("log entry missing 'status' field or not a number: %+v", logEntry)
			}
			if int(status) != tt.expectedStatus {
				t.Errorf("expected log status %d, got %d", tt.expectedStatus, int(status))
			}
		})
	}
}
