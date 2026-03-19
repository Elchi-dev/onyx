package proxy

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// defaultTransport is shared across all proxied routes.
// It enforces timeouts and connection limits so a slow or dead backend
// cannot exhaust goroutines or file descriptors.
var defaultTransport = &http.Transport{
	// Dial timeout: abort if we cannot reach the backend in 10s.
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	// TLS handshake timeout.
	TLSHandshakeTimeout: 10 * time.Second,
	// How long to wait for a backend's response headers.
	ResponseHeaderTimeout: 60 * time.Second,
	// Idle connection pool — reuse TCP connections to backends.
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
	// Disable HTTP/2 to backends for now; add when we have proper testing.
	ForceAttemptHTTP2: false,
}

// Router routes incoming HTTP requests to the correct backend by Host header.
// It is safe for concurrent use.
type Router struct {
	mu           sync.RWMutex
	routes       map[string]*httputil.ReverseProxy
	log          *slog.Logger
	eventHandler EventHandler
}

// New creates a Router with the given logger and optional event handler.
// Pass nil for eventHandler to disable event emission.
func New(log *slog.Logger, eventHandler EventHandler) *Router {
	if eventHandler == nil {
		eventHandler = func(RequestEvent) {}
	}
	return &Router{
		routes:       make(map[string]*httputil.ReverseProxy),
		log:          log,
		eventHandler: eventHandler,
	}
}

// AddRoute registers a host → target proxy route.
// Calling AddRoute with an existing host replaces the route atomically.
func (r *Router) AddRoute(host, target string) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parsing target %q for host %q: %w", target, host, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// Use the shared transport so all backends share the connection pool.
	proxy.Transport = defaultTransport
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		r.log.Error("proxy backend error",
			"host", host,
			"target", target,
			"error", err,
		)
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
	}

	r.mu.Lock()
	r.routes[host] = proxy
	r.mu.Unlock()

	r.log.Info("route registered", "host", host, "target", target)
	return nil
}

// RemoveRoute removes the route for host. Safe to call concurrently.
func (r *Router) RemoveRoute(host string) {
	r.mu.Lock()
	delete(r.routes, host)
	r.mu.Unlock()
	r.log.Info("route removed", "host", host)
}

// ServeHTTP dispatches the request to the registered backend.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	r.mu.RLock()
	proxy, ok := r.routes[req.Host]
	r.mu.RUnlock()

	if !ok {
		r.log.Warn("no route for host", "host", req.Host)
		http.Error(w, "502 Bad Gateway — no route configured for this host", http.StatusBadGateway)
		return
	}

	rec := newStatusRecorder(w)
	proxy.ServeHTTP(rec, req)

	r.eventHandler(RequestEvent{
		Timestamp: start,
		Host:      req.Host,
		Method:    req.Method,
		Path:      req.URL.Path,
		Status:    rec.status,
		LatencyMs: time.Since(start).Milliseconds(),
		ClientIP:  req.RemoteAddr,
	})
}

// statusRecorder captures the HTTP status code written by a downstream handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
