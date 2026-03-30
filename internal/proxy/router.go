package proxy

import (
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Elchi-dev/onyx/internal/ratelimit"
)

// RouteData is the full configuration for one proxy route.
type RouteData struct {
	Host        string
	Target      string
	HTTPS       bool
	WWWRedirect string // "strip" removes www, "add" adds www
	Gzip        bool
	MaxBodySize int64
	TimeoutSecs int
	Paths       []PathRule
	StaticRoot  string
	StaticSPA   bool
	RespHeaders map[string]string
	RateLimit   RateLimitData
}

// PathRule maps a URL path prefix to an optional different backend.
type PathRule struct {
	Path   string
	Target string // empty = use RouteData.Target
}

// RateLimitData holds rate limit settings for a route.
type RateLimitData struct {
	RequestsPerSecond float64
	Burst             int
}

// defaultTransport is shared across all proxied routes.
var defaultTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 60 * time.Second,
	MaxIdleConns:          200,
	MaxIdleConnsPerHost:   20,
	IdleConnTimeout:       90 * time.Second,
	ForceAttemptHTTP2:     false,
}

// Router routes HTTP requests to the correct backend by host.
type Router struct {
	mu           sync.RWMutex
	routes       map[string]http.Handler
	log          *slog.Logger
	eventHandler EventHandler
}

// New creates a Router.
func New(log *slog.Logger, eventHandler EventHandler) *Router {
	if eventHandler == nil {
		eventHandler = func(RequestEvent) {}
	}
	return &Router{
		routes:       make(map[string]http.Handler),
		log:          log,
		eventHandler: eventHandler,
	}
}

// SetEventHandler sets the event handler after creation (avoids circular deps).
func (r *Router) SetEventHandler(h EventHandler) {
	r.eventHandler = h
}

// AddRoute registers a route from a RouteData struct.
func (r *Router) AddRoute(rd RouteData) error {
	handler, err := r.buildHandler(rd)
	if err != nil {
		return fmt.Errorf("building handler for %q: %w", rd.Host, err)
	}
	r.mu.Lock()
	r.routes[rd.Host] = handler
	// Also register without/with www if redirect is configured.
	if rd.WWWRedirect == "strip" {
		r.routes["www."+rd.Host] = handler
	} else if rd.WWWRedirect == "add" {
		bare := strings.TrimPrefix(rd.Host, "www.")
		if bare != rd.Host {
			r.routes[bare] = handler
		}
	}
	r.mu.Unlock()
	r.log.Info("route registered", "host", rd.Host, "target", rd.Target,
		"paths", len(rd.Paths), "static", rd.StaticRoot != "")
	return nil
}

// RemoveRoute removes a route by host.
func (r *Router) RemoveRoute(host string) {
	r.mu.Lock()
	delete(r.routes, host)
	delete(r.routes, "www."+host)
	r.mu.Unlock()
	r.log.Info("route removed", "host", host)
}

// ServeHTTP dispatches to the correct backend.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Strip port from host for matching.
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	r.mu.RLock()
	handler, ok := r.routes[host]
	r.mu.RUnlock()

	if !ok {
		r.log.Warn("no route for host", "host", host)
		http.Error(w, "502 Bad Gateway — no route configured for this host", http.StatusBadGateway)
		return
	}

	rec := newStatusRecorder(w)
	handler.ServeHTTP(rec, req)

	r.eventHandler(RequestEvent{
		Timestamp: start,
		Host:      host,
		Method:    req.Method,
		Path:      req.URL.Path,
		Status:    rec.status,
		LatencyMs: time.Since(start).Milliseconds(),
		ClientIP:  req.RemoteAddr,
	})
}

// buildHandler constructs the full per-route handler chain.
func (r *Router) buildHandler(rd RouteData) (http.Handler, error) {
	// Build the core handler (path router + static + proxy).
	core, err := r.buildCoreHandler(rd)
	if err != nil {
		return nil, err
	}

	// Wrap with per-route middleware (innermost first).
	var h http.Handler = core

	// Response headers injection (must be outermost to catch all responses).
	if len(rd.RespHeaders) > 0 {
		headers := rd.RespHeaders
		prev := h
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			prev.ServeHTTP(w, req)
		})
	}

	// Gzip compression.
	if rd.Gzip {
		h = gzipMiddleware(h)
	}

	// Per-route rate limiting.
	if rd.RateLimit.RequestsPerSecond > 0 {
		burst := rd.RateLimit.Burst
		if burst <= 0 {
			burst = int(rd.RateLimit.RequestsPerSecond) * 2
		}
		limiter := ratelimit.New(rd.RateLimit.RequestsPerSecond, burst)
		h = ratelimit.Middleware(limiter)(h)
	}

	// Per-route body size limit.
	if rd.MaxBodySize > 0 {
		maxBytes := rd.MaxBodySize
		prev := h
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.ContentLength > maxBytes {
				http.Error(w, "413 Request Too Large", http.StatusRequestEntityTooLarge)
				return
			}
			req.Body = http.MaxBytesReader(w, req.Body, maxBytes)
			prev.ServeHTTP(w, req)
		})
	}

	// www redirect (outermost).
	if rd.WWWRedirect != "" {
		h = wwwRedirectMiddleware(rd.WWWRedirect, rd.Host, h)
	}

	return h, nil
}

