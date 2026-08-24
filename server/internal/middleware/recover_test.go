package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		expectedStatus int
		expectedBody   string
		expectedError  string
		expectLogged   bool
	}{
		{
			name: "No panic",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte("implicit ok body")); err != nil {
					t.Fatalf("could not write response: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "implicit ok body",
			expectLogged:   false,
		},
		{
			name: "Panic with string",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic("There was a panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  api.InternalErrorMsg,
			expectLogged:   true,
		},
		{
			name: "Panic with error value",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic(errors.New("boom"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  api.InternalErrorMsg,
			expectLogged:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

			middlewareFunc := middleware.Recover(logger)
			handler := middlewareFunc(tt.handlerFunc)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected response status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				if ct := w.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %q", ct)
				}

				var response struct {
					Error string `json:"error"`
				}

				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if response.Error != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response.Error)
				}
			} else if w.Body.String() != tt.expectedBody {
				t.Errorf("expected response body %q, got %q", tt.expectedBody, w.Body.String())
			}

			logged := logBuffer.Len() > 0
			if logged != tt.expectLogged {
				t.Errorf("expected logged=%v, got logged=%v (log output: %s)", tt.expectLogged, logged, logBuffer.String())
			}

			if tt.expectLogged {
				var logEntry map[string]any
				if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
					t.Fatalf("could not parse log output as JSON: %v", err)
				}

				if _, ok := logEntry["error"]; !ok {
					t.Error("expected log entry to contain an 'error' field")
				}
			}
		})
	}
}
