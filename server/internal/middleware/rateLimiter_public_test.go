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
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

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