// buildCoreHandler builds the path-routing + static + proxy handler.
func (r *Router) buildCoreHandler(rd RouteData) (http.Handler, error) {
	// Sort paths longest-first for correct prefix matching.
	paths := make([]PathRule, len(rd.Paths))
	copy(paths, rd.Paths)
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i].Path) > len(paths[j].Path)
	})

	// Build proxy per target.
	proxies := map[string]*httputil.ReverseProxy{}

	buildProxy := func(target string) (*httputil.ReverseProxy, error) {
		if target == "" {
			return nil, nil
		}
		if p, ok := proxies[target]; ok {
			return p, nil
		}
		targetURL, err := url.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("parsing target %q: %w", target, err)
		}
		transport := defaultTransport
		if rd.TimeoutSecs > 0 {
			transport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: time.Duration(rd.TimeoutSecs) * time.Second,
				MaxIdleConns:          200,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       90 * time.Second,
			}
		}
		p := httputil.NewSingleHostReverseProxy(targetURL)
		p.Transport = transport
		p.FlushInterval = -1 // streaming/WebSocket support
		// Properly handle WebSocket and SSE upgrades.
		origDirector := p.Director
		p.Director = func(req *http.Request) {
			origDirector(req)
			// Preserve upgrade headers for WebSocket.
			if upgrade := req.Header.Get("Upgrade"); upgrade != "" {
				req.Header.Set("Connection", "upgrade")
				req.Header.Set("Upgrade", upgrade)
			}
		}
		p.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			r.log.Error("proxy backend error", "host", rd.Host, "target", target, "error", err)
			http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
		}
		proxies[target] = p
		return p, nil
	}

	// Build default proxy.
	defaultProxy, err := buildProxy(rd.Target)
	if err != nil {
		return nil, err
	}

	// Build per-path proxies.
	pathProxies := make([]*httputil.ReverseProxy, len(paths))
	for i, rule := range paths {
		target := rule.Target
		if target == "" {
			target = rd.Target
		}
		pp, err := buildProxy(target)
		if err != nil {
			return nil, err
		}
		pathProxies[i] = pp
	}

	// Static file handler.
	var staticHandler http.Handler
	if rd.StaticRoot != "" {
		var fs http.FileSystem = http.Dir(rd.StaticRoot)
		if rd.StaticSPA {
			fs = spaDir{http.Dir(rd.StaticRoot)}
		}
		staticHandler = http.FileServer(fs)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqPath := req.URL.Path

		// Check path rules first (longest match).
		for i, rule := range paths {
			if strings.HasPrefix(reqPath, rule.Path) {
				if pathProxies[i] != nil {
					pathProxies[i].ServeHTTP(w, req)
					return
				}
			}
		}

		// Static files.
		if staticHandler != nil {
			staticHandler.ServeHTTP(w, req)
			return
		}

		// Default proxy.
		if defaultProxy != nil {
			defaultProxy.ServeHTTP(w, req)
			return
		}

		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
	}), nil
}

// ── Middleware helpers ─────────────────────────────────────────────────────────

func wwwRedirectMiddleware(mode, host string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqHost := r.Host
		if h, _, err := net.SplitHostPort(reqHost); err == nil {
			reqHost = h
		}
		var redirect bool
		var target string
		if mode == "strip" && strings.HasPrefix(reqHost, "www.") {
			redirect = true
			target = strings.TrimPrefix(reqHost, "www.")
		} else if mode == "add" && !strings.HasPrefix(reqHost, "www.") {
			redirect = true
			target = "www." + reqHost
		}
		if redirect {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			http.Redirect(w, r, scheme+"://"+target+r.RequestURI, http.StatusMovedPermanently)
			return
		}
		_ = host
		next.ServeHTTP(w, r)
	})
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		r.Header.Del("Accept-Encoding")
		next.ServeHTTP(&gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	io.Writer
}

func (g *gzipWriter) Write(b []byte) (int, error) { return g.Writer.Write(b) }

// spaDir wraps http.Dir with SPA fallback (returns index.html for missing files).
type spaDir struct{ http.FileSystem }

func (s spaDir) Open(name string) (http.File, error) {
	f, err := s.FileSystem.Open(name)
	if err != nil {
		return s.FileSystem.Open("/index.html")
	}
	return f, nil
}

// statusRecorder captures the HTTP status code.
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
