package config_test

import (
	"os"
	"testing"

	"github.com/Elchi-dev/onyx/internal/config"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()
	if cfg.Server.HTTPPort != 80 {
		t.Errorf("want HTTPPort=80, got %d", cfg.Server.HTTPPort)
	}
	if !cfg.Dashboard.Enabled {
		t.Error("want dashboard enabled by default")
	}
}

func TestLoadAndWrite(t *testing.T) {
	const content = `
[server]
http_port = 9090
data_dir  = "/tmp/onyx-test"

[dashboard]
enabled = true
port    = 9091

[[routes]]
host    = "test.local"
target  = "http://localhost:3000"
enabled = true
`
	f, err := os.CreateTemp("", "onyx-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.HTTPPort != 9090 {
		t.Errorf("want 9090, got %d", cfg.Server.HTTPPort)
	}
	if len(cfg.Routes) != 1 || cfg.Routes[0].Host != "test.local" {
		t.Errorf("unexpected routes: %+v", cfg.Routes)
	}
}

func TestValidateEmptyHost(t *testing.T) {
	cfg := config.Defaults()
	cfg.Routes = append(cfg.Routes, config.RouteConfig{Host: "", Target: "http://x"})
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty host")
	}
}
