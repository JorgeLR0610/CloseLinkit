package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu              sync.Mutex
	limiters        map[string]*visitor
	rate            rate.Limit
	burst           int
	expirationTime  time.Duration
	runningInterval time.Duration
}

func NewIPRateLimiter(r rate.Limit, burst int, expirationTime, runningInterval time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limiters:        make(map[string]*visitor),
		rate:            r,
		burst:           burst,
		expirationTime:  expirationTime,
		runningInterval: runningInterval,
	}
}

func (rl *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.limiters[ip]
	if !exists {
		visitor := &visitor{limiter: rate.NewLimiter(rl.rate, rl.burst), lastSeen: time.Now()}
		rl.limiters[ip] = visitor
		return visitor.limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *IPRateLimiter) CleanInactiveIPs() {
	for {
		time.Sleep(rl.runningInterval)
		rl.mu.Lock()
		for ip, v := range rl.limiters {
			if time.Since(v.lastSeen) > rl.expirationTime {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimiting(rl *IPRateLimiter, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			limiter := rl.getVisitor(ip)
			if !limiter.Allow() {
				if err := response.WriteError(w, http.StatusTooManyRequests, "Too Many Requests"); err != nil {
					logger.Error(
						"could not write error response",
						slog.Any("error", err),
					)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
