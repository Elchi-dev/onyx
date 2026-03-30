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
	"github.com/Elchi-dev/onyx/internal/nginx"
	"github.com/Elchi-dev/onyx/internal/proxy"
	tlsmgr "github.com/Elchi-dev/onyx/internal/tls"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkSameOrigin,
}

func checkSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	host := r.Host
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")
	return origin == host
}

// ProxyRequestEvent is sent to dashboard clients via WebSocket.
type ProxyRequestEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	ClientIP  string    `json:"client_ip"`
}

// RouteManager is the subset of proxy.Router used by the dashboard.
type RouteManager interface {
	AddRoute(proxy.RouteData) error
	RemoveRoute(string)
}

// BroadcastRequest records and broadcasts a proxy request event.
func (d *Dashboard) BroadcastRequest(e proxy.RequestEvent) {
	d.stats.record()
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
	tls        *tlsmgr.Manager
	router     RouteManager
	mux        *http.ServeMux
	loginLimit *loginLimiter
	stats      *serverStats
	version    string
	startTime  time.Time
}

// New creates a Dashboard.
func New(log *slog.Logger, db *database.DB, tls *tlsmgr.Manager) *Dashboard {
	d := &Dashboard{
		log:        log,
		hub:        NewHub(log),
		db:         db,
		tls:        tls,
		mux:        http.NewServeMux(),
		loginLimit: newLoginLimiter(5, time.Minute),
		stats:      newServerStats(),
		startTime:  time.Now(),
	}
	d.registerRoutes()
	go d.sessionCleanup()
	return d
}

// SetVersion stores the version string.
func (d *Dashboard) SetVersion(v string) { d.version = v }

// SetRouter wires the proxy router so the dashboard can add/remove routes live.
func (d *Dashboard) SetRouter(r RouteManager) { d.router = r }

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func (d *Dashboard) registerRoutes() {
	d.mux.HandleFunc("/", d.handleIndex)
	d.mux.HandleFunc("/login", d.handleLogin)
	d.mux.HandleFunc("/logout", d.handleLogout)
	d.mux.HandleFunc("/ws", d.requireAuth(d.handleWebSocket))

	d.mux.HandleFunc("/api/routes", d.requireAuth(d.handleRoutesAPI))
	d.mux.HandleFunc("/api/routes/", d.requireAuth(d.handleRouteByHost))
	d.mux.HandleFunc("/api/stats", d.requireAuth(d.handleStatsAPI))
	d.mux.HandleFunc("/api/about", d.requireAuth(d.handleAboutAPI))
	d.mux.HandleFunc("/api/certs", d.requireAuth(d.handleCertsAPI))
	d.mux.HandleFunc("/api/settings/password", d.requireAuth(d.handleChangePassword))
	d.mux.HandleFunc("/api/import/nginx", d.requireAuth(d.handleImportNginx))
}

// ── Page handlers ─────────────────────────────────────────────────────────────

func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	cookie, err := r.Cookie("onyx_session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	valid, _, err := d.db.ValidateSession(cookie.Value)
	if err != nil || !valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
	if !d.loginLimit.allow(r) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(loginRateLimitHTML))
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
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "onyx_session", Value: session.Token,
		Expires: session.ExpiresAt, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Path: "/",
	})
	d.log.Info("successful login", "remote", r.RemoteAddr, "remember_me", rememberMe)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Dashboard) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("onyx_session"); err == nil {
		_ = d.db.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "onyx_session", Value: "", Expires: time.Unix(0, 0), Path: "/"})
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

// routeAPIPayload is the JSON shape for creating/updating routes.
type routeAPIPayload struct {
	Host        string            `json:"host"`
	Target      string            `json:"target"`
	HTTPS       bool              `json:"https"`
	WWWRedirect string            `json:"www_redirect"`
	Gzip        bool              `json:"gzip"`
	MaxBodySize int64             `json:"max_body_size"`
	TimeoutSecs int               `json:"timeout_secs"`
	StaticRoot  string            `json:"static_root"`
	StaticSPA   bool              `json:"static_spa"`
	RespHeaders map[string]string `json:"headers"`
	Paths       []struct {
		Path   string `json:"path"`
		Target string `json:"target"`
	} `json:"paths"`
	// PATCH-only fields.
	Enabled *bool `json:"enabled,omitempty"`
}

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
		var body routeAPIPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Host == "" || (body.Target == "" && body.StaticRoot == "") {
			jsonError(w, "host and target (or static_root) are required", http.StatusBadRequest)
			return
		}
		dbRoute := payloadToDBRoute(body)
		dbRoute.Enabled = true
		if err := d.db.UpsertRoute(dbRoute); err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		if d.router != nil {
			_ = d.router.AddRoute(dbToRouteData(dbRoute))
		}
		if body.HTTPS && d.tls != nil {
			d.tls.AddHost(body.Host)
		}
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]string{"status": "created", "host": body.Host})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

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
		if d.router != nil {
			d.router.RemoveRoute(host)
		}
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]string{"status": "deleted", "host": host})

	case http.MethodPatch:
		var body routeAPIPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		// Load existing route.
		routes, err := d.db.ListRoutes()
		if err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		var existing database.Route
		found := false
		for _, rt := range routes {
			if rt.Host == host {
				existing = rt
				found = true
				break
			}
		}
		if !found {
			jsonError(w, "route not found", http.StatusNotFound)
			return
		}
		// Apply changes.
		if body.Enabled != nil {
			existing.Enabled = *body.Enabled
		}
		if body.Target != "" {
			existing.Target = body.Target
		}
		existing.HTTPS = body.HTTPS
		existing.WWWRedirect = body.WWWRedirect
		existing.Gzip = body.Gzip
		if body.MaxBodySize > 0 {
			existing.MaxBodySize = body.MaxBodySize
		}
		if body.TimeoutSecs > 0 {
			existing.TimeoutSecs = body.TimeoutSecs
		}
		if body.StaticRoot != "" {
			existing.StaticRoot = body.StaticRoot
		}
		existing.StaticSPA = body.StaticSPA
		if body.RespHeaders != nil {
			existing.RespHeaders = body.RespHeaders
		}
		if body.Paths != nil {
			existing.Paths = []database.PathEntry{}
			for _, p := range body.Paths {
				existing.Paths = append(existing.Paths, database.PathEntry{Path: p.Path, Target: p.Target})
			}
		}
		if err := d.db.UpsertRoute(existing); err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		if d.router != nil && existing.Enabled {
			_ = d.router.AddRoute(dbToRouteData(existing))
		} else if d.router != nil {
			d.router.RemoveRoute(host)
		}
		if existing.HTTPS && d.tls != nil {
			d.tls.AddHost(host)
		}
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
		jsonOK(w, map[string]string{"status": "updated", "host": host})

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
		"global":      global,
		"per_route":   perRoute,
		"uptime":      d.stats.uptimeString(),
		"req_per_min": d.stats.reqPerMin(),
	})
}

