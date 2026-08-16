package middleware

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiter_SeparatesIPs(t *testing.T) {
	rl := NewIPRateLimiter(1, 1, time.Minute, time.Minute)

	limiterA := rl.getVisitor("1.1.1.1")

	// Consumes limiterA token
	limiterA.Allow()

	limiterB := rl.getVisitor("2.2.2.2")
	if !limiterB.Allow() {
		t.Fatal("IP B should have its own independent limiter")
	}
}

func TestIPRateLimiter_RefillsOverTime(t *testing.T) {
	rl := NewIPRateLimiter(rate.Every(50*time.Millisecond), 1, time.Minute, time.Minute)
	limiter := rl.getVisitor("1.2.3.4")

	limiter.Allow()

	time.Sleep(60 * time.Millisecond)

	if !limiter.Allow() {
		t.Fatal("token should have refilled after waiting")
	}
}

func TestIPRateLimiter_CleansExpiredEntries(t *testing.T) {
	rl := NewIPRateLimiter(1, 1, 10*time.Millisecond, 5*time.Millisecond)

	rl.getVisitor("1.2.3.4")

	go rl.CleanInactiveIPs()

	time.Sleep(50 * time.Millisecond)

	rl.mu.Lock()
	_, exists := rl.limiters["1.2.3.4"]
	rl.mu.Unlock()

	if exists {
		t.Fatal("expected expired IP to be removed from map")
	}
}

// The test passes if there is no panic or race condition detected
func TestIPRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewIPRateLimiter(100, 10, time.Minute, time.Minute)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ip := fmt.Sprintf("10.0.0.%d", n%5)
			limiter := rl.getVisitor(ip)
			limiter.Allow()
		}(i)
	}
	wg.Wait()
}
