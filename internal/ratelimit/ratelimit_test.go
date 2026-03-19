package ratelimit_test

import (
	"testing"

	"github.com/Elchi-dev/onyx/internal/ratelimit"
)

func TestAllowAndDeny(t *testing.T) {
	// Burst of 3 — first 3 should pass, 4th should be denied.
	l := ratelimit.New(0, 3) // rate=0 so no refill during test
	for i := 0; i < 3; i++ {
		if !l.Allow("client-a") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
	if l.Allow("client-a") {
		t.Error("4th request should be denied")
	}
}

func TestSeparateKeys(t *testing.T) {
	l := ratelimit.New(0, 1)
	if !l.Allow("client-a") {
		t.Error("client-a first request should be allowed")
	}
	if !l.Allow("client-b") {
		t.Error("client-b first request should be allowed (separate bucket)")
	}
}
