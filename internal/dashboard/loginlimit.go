package dashboard

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// loginLimiter rate-limits login attempts per IP.
// Allows maxAttempts per window; rejects further attempts until the window expires.
type loginLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

// newLoginLimiter creates a limiter allowing maxAttempts per window per IP.
func newLoginLimiter(maxAttempts int, window time.Duration) *loginLimiter {
	l := &loginLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}
	go l.cleanup()
	return l
}

// allow returns true if the IP is within its attempt budget.
// It records the attempt regardless of outcome.
func (l *loginLimiter) allow(r *http.Request) bool {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Remove attempts outside the window.
	filtered := l.attempts[ip][:0]
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	l.attempts[ip] = filtered

	if len(l.attempts[ip]) >= l.maxAttempts {
		return false
	}

	l.attempts[ip] = append(l.attempts[ip], now)
	return true
}

// cleanup removes stale entries every 10 minutes to prevent memory growth.
func (l *loginLimiter) cleanup() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		l.mu.Lock()
		cutoff := time.Now().Add(-l.window)
		for ip, times := range l.attempts {
			filtered := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(l.attempts, ip)
			} else {
				l.attempts[ip] = filtered
			}
		}
		l.mu.Unlock()
	}
}
