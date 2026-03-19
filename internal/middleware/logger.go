package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLogger returns middleware that logs every request's method, path,
// status code, and elapsed time.
func RequestLogger(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"latency", time.Since(start).String(),
				"remote", r.RemoteAddr,
			)
		})
	}
}

// statusRecorder captures the HTTP status code written by a downstream handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Status returns the captured status code.
func (r *statusRecorder) Status() int { return r.status }
