// Package database manages the Onyx SQLite database.
package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB and exposes Onyx-specific operations.
type DB struct {
	sql *sql.DB
}

// Open opens (or creates) the Onyx SQLite database at dbPath.
func Open(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database at %q: %w", dbPath, err)
	}
	if _, err := sqlDB.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return nil, fmt.Errorf("configuring SQLite: %w", err)
	}
	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return db, nil
}

func (db *DB) Close() error { return db.sql.Close() }

func Exists(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return err == nil
}

func (db *DB) migrate() error {
	for i, m := range migrations {
		if _, err := db.sql.Exec(m); err != nil {
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}

var migrations = []string{
	// 1 — settings
	`CREATE TABLE IF NOT EXISTS settings (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 2 — routes
	`CREATE TABLE IF NOT EXISTS routes (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		host       TEXT NOT NULL UNIQUE,
		target     TEXT NOT NULL,
		enabled    INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 3 — sessions
	`CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 4 — route stats
	`CREATE TABLE IF NOT EXISTS route_stats (
		host         TEXT PRIMARY KEY,
		total        INTEGER NOT NULL DEFAULT 0,
		errors       INTEGER NOT NULL DEFAULT 0,
		total_ms     INTEGER NOT NULL DEFAULT 0,
		last_seen_at DATETIME,
		updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 5-13 — v0.2/v0.3 route columns (all safe to re-run)
	`ALTER TABLE routes ADD COLUMN https INTEGER NOT NULL DEFAULT 0;`,
	`ALTER TABLE routes ADD COLUMN www_redirect TEXT NOT NULL DEFAULT '';`,
	`ALTER TABLE routes ADD COLUMN gzip INTEGER NOT NULL DEFAULT 0;`,
	`ALTER TABLE routes ADD COLUMN max_body_size INTEGER NOT NULL DEFAULT 0;`,
	`ALTER TABLE routes ADD COLUMN timeout_secs INTEGER NOT NULL DEFAULT 0;`,
	`ALTER TABLE routes ADD COLUMN paths_json TEXT NOT NULL DEFAULT '[]';`,
	`ALTER TABLE routes ADD COLUMN resp_headers_json TEXT NOT NULL DEFAULT '{}';`,
	`ALTER TABLE routes ADD COLUMN static_root TEXT NOT NULL DEFAULT '';`,
	`ALTER TABLE routes ADD COLUMN static_spa INTEGER NOT NULL DEFAULT 0;`,
}

// ── Settings ──────────────────────────────────────────────────────────────────

func (db *DB) SetSetting(key, value string) error {
	_, err := db.sql.Exec(
		`INSERT INTO settings(key,value,updated_at) VALUES(?,?,CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`,
		key, value,
	)
	return err
}

func (db *DB) GetSetting(key string) (string, bool, error) {
	var value string
	err := db.sql.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// ── Routes ────────────────────────────────────────────────────────────────────

// PathEntry is a path-based routing rule stored in the DB.
type PathEntry struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

// Route is a stored proxy route with all configuration.
type Route struct {
	ID          int64
	Host        string
	Target      string
	Enabled     bool
	HTTPS       bool
	WWWRedirect string
	Gzip        bool
	MaxBodySize int64
	TimeoutSecs int
	Paths       []PathEntry
	RespHeaders map[string]string
	StaticRoot  string
	StaticSPA   bool
}

// UpsertRoute inserts or replaces a route record.
func (db *DB) UpsertRoute(r Route) error {
	pathsJSON, _ := json.Marshal(r.Paths)
	headersJSON, _ := json.Marshal(r.RespHeaders)
	_, err := db.sql.Exec(`
		INSERT INTO routes(
			host,target,enabled,https,www_redirect,gzip,max_body_size,timeout_secs,
			paths_json,resp_headers_json,static_root,static_spa,updated_at
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
		ON CONFLICT(host) DO UPDATE SET
			target=excluded.target, enabled=excluded.enabled,
			https=excluded.https, www_redirect=excluded.www_redirect,
			gzip=excluded.gzip, max_body_size=excluded.max_body_size,
			timeout_secs=excluded.timeout_secs, paths_json=excluded.paths_json,
			resp_headers_json=excluded.resp_headers_json,
			static_root=excluded.static_root, static_spa=excluded.static_spa,
			updated_at=CURRENT_TIMESTAMP`,
		r.Host, r.Target, boolToInt(r.Enabled), boolToInt(r.HTTPS),
		r.WWWRedirect, boolToInt(r.Gzip), r.MaxBodySize, r.TimeoutSecs,
		string(pathsJSON), string(headersJSON),
		r.StaticRoot, boolToInt(r.StaticSPA),
	)
	return err
}

// SetRouteEnabled enables or disables a route.
func (db *DB) SetRouteEnabled(host string, enabled bool) error {
	_, err := db.sql.Exec(
		`UPDATE routes SET enabled=?, updated_at=CURRENT_TIMESTAMP WHERE host=?`,
		boolToInt(enabled), host,
	)
	return err
}

// SetRouteHTTPS toggles HTTPS for a route.
func (db *DB) SetRouteHTTPS(host string, https bool) error {
	_, err := db.sql.Exec(
		`UPDATE routes SET https=?, updated_at=CURRENT_TIMESTAMP WHERE host=?`,
		boolToInt(https), host,
	)
	return err
}

// ListRoutes returns all stored routes.
func (db *DB) ListRoutes() ([]Route, error) {
	rows, err := db.sql.Query(`
		SELECT id,host,target,enabled,
		       COALESCE(https,0), COALESCE(www_redirect,''), COALESCE(gzip,0),
		       COALESCE(max_body_size,0), COALESCE(timeout_secs,0),
		       COALESCE(paths_json,'[]'), COALESCE(resp_headers_json,'{}'),
		       COALESCE(static_root,''), COALESCE(static_spa,0)
		FROM routes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var routes []Route
	for rows.Next() {
		var r Route
		var en, ht, gz, spa int
		var pathsJSON, headersJSON string
		if err := rows.Scan(
			&r.ID, &r.Host, &r.Target, &en,
			&ht, &r.WWWRedirect, &gz,
			&r.MaxBodySize, &r.TimeoutSecs,
			&pathsJSON, &headersJSON,
			&r.StaticRoot, &spa,
		); err != nil {
			return nil, err
		}
		r.Enabled = en == 1
		r.HTTPS = ht == 1
		r.Gzip = gz == 1
		r.StaticSPA = spa == 1
		_ = json.Unmarshal([]byte(pathsJSON), &r.Paths)
		_ = json.Unmarshal([]byte(headersJSON), &r.RespHeaders)
		if r.Paths == nil {
			r.Paths = []PathEntry{}
		}
		if r.RespHeaders == nil {
			r.RespHeaders = map[string]string{}
		}
		routes = append(routes, r)
	}
	return routes, rows.Err()
}

// DeleteRoute removes a route by host.
func (db *DB) DeleteRoute(host string) error {
	_, err := db.sql.Exec(`DELETE FROM routes WHERE host=?`, host)
	return err
}

// ── Sessions ──────────────────────────────────────────────────────────────────

func (db *DB) SaveSession(token string, expiresAt time.Time) error {
	_, err := db.sql.Exec(
		`INSERT INTO sessions(token,expires_at) VALUES(?,?)
		 ON CONFLICT(token) DO UPDATE SET expires_at=excluded.expires_at`,
		token, expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (db *DB) ValidateSession(token string) (bool, time.Time, error) {
	var expiresStr string
	err := db.sql.QueryRow(`SELECT expires_at FROM sessions WHERE token=?`, token).Scan(&expiresStr)
	if err == sql.ErrNoRows {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, time.Time{}, err
	}
	expires, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return false, time.Time{}, err
	}
	if time.Now().After(expires) {
		_ = db.DeleteSession(token)
		return false, time.Time{}, nil
	}
	return true, expires, nil
}

func (db *DB) DeleteSession(token string) error {
	_, err := db.sql.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

func (db *DB) PurgeExpiredSessions() error {
	_, err := db.sql.Exec(`DELETE FROM sessions WHERE expires_at < ?`,
		time.Now().UTC().Format(time.RFC3339))
	return err
}

// ── Route Stats ───────────────────────────────────────────────────────────────

type RouteStats struct {
	Host       string
	Total      int64
	Errors     int64
	TotalMS    int64
	AvgLatency float64
	LastSeen   time.Time
}

func (db *DB) RecordRequest(host string, statusCode int, latencyMs int64) error {
	isError := 0
	if statusCode >= 500 {
		isError = 1
	}
	_, err := db.sql.Exec(`
		INSERT INTO route_stats(host,total,errors,total_ms,last_seen_at,updated_at)
		VALUES(?,1,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
		ON CONFLICT(host) DO UPDATE SET
			total=total+1, errors=errors+excluded.errors,
			total_ms=total_ms+excluded.total_ms,
			last_seen_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP`,
		host, isError, latencyMs,
	)
	return err
}

func (db *DB) ListRouteStats() ([]RouteStats, error) {
	rows, err := db.sql.Query(`
		SELECT host,total,errors,total_ms,COALESCE(last_seen_at,'')
		FROM route_stats ORDER BY total DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []RouteStats
	for rows.Next() {
		var s RouteStats
		var lastSeen string
		if err := rows.Scan(&s.Host, &s.Total, &s.Errors, &s.TotalMS, &lastSeen); err != nil {
			return nil, err
		}
		if s.Total > 0 {
			s.AvgLatency = float64(s.TotalMS) / float64(s.Total)
		}
		if lastSeen != "" {
			s.LastSeen, _ = time.Parse("2006-01-02 15:04:05", lastSeen)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

type GlobalStats struct {
	TotalRequests int64
	TotalErrors   int64
	AvgLatency    float64
	RouteCount    int
}

func (db *DB) GetGlobalStats() (GlobalStats, error) {
	var gs GlobalStats
	err := db.sql.QueryRow(`
		SELECT COALESCE(SUM(total),0), COALESCE(SUM(errors),0),
		       COALESCE(CAST(SUM(total_ms) AS REAL)/NULLIF(SUM(total),0),0)
		FROM route_stats`).Scan(&gs.TotalRequests, &gs.TotalErrors, &gs.AvgLatency)
	if err != nil {
		return gs, err
	}
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM routes`).Scan(&gs.RouteCount); err != nil {
		return gs, err
	}
	return gs, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
