package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
	"golang.org/x/time/rate"
)

func TestRateLimiting(t *testing.T) {
	tests := []struct {
		name         string
		rate         rate.Limit
		burst        int
		numRequests  int
		wantLastCode int
	}{
		{
			name:         "within burst allows request",
			rate:         1,
			burst:        3,
			numRequests:  3,
			wantLastCode: http.StatusOK,
		},
		{
			name:         "exceeding burst returns 429",
			rate:         1,
			burst:        3,
			numRequests:  4,
			wantLastCode: http.StatusTooManyRequests,
		},
		{
			name:         "single request always allowed",
			rate:         1,
			burst:        1,
			numRequests:  1,
			wantLastCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		var logBuffer bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

		t.Run(tt.name, func(t *testing.T) {
			rl := middleware.NewIPRateLimiter(tt.rate, tt.burst, time.Minute, time.Minute)
			handler := middleware.RateLimiting(rl, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "9.9.9.9:12345"

			var lastCode int
			for range tt.numRequests {
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				lastCode = rec.Code
			}

			if lastCode != tt.wantLastCode {
				t.Errorf("request %d: got status %d, want %d", tt.numRequests, lastCode, tt.wantLastCode)
			}
		})
	}
}

func TestRateLimiting_WithClientIP(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	t.Run("trustProxy true isolates clients behind same proxy", func(t *testing.T) {
		rl := middleware.NewIPRateLimiter(1, 1, time.Minute, time.Minute)
		handler := middleware.ClientIP(true)(
			middleware.RateLimiting(rl, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)

		// First request from Client A behind proxy 10.0.0.1
		reqA1 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA1.RemoteAddr = "10.0.0.1:5000"
		reqA1.Header.Set("X-Real-IP", "198.51.100.10")
		recA1 := httptest.NewRecorder()
		handler.ServeHTTP(recA1, reqA1)
		if recA1.Code != http.StatusOK {
			t.Fatalf("expected 200 for Client A first request, got %d", recA1.Code)
		}

		// First request from Client B behind same proxy 10.0.0.1
		reqB1 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqB1.RemoteAddr = "10.0.0.1:5000"
		reqB1.Header.Set("X-Real-IP", "198.51.100.20")
		recB1 := httptest.NewRecorder()
		handler.ServeHTTP(recB1, reqB1)
		if recB1.Code != http.StatusOK {
			t.Fatalf("expected 200 for Client B first request, got %d", recB1.Code)
		}

		// Second request from Client A (burst exhausted)
		reqA2 := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA2.RemoteAddr = "10.0.0.1:5000"
		reqA2.Header.Set("X-Real-IP", "198.51.100.10")
		recA2 := httptest.NewRecorder()
		handler.ServeHTTP(recA2, reqA2)
		if recA2.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 for Client A second request, got %d", recA2.Code)
		}
	})

	t.Run("trustProxy false aggregates clients behind same proxy", func(t *testing.T) {
		rl := middleware.NewIPRateLimiter(1, 1, time.Minute, time.Minute)
		handler := middleware.ClientIP(false)(
			middleware.RateLimiting(rl, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		)

		// First request from Client A behind proxy
		reqA := httptest.NewRequest(http.MethodGet, "/", nil)
		reqA.RemoteAddr = "10.0.0.1:5000"
		reqA.Header.Set("X-Real-IP", "198.51.100.10")
		recA := httptest.NewRecorder()
		handler.ServeHTTP(recA, reqA)
		if recA.Code != http.StatusOK {
			t.Fatalf("expected 200 for first request, got %d", recA.Code)
		}

		// Second request from Client B behind same proxy (X-Real-IP ignored, hits same 10.0.0.1 bucket)
		reqB := httptest.NewRequest(http.MethodGet, "/", nil)
		reqB.RemoteAddr = "10.0.0.1:5000"
		reqB.Header.Set("X-Real-IP", "198.51.100.20")
		recB := httptest.NewRecorder()
		handler.ServeHTTP(recB, reqB)
		if recB.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 for second request from same proxy when trustProxy is false, got %d", recB.Code)
		}
	})
}
