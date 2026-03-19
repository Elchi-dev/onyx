// Package database manages the Onyx SQLite database.
// It provides connection setup, schema migrations, and typed data access methods.
// All other packages interact with the database only through this package's API.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver — no CGo, no external libraries.
)

// DB wraps sql.DB and exposes Onyx-specific operations.
type DB struct {
	sql *sql.DB
}

// Open opens (or creates) the Onyx SQLite database at dbPath.
// It configures WAL mode, enables foreign keys, and runs all migrations.
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

// Close closes the database connection.
func (db *DB) Close() error { return db.sql.Close() }

// Exists reports whether the database file already exists at dbPath.
func Exists(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return err == nil
}

// migrate applies all schema migrations in order.
// Migrations are append-only — never modify an existing entry.
func (db *DB) migrate() error {
	for i, m := range migrations {
		if _, err := db.sql.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}

var migrations = []string{
	// 1 — key-value settings store
	`CREATE TABLE IF NOT EXISTS settings (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 2 — proxy routes
	`CREATE TABLE IF NOT EXISTS routes (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		host       TEXT NOT NULL UNIQUE,
		target     TEXT NOT NULL,
		enabled    INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 3 — persistent sessions (replaces in-memory map)
	`CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	// 4 — per-route request statistics
	`CREATE TABLE IF NOT EXISTS route_stats (
		host         TEXT PRIMARY KEY,
		total        INTEGER NOT NULL DEFAULT 0,
		errors       INTEGER NOT NULL DEFAULT 0,
		total_ms     INTEGER NOT NULL DEFAULT 0,
		last_seen_at DATETIME,
		updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
}

// ── Settings ──────────────────────────────────────────────────────────────────

// SetSetting stores or updates a key-value pair.
func (db *DB) SetSetting(key, value string) error {
	_, err := db.sql.Exec(
		`INSERT INTO settings(key,value,updated_at) VALUES(?,?,CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// GetSetting retrieves a value by key.
// Returns ("", false, nil) when the key does not exist.
func (db *DB) GetSetting(key string) (value string, found bool, err error) {
	e := db.sql.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&value)
	if e == sql.ErrNoRows {
		return "", false, nil
	}
	if e != nil {
		return "", false, fmt.Errorf("get setting %q: %w", key, e)
	}
	return value, true, nil
}

// ── Routes ────────────────────────────────────────────────────────────────────

// Route is a stored proxy route record.
type Route struct {
	ID      int64
	Host    string
	Target  string
	Enabled bool
}

// UpsertRoute inserts or updates a route record.
func (db *DB) UpsertRoute(host, target string, enabled bool) error {
	_, err := db.sql.Exec(
		`INSERT INTO routes(host,target,enabled,updated_at) VALUES(?,?,?,CURRENT_TIMESTAMP)
		 ON CONFLICT(host) DO UPDATE SET target=excluded.target, enabled=excluded.enabled,
		 updated_at=CURRENT_TIMESTAMP`,
		host, target, boolToInt(enabled),
	)
	if err != nil {
		return fmt.Errorf("upsert route %q: %w", host, err)
	}
	return nil
}

// SetRouteEnabled enables or disables a route without changing its target.
func (db *DB) SetRouteEnabled(host string, enabled bool) error {
	_, err := db.sql.Exec(
		`UPDATE routes SET enabled=?, updated_at=CURRENT_TIMESTAMP WHERE host=?`,
		boolToInt(enabled), host,
	)
	if err != nil {
		return fmt.Errorf("set route enabled %q: %w", host, err)
	}
	return nil
}

// ListRoutes returns all stored routes ordered by insertion time.
func (db *DB) ListRoutes() ([]Route, error) {
	rows, err := db.sql.Query(`SELECT id,host,target,enabled FROM routes ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	defer rows.Close()
	var routes []Route
	for rows.Next() {
		var r Route
		var e int
		if err := rows.Scan(&r.ID, &r.Host, &r.Target, &e); err != nil {
			return nil, err
		}
		r.Enabled = e == 1
		routes = append(routes, r)
	}
	return routes, rows.Err()
}

// DeleteRoute removes a route by host. Silently succeeds if not found.
func (db *DB) DeleteRoute(host string) error {
	_, err := db.sql.Exec(`DELETE FROM routes WHERE host=?`, host)
	if err != nil {
		return fmt.Errorf("delete route %q: %w", host, err)
	}
	return nil
}

// ── Sessions ──────────────────────────────────────────────────────────────────

// SaveSession persists a session token with its expiry time.
func (db *DB) SaveSession(token string, expiresAt time.Time) error {
	_, err := db.sql.Exec(
		`INSERT INTO sessions(token,expires_at) VALUES(?,?)
		 ON CONFLICT(token) DO UPDATE SET expires_at=excluded.expires_at`,
		token, expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

// ValidateSession checks if a token exists and is not expired.
// Returns true and the expiry time if valid.
func (db *DB) ValidateSession(token string) (bool, time.Time, error) {
	var expiresStr string
	err := db.sql.QueryRow(
		`SELECT expires_at FROM sessions WHERE token=?`, token,
	).Scan(&expiresStr)
	if err == sql.ErrNoRows {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, time.Time{}, fmt.Errorf("validate session: %w", err)
	}
	expires, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("parsing session expiry: %w", err)
	}
	if time.Now().After(expires) {
		_ = db.DeleteSession(token)
		return false, time.Time{}, nil
	}
	return true, expires, nil
}

// DeleteSession removes a single session token.
func (db *DB) DeleteSession(token string) error {
	_, err := db.sql.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

// PurgeExpiredSessions removes all sessions whose expiry has passed.
// Call this periodically (e.g. every hour) to keep the table small.
func (db *DB) PurgeExpiredSessions() error {
	_, err := db.sql.Exec(
		`DELETE FROM sessions WHERE expires_at < ?`,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ── Route Stats ───────────────────────────────────────────────────────────────

// RouteStats holds aggregated request statistics for one route.
type RouteStats struct {
	Host       string
	Total      int64
	Errors     int64
	TotalMS    int64
	AvgLatency float64
	LastSeen   time.Time
}

// RecordRequest atomically increments statistics for the given host.
func (db *DB) RecordRequest(host string, statusCode int, latencyMs int64) error {
	isError := 0
	if statusCode >= 500 {
		isError = 1
	}
	_, err := db.sql.Exec(`
		INSERT INTO route_stats(host, total, errors, total_ms, last_seen_at, updated_at)
		VALUES(?, 1, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(host) DO UPDATE SET
			total      = total + 1,
			errors     = errors + excluded.errors,
			total_ms   = total_ms + excluded.total_ms,
			last_seen_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP`,
		host, isError, latencyMs,
	)
	if err != nil {
		return fmt.Errorf("record request for %q: %w", host, err)
	}
	return nil
}

// ListRouteStats returns aggregated stats for all routes.
func (db *DB) ListRouteStats() ([]RouteStats, error) {
	rows, err := db.sql.Query(`
		SELECT host, total, errors, total_ms,
		       COALESCE(last_seen_at,'') as last_seen
		FROM route_stats ORDER BY total DESC`)
	if err != nil {
		return nil, fmt.Errorf("list route stats: %w", err)
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

// GlobalStats holds aggregated totals across all routes.
type GlobalStats struct {
	TotalRequests int64
	TotalErrors   int64
	AvgLatency    float64
	RouteCount    int
}

// GetGlobalStats returns totals across all routes plus route count.
func (db *DB) GetGlobalStats() (GlobalStats, error) {
	var gs GlobalStats
	err := db.sql.QueryRow(`
		SELECT COALESCE(SUM(total),0), COALESCE(SUM(errors),0),
		       COALESCE(CAST(SUM(total_ms) AS REAL)/NULLIF(SUM(total),0), 0)
		FROM route_stats`).Scan(&gs.TotalRequests, &gs.TotalErrors, &gs.AvgLatency)
	if err != nil {
		return gs, fmt.Errorf("global stats: %w", err)
	}
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM routes`).Scan(&gs.RouteCount); err != nil {
		return gs, fmt.Errorf("route count: %w", err)
	}
	return gs, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
