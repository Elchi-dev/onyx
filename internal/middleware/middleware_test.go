package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Elchi-dev/onyx/internal/middleware"
)

func TestChain(t *testing.T) {
	var order []string

	a := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "A-before")
			next.ServeHTTP(w, r)
			order = append(order, "A-after")
		})
	}
	b := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "B-before")
			next.ServeHTTP(w, r)
			order = append(order, "B-after")
		})
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	h := middleware.Chain(final, a, b)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	want := []string{"A-before", "B-before", "handler", "B-after", "A-after"}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("step %d: want %q, got %q", i, v, order[i])
		}
	}
}

func TestSecureHeaders(t *testing.T) {
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		middleware.SecureHeaders(),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
}
