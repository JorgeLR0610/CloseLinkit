package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		trustProxy bool
		xRealIP    string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "trustProxy false ignores X-Real-IP and uses RemoteAddr host",
			trustProxy: false,
			xRealIP:    "203.0.113.195",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.1",
		},
		{
			name:       "trustProxy true with valid IPv4 in X-Real-IP",
			trustProxy: true,
			xRealIP:    "203.0.113.195",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "trustProxy true with valid IPv6 in X-Real-IP",
			trustProxy: true,
			xRealIP:    "2001:db8::1",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "trustProxy true with invalid IP string in X-Real-IP falls back to RemoteAddr",
			trustProxy: true,
			xRealIP:    "not-an-ip",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.1",
		},
		{
			name:       "trustProxy true with out of range IP in X-Real-IP falls back to RemoteAddr",
			trustProxy: true,
			xRealIP:    "999.999.999.999",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.1",
		},
		{
			name:       "trustProxy true with empty X-Real-IP falls back to RemoteAddr",
			trustProxy: true,
			xRealIP:    "",
			remoteAddr: "198.51.100.1:12345",
			expectedIP: "198.51.100.1",
		},
		{
			name:       "trustProxy true with IPv6 RemoteAddr and port",
			trustProxy: true,
			xRealIP:    "",
			remoteAddr: "[2001:db8::2]:8080",
			expectedIP: "2001:db8::2",
		},
		{
			name:       "RemoteAddr without port returns raw RemoteAddr",
			trustProxy: true,
			xRealIP:    "",
			remoteAddr: "198.51.100.1",
			expectedIP: "198.51.100.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedIP string

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedIP = middleware.GetClientIP(r)
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.ClientIP(tt.trustProxy)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if capturedIP != tt.expectedIP {
				t.Errorf("expected resolved IP %q, got %q", tt.expectedIP, capturedIP)
			}
		})
	}
}

func TestGetClientIP_WithoutMiddleware(t *testing.T) {
	t.Run("falls back to RemoteAddr with port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.10:54321"

		ip := middleware.GetClientIP(req)
		if ip != "192.0.2.10" {
			t.Errorf("expected %q, got %q", "192.0.2.10", ip)
		}
	})

	t.Run("falls back to raw RemoteAddr when without port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.10"

		ip := middleware.GetClientIP(req)
		if ip != "192.0.2.10" {
			t.Errorf("expected %q, got %q", "192.0.2.10", ip)
		}
	})
}
