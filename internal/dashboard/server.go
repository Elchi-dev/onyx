package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Elchi-dev/onyx/internal/auth"
	"github.com/Elchi-dev/onyx/internal/database"
	"github.com/Elchi-dev/onyx/internal/proxy"
)

// upgrader upgrades HTTP connections to WebSocket.
// CheckOrigin validates that the request comes from the same host as the dashboard
// to prevent cross-site WebSocket hijacking.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkSameOrigin,
}

// checkSameOrigin rejects WebSocket upgrades from a different origin.
// This prevents any website from silently opening a WebSocket to the dashboard.
func checkSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No origin header — allow (direct/curl access).
		return true
	}
	host := r.Host
	// Strip scheme from origin for comparison.
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")
	return origin == host
}

// ProxyRequestEvent is the JSON shape sent to dashboard clients over WebSocket.
type ProxyRequestEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	ClientIP  string    `json:"client_ip"`
}

// BroadcastRequest converts a proxy.RequestEvent, records it in the database,
// and broadcasts it to all connected dashboard clients.
func (d *Dashboard) BroadcastRequest(e proxy.RequestEvent) {
	d.stats.record()

	// Persist to database for server-side stats (non-blocking best-effort).
	if err := d.db.RecordRequest(e.Host, e.Status, e.LatencyMs); err != nil {
		d.log.Warn("recording request stat", "error", err)
	}

	d.hub.Broadcast(Event{
		Type: "request",
		Payload: ProxyRequestEvent{
			Timestamp: e.Timestamp,
			Host:      e.Host,
			Method:    e.Method,
			Path:      e.Path,
			Status:    e.Status,
			LatencyMs: e.LatencyMs,
			ClientIP:  e.ClientIP,
		},
	})
}

// Dashboard manages the dashboard HTTP and WebSocket server.
type Dashboard struct {
	log        *slog.Logger
	hub        *Hub
	db         *database.DB
	mux        *http.ServeMux
	loginLimit *loginLimiter
	stats      *serverStats
	version    string
	startTime  time.Time
}

// New creates a Dashboard, wires all routes, and starts background cleanup.
func New(log *slog.Logger, db *database.DB) *Dashboard {
	d := &Dashboard{
		log: log,
		hub: NewHub(log),
		db:  db,
		mux: http.NewServeMux(),
		// Max 5 login attempts per minute per IP.
		loginLimit: newLoginLimiter(5, time.Minute),
		stats:      newServerStats(),
		startTime:  time.Now(),
	}
	d.registerRoutes()
	go d.sessionCleanup()
	return d
}

// SetVersion stores the binary version string for the about API.
func (d *Dashboard) SetVersion(v string) { d.version = v }

// ServeHTTP implements http.Handler.
func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func (d *Dashboard) registerRoutes() {
	d.mux.HandleFunc("/", d.handleIndex)
	d.mux.HandleFunc("/login", d.handleLogin)
	d.mux.HandleFunc("/logout", d.handleLogout)
	d.mux.HandleFunc("/ws", d.requireAuth(d.handleWebSocket))

	// API — all JSON, all require auth.
	d.mux.HandleFunc("/api/routes", d.requireAuth(d.handleRoutesAPI))
	d.mux.HandleFunc("/api/routes/", d.requireAuth(d.handleRouteByHost))
	d.mux.HandleFunc("/api/stats", d.requireAuth(d.handleStatsAPI))
	d.mux.HandleFunc("/api/about", d.requireAuth(d.handleAboutAPI))
}

// ── Page handlers ─────────────────────────────────────────────────────────────

func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(dashboardHTML))
}

func (d *Dashboard) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(loginHTML))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limit: max 5 attempts per minute per IP.
	if !d.loginLimit.allow(r) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(loginRateLimitHTML))
		d.log.Warn("login rate limit exceeded", "remote", r.RemoteAddr)
		return
	}

	password := r.FormValue("password")
	rememberMe := r.FormValue("remember_me") == "on"

	hash, ok, err := d.db.GetSetting(auth.SettingKeyPasswordHash)
	if err != nil || !ok {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPassword(hash, password) {
		d.log.Warn("failed login attempt", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(loginErrorHTML))
		return
	}

	session, err := auth.NewSession(rememberMe)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if err := d.db.SaveSession(session.Token, session.ExpiresAt); err != nil {
		d.log.Error("saving session", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "onyx_session",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	d.log.Info("successful login", "remote", r.RemoteAddr, "remember_me", rememberMe)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Dashboard) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("onyx_session"); err == nil {
		_ = d.db.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "onyx_session",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (d *Dashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		d.log.Error("WebSocket upgrade failed", "error", err)
		return
	}
	d.hub.Register(conn)
}

// ── API handlers ──────────────────────────────────────────────────────────────

// handleRoutesAPI handles GET (list) and POST (create) for /api/routes.
func (d *Dashboard) handleRoutesAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		routes, err := d.db.ListRoutes()
		if err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		jsonOK(w, routes)

	case http.MethodPost:
		var body struct {
			Host   string `json:"host"`
			Target string `json:"target"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Host == "" || body.Target == "" {
			jsonError(w, "host and target are required", http.StatusBadRequest)
			return
		}
		if err := d.db.UpsertRoute(body.Host, body.Target, true); err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		// Broadcast a routes-changed event so connected clients refresh their list.
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]string{"status": "created", "host": body.Host})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// handleRouteByHost handles DELETE and PATCH for /api/routes/{host}.
func (d *Dashboard) handleRouteByHost(w http.ResponseWriter, r *http.Request) {
	host := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	if host == "" {
		jsonError(w, "host required in path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := d.db.DeleteRoute(host); err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]string{"status": "deleted", "host": host})

	case http.MethodPatch:
		var body struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Enabled == nil {
			jsonError(w, "enabled field required", http.StatusBadRequest)
			return
		}
		if err := d.db.SetRouteEnabled(host, *body.Enabled); err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]any{"status": "updated", "host": host, "enabled": *body.Enabled})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Dashboard) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	global, err := d.db.GetGlobalStats()
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	perRoute, err := d.db.ListRouteStats()
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"global":    global,
		"per_route": perRoute,
		"uptime":    d.stats.uptimeString(),
	})
}

func (d *Dashboard) handleAboutAPI(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"version":    d.version,
		"uptime":     d.stats.uptimeString(),
		"start_time": d.startTime.Format(time.RFC3339),
	})
}

// ── Auth middleware ───────────────────────────────────────────────────────────

func (d *Dashboard) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("onyx_session")
		if err != nil {
			if isAPIRequest(r) {
				jsonError(w, "unauthorized", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		valid, _, err := d.db.ValidateSession(cookie.Value)
		if err != nil || !valid {
			if isAPIRequest(r) {
				jsonError(w, "session expired", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		next(w, r)
	}
}

// isAPIRequest reports whether the request is to an /api/ endpoint.
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ws"
}

// ── Background tasks ──────────────────────────────────────────────────────────

// sessionCleanup purges expired sessions from the database every hour.
func (d *Dashboard) sessionCleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if err := d.db.PurgeExpiredSessions(); err != nil {
			d.log.Warn("purging expired sessions", "error", err)
		} else {
			d.log.Debug("purged expired sessions")
		}
	}
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// StartServer starts the dashboard HTTP server. Blocks until ctx is canceled.
func StartServer(ctx context.Context, addr string, handler http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("dashboard listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("dashboard server: %w", err)
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	}
}
