// Package traffic implements weighted traffic splitting across multiple backends.
// Use it for canary deployments and A/B routing.
package traffic

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Backend is an upstream target with an associated percentage weight.
type Backend struct {
	Target string // e.g. "http://localhost:3000"
	Weight int    // 0-100; all weights in a Splitter must sum to 100
}

// Splitter distributes incoming requests across backends by weight.
type Splitter struct {
	backends []*httputil.ReverseProxy
	weights  []int
	total    int
}

// NewSplitter creates a Splitter from the given backends.
// Returns an error if weights do not sum to 100.
func NewSplitter(backends []Backend) (*Splitter, error) {
	total := 0
	for _, b := range backends {
		total += b.Weight
	}
	if total != 100 {
		return nil, fmt.Errorf("backend weights must sum to 100, got %d", total)
	}
	s := &Splitter{
		backends: make([]*httputil.ReverseProxy, len(backends)),
		weights:  make([]int, len(backends)),
		total:    total,
	}
	for i, b := range backends {
		u, err := url.Parse(b.Target)
		if err != nil {
			return nil, fmt.Errorf("parsing target %q: %w", b.Target, err)
		}
		s.backends[i] = httputil.NewSingleHostReverseProxy(u)
		s.weights[i] = b.Weight
	}
	return s, nil
}

// ServeHTTP selects a backend by weight and proxies the request to it.
func (s *Splitter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n := rand.IntN(s.total) //nolint:gosec // Non-cryptographic weighted selection.
	cumulative := 0
	for i, weight := range s.weights {
		cumulative += weight
		if n < cumulative {
			s.backends[i].ServeHTTP(w, r)
			return
		}
	}
	s.backends[0].ServeHTTP(w, r)
}
