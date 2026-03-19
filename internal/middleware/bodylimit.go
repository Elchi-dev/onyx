package middleware

import (
	"fmt"
	"net/http"
)

// BodyLimit returns middleware that caps the size of incoming request bodies.
// Requests exceeding maxBytes receive a 413 Request Entity Too Large response.
// A sensible default is 10 MB (10 << 20).
func BodyLimit(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w,
					fmt.Sprintf("413 Request Too Large (max %d bytes)", maxBytes),
					http.StatusRequestEntityTooLarge,
				)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
