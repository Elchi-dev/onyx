package middleware

import (
	"github.com/Elchi-dev/onyx/internal/ratelimit"
)

// RateLimit returns middleware that enforces the given Limiter.
// Requests that exceed the limit receive a 429 Too Many Requests response.
func RateLimit(limiter *ratelimit.Limiter) Middleware {
	return ratelimit.Middleware(limiter)
}
