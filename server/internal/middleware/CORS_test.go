package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://localhost:5173"}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Fatalf("could not write response: %v", err)
		}
	})

	corsMiddleware := middleware.CORSMiddleware(allowedOrigins)
	handlerToTest := corsMiddleware(nextHandler)

	tests := []struct {
		name               string
		method             string
		origin             string
		expectedCORSHeader string
		expectedStatus     int
	}{
		{
			name:               "Allowed origin receives correct header",
			method:             http.MethodGet,
			origin:             "http://localhost:5173",
			expectedCORSHeader: "http://localhost:5173",
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "Preflight OPTIONS request works",
			method:             http.MethodOptions,
			origin:             "http://localhost:5173",
			expectedCORSHeader: "http://localhost:5173",
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "Disallowed origin does not receive header",
			method:             http.MethodGet,
			origin:             "http://hacker.xyz",
			expectedCORSHeader: "",
			expectedStatus:     http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			corsHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if corsHeader != tt.expectedCORSHeader {
				t.Errorf("expected Access-Control-Allow-Origin header to be %q, got %q", tt.expectedCORSHeader, corsHeader)
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
