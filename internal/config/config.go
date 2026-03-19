// Package config handles loading, validating, and writing the Onyx TOML config.
// It is the single source of truth for all configuration values.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the root configuration structure for Onyx.
type Config struct {
	Server    ServerConfig    `toml:"server"`
	Dashboard DashboardConfig `toml:"dashboard"`
	Routes    []RouteConfig   `toml:"routes"`
}

// ServerConfig holds core proxy listener settings.
type ServerConfig struct {
	HTTPPort  int    `toml:"http_port"`
	HTTPSPort int    `toml:"https_port"`
	DataDir   string `toml:"data_dir"`
}

// DashboardConfig holds settings for the live WebSocket dashboard.
type DashboardConfig struct {
	Enabled bool `toml:"enabled"`
	Port    int  `toml:"port"`
}

// RouteConfig defines one proxy route: an incoming hostname mapped to a backend.
type RouteConfig struct {
	Host      string          `toml:"host"`
	Target    string          `toml:"target"`
	Enabled   bool            `toml:"enabled"`
	RateLimit RateLimitConfig `toml:"rate_limit"`
}

// RateLimitConfig holds optional rate limiting settings for a single route.
type RateLimitConfig struct {
	RequestsPerSecond float64 `toml:"requests_per_second"`
	Burst             int     `toml:"burst"`
}

// Defaults returns a Config with sane default values pre-filled.
func Defaults() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort:  80,
			HTTPSPort: 443,
			DataDir:   defaultDataDir(),
		},
		Dashboard: DashboardConfig{
			Enabled: true,
			Port:    8080,
		},
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/etc/onyx"
	}
	return filepath.Join(home, ".config", "onyx")
}

// Load reads and parses the TOML config at path, layered on top of Defaults.
func Load(path string) (*Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// Write serialises cfg to TOML and writes it to path, creating directories as needed.
func (c *Config) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return nil
}

// Validate checks cfg for logical errors.
func (c *Config) Validate() error {
	if c.Server.HTTPPort < 1 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("server.http_port must be 1–65535")
	}
	for i, r := range c.Routes {
		if r.Host == "" {
			return fmt.Errorf("routes[%d]: host is required", i)
		}
		if r.Target == "" {
			return fmt.Errorf("routes[%d]: target is required", i)
		}
	}
	return nil
}

// DBPath returns the canonical path to the Onyx SQLite database.
func (c *Config) DBPath() string {
	return filepath.Join(c.Server.DataDir, "onyx.db")
}

// ConfigPath returns the canonical path to the Onyx config file.
func (c *Config) ConfigPath() string {
	return filepath.Join(c.Server.DataDir, "onyx.toml")
}