func (d *Dashboard) handleAboutAPI(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"version":    d.version,
		"uptime":     d.stats.uptimeString(),
		"start_time": d.startTime.Format(time.RFC3339),
	})
}

func (d *Dashboard) handleCertsAPI(w http.ResponseWriter, r *http.Request) {
	if d.tls == nil {
		jsonOK(w, []any{})
		return
	}
	jsonOK(w, d.tls.CertStatuses())
}

func (d *Dashboard) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(body.New) < 8 {
		jsonError(w, "new password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	hash, ok, err := d.db.GetSetting(auth.SettingKeyPasswordHash)
	if err != nil || !ok {
		jsonError(w, "server error", http.StatusInternalServerError)
		return
	}
	if !auth.CheckPassword(hash, body.Current) {
		jsonError(w, "current password is incorrect", http.StatusUnauthorized)
		return
	}
	newHash, err := auth.HashPassword(body.New)
	if err != nil {
		jsonError(w, "server error", http.StatusInternalServerError)
		return
	}
	if err := d.db.SetSetting(auth.SettingKeyPasswordHash, newHash); err != nil {
		jsonError(w, "server error", http.StatusInternalServerError)
		return
	}
	d.log.Info("dashboard password changed")
	jsonOK(w, map[string]string{"status": "ok"})
}

func (d *Dashboard) handleImportNginx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Config == "" {
		jsonError(w, "config is required", http.StatusBadRequest)
		return
	}
	routes, err := nginx.ParseConfig(body.Config)
	if err != nil {
		jsonError(w, "parse error: "+err.Error(), http.StatusBadRequest)
		return
	}
	imported := 0
	var warnings []string
	for _, route := range routes {
		if err := d.db.UpsertRoute(route); err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to save %s: %v", route.Host, err))
			continue
		}
		if d.router != nil && route.Enabled {
			_ = d.router.AddRoute(dbToRouteData(route))
		}
		imported++
	}
	if imported > 0 {
		d.hub.Broadcast(Event{Type: "routes_changed", Payload: nil})
	}
	jsonOK(w, map[string]any{
		"imported": imported,
		"total":    len(routes),
		"warnings": warnings,
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

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ws"
}

func (d *Dashboard) sessionCleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if err := d.db.PurgeExpiredSessions(); err != nil {
			d.log.Warn("purging expired sessions", "error", err)
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

// ── Conversion helpers ────────────────────────────────────────────────────────

func payloadToDBRoute(b routeAPIPayload) database.Route {
	r := database.Route{
		Host:        b.Host,
		Target:      b.Target,
		HTTPS:       b.HTTPS,
		WWWRedirect: b.WWWRedirect,
		Gzip:        b.Gzip,
		MaxBodySize: b.MaxBodySize,
		TimeoutSecs: b.TimeoutSecs,
		StaticRoot:  b.StaticRoot,
		StaticSPA:   b.StaticSPA,
		RespHeaders: b.RespHeaders,
	}
	if r.RespHeaders == nil {
		r.RespHeaders = map[string]string{}
	}
	for _, p := range b.Paths {
		r.Paths = append(r.Paths, database.PathEntry{Path: p.Path, Target: p.Target})
	}
	if r.Paths == nil {
		r.Paths = []database.PathEntry{}
	}
	return r
}

func dbToRouteData(r database.Route) proxy.RouteData {
	rd := proxy.RouteData{
		Host:        r.Host,
		Target:      r.Target,
		HTTPS:       r.HTTPS,
		WWWRedirect: r.WWWRedirect,
		Gzip:        r.Gzip,
		MaxBodySize: r.MaxBodySize,
		TimeoutSecs: r.TimeoutSecs,
		StaticRoot:  r.StaticRoot,
		StaticSPA:   r.StaticSPA,
		RespHeaders: r.RespHeaders,
	}
	for _, p := range r.Paths {
		rd.Paths = append(rd.Paths, proxy.PathRule{Path: p.Path, Target: p.Target})
	}
	return rd
}
