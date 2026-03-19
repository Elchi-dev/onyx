package dashboard

import (
	"fmt"
	"sync/atomic"
	"time"
)

// serverStats tracks runtime statistics for the dashboard API.
type serverStats struct {
	startTime time.Time
	reqCount  atomic.Int64
}

// newServerStats creates a stats tracker starting from now.
func newServerStats() *serverStats {
	return &serverStats{startTime: time.Now()}
}

// record increments the in-process request counter.
func (s *serverStats) record() {
	s.reqCount.Add(1)
}

// uptime returns the duration since the server started.
func (s *serverStats) uptime() time.Duration {
	return time.Since(s.startTime)
}

// uptimeString returns a human-readable uptime string.
func (s *serverStats) uptimeString() string {
	d := s.uptime()
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
