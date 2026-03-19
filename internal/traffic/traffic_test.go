package traffic_test

import (
	"testing"

	"github.com/Elchi-dev/onyx/internal/traffic"
)

func TestNewSplitterWeightsMustSum100(t *testing.T) {
	_, err := traffic.NewSplitter([]traffic.Backend{
		{Target: "http://localhost:3001", Weight: 60},
		{Target: "http://localhost:3002", Weight: 30},
		// 60+30 = 90, not 100
	})
	if err == nil {
		t.Error("expected error when weights do not sum to 100")
	}
}

func TestNewSplitterValid(t *testing.T) {
	_, err := traffic.NewSplitter([]traffic.Backend{
		{Target: "http://localhost:3001", Weight: 70},
		{Target: "http://localhost:3002", Weight: 30},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
