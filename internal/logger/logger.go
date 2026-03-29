// Package logger provides a structured, leveled logger for Onyx.
// It wraps slog so all packages share one consistent format without
// introducing an external logging dependency.
package logger

import (
	"log/slog"
	"os"
)

// New returns a slog.Logger. Development mode uses human-readable text;
// production uses JSON for log aggregators.
func New(level slog.Level, development bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if development {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

// Default returns a text logger at debug level, suitable for development.
func Default() *slog.Logger {
	return New(slog.LevelDebug, true)
}
