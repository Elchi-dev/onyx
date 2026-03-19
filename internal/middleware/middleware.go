// Package middleware provides composable HTTP middleware for the Onyx proxy.
//
// All middleware follows the standard Go pattern:
//
//	func(next http.Handler) http.Handler
//
// Use Chain to stack them in a readable order:
//
//	handler := middleware.Chain(router,
//		middleware.Recovery(log),
//		middleware.RequestLogger(log),
//		middleware.RateLimit(limiter),
//	)
package middleware

import "net/http"

// Middleware is the standard middleware function signature.
type Middleware func(http.Handler) http.Handler

// Chain wraps handler with the given middlewares.
// Middlewares are applied in the order provided — first listed is outermost.
//
// Example execution order for Chain(h, A, B, C):
//
//	Request  →  A  →  B  →  C  →  handler
//	Response ←  A  ←  B  ←  C  ←  handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse so first listed wraps outermost.
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
