package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name         string
		incomingID   string
		expectReused bool
	}{
		{
			name:         "honors valid incoming UUID",
			incomingID:   "16fd2706-8baf-433b-82eb-8c7fada847da",
			expectReused: true,
		},
		{
			name:         "generates new UUID when header is absent",
			incomingID:   "",
			expectReused: false,
		},
		{
			name:         "generates new UUID when header is invalid",
			incomingID:   "not-a-uuid",
			expectReused: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var idFromContext string

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				idFromContext, _ = r.Context().Value(middleware.RequestIDKey).(string)
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequestID(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incomingID != "" {
				req.Header.Set("X-Request-ID", tt.incomingID)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			idFromHeader := rr.Header().Get("X-Request-ID")

			if idFromHeader == "" {
				t.Fatal("expected X-Request-ID header to be set")
			}

			if _, err := uuid.Parse(idFromHeader); err != nil {
				t.Errorf("expected X-Request-ID to be a valid UUID, got %q: %v", idFromHeader, err)
			}

			if idFromContext != idFromHeader {
				t.Errorf("expected context ID (%q) to match header ID (%q)", idFromContext, idFromHeader)
			}

			if tt.expectReused && idFromHeader != tt.incomingID {
				t.Errorf("expected incoming ID %q to be reused, got %q", tt.incomingID, idFromHeader)
			}

			if !tt.expectReused && tt.incomingID != "" && idFromHeader == tt.incomingID {
				t.Errorf("expected a new ID to be generated, but incoming ID %q was reused", tt.incomingID)
			}
		})
	}

	t.Run("generates a different ID for each request without incoming header", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := middleware.RequestID(nextHandler)

		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		id1 := rr1.Header().Get("X-Request-ID")
		id2 := rr2.Header().Get("X-Request-ID")

		if id1 == id2 {
			t.Errorf("expected different request IDs, got the same value twice: %q", id1)
		}
	})
}
