package middleware

import (
	"context"
	"net"
	"net/http"
)

type clientIPKey struct{}

// ClientIP resolves the client IP once per request and stores it in the context.
// If trustProxy is true, it uses the X-Real-IP header set by Nginx (only when it
// is a valid IP); otherwise it uses the connection's remote address.
func ClientIP(trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := remoteIP(r)
			if trustProxy {
				if real := net.ParseIP(r.Header.Get("X-Real-IP")); real != nil {
					ip = real.String()
				}
			}
			ctx := context.WithValue(r.Context(), clientIPKey{}, ip)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClientIP returns the IP stored by ClientIP, falling back to the remote
// address when the middleware did not run (for example, in unit tests).
func GetClientIP(r *http.Request) string {
	if ip, ok := r.Context().Value(clientIPKey{}).(string); ok {
		return ip
	}
	return remoteIP(r)
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
