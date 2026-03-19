// Package proxy implements the Onyx core reverse proxy engine.
package proxy

import "time"

// RequestEvent describes a single proxied request.
// It is emitted after every request and picked up by the dashboard broadcaster.
type RequestEvent struct {
	Timestamp time.Time
	Host      string
	Method    string
	Path      string
	Status    int
	LatencyMs int64
	ClientIP  string
}

// EventHandler is a callback invoked after each proxied request.
// Implementations must be non-blocking.
type EventHandler func(RequestEvent)
