// Package ratelimit provides a per-key token bucket rate limiter.
// It is intentionally decoupled from HTTP — the middleware glue lives in
// internal/middleware so this package stays testable and reusable.
package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// bucket is the token state for a single client key.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// Limiter is a token bucket rate limiter keyed by arbitrary string keys
// (typically client IP addresses).
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens refilled per second
	burst   float64 // maximum token capacity
}

// New creates a Limiter allowing rate requests per second with burst capacity.
// Stale bucket entries are cleaned up automatically every 5 minutes.
func New(rate float64, burst int) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   float64(burst),
	}
	go l.cleanup()
	return l
}

// Allow returns true if the given key is within its rate limit budget.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, lastSeen: now}
		l.buckets[key] = b
	}

	// Refill tokens proportional to elapsed time.
	b.tokens += now.Sub(b.lastSeen).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanup removes entries idle for more than 10 minutes to prevent memory growth.
func (l *Limiter) cleanup() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		l.mu.Lock()
		for key, b := range l.buckets {
			if time.Since(b.lastSeen) > 10*time.Minute {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}

// Middleware wraps next and enforces the limiter by client IP.
// Over-limit requests receive 429 Too Many Requests.
func Middleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !limiter.Allow(ip) {
				http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
