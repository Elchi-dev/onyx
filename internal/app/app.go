// Package app is the dependency wiring layer for Onyx.
// It owns the lifecycle of all major components and connects them together.
// CLI commands only need to know about app.App -- not individual internals.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Elchi-dev/onyx/internal/config"
	"github.com/Elchi-dev/onyx/internal/dashboard"
	"github.com/Elchi-dev/onyx/internal/database"
	"github.com/Elchi-dev/onyx/internal/logger"
	"github.com/Elchi-dev/onyx/internal/middleware"
	"github.com/Elchi-dev/onyx/internal/proxy"
	"github.com/Elchi-dev/onyx/internal/ratelimit"
)

// defaultBodyLimit is the maximum allowed request body size (10 MiB).
const defaultBodyLimit = 10 << 20

// App holds all wired components and orchestrates startup and shutdown.
type App struct {
	cfg     *config.Config
	log     *slog.Logger
	db      *database.DB
	dash    *dashboard.Dashboard
	router  *proxy.Router
	version string
}

// New wires all components from the given config and logger.
func New(cfg *config.Config, log *slog.Logger, version string) (*App, error) {
	db, err := database.Open(cfg.DBPath())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Create router first (with no event handler yet) so we can pass it to
	// the dashboard as the RouteManager. The event handler is wired after.
	router := proxy.New(log, nil)

	// Dashboard gets the router so it can add/remove routes live from the UI.
	dash := dashboard.New(log, db, router)
	dash.SetVersion(version)

	// Now wire the event handler: proxy events -> dashboard broadcast.
	router.SetEventHandler(dash.BroadcastRequest)

	// Load routes from TOML config.
	for _, r := range cfg.Routes {
		if r.Enabled {
			if err := router.AddRoute(r.Host, r.Target); err != nil {
				log.Warn("skipping invalid config route", "host", r.Host, "error", err)
			}
		}
	}

	// Load routes stored in the database (added via CLI or dashboard).
	dbRoutes, err := db.ListRoutes()
	if err != nil {
		return nil, fmt.Errorf("loading db routes: %w", err)
	}
	for _, r := range dbRoutes {
		if r.Enabled {
			_ = router.AddRoute(r.Host, r.Target)
		}
	}

	log.Info("routes loaded",
		"config", len(cfg.Routes),
		"database", len(dbRoutes),
	)

	return &App{cfg: cfg, log: log, db: db, dash: dash, router: router, version: version}, nil
}

// Run starts all servers and blocks until SIGTERM or SIGINT.
func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	proxyAddr := fmt.Sprintf(":%d", a.cfg.Server.HTTPPort)
	dashAddr := fmt.Sprintf(":%d", a.cfg.Dashboard.Port)

	a.log.Info("Onyx starting",
		slog.String("proxy", proxyAddr),
		slog.String("dashboard", dashAddr),
		slog.String("version", a.version),
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := proxy.StartServer(ctx, proxyAddr, a.proxyHandler(), a.log); err != nil {
			errCh <- err
		}
	}()

	if a.cfg.Dashboard.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := dashboard.StartServer(ctx, dashAddr, a.dash, a.log); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	a.log.Info("Onyx stopped cleanly")
	return nil
}

// Close releases resources.
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// DB exposes the database for CLI commands that need direct access.
func (a *App) DB() *database.DB { return a.db }

// proxyHandler builds the middleware-wrapped proxy handler.
// Stack (outermost first): Recovery -> BodyLimit -> RequestLogger -> SecureHeaders -> RateLimit -> Router
func (a *App) proxyHandler() http.Handler {
	globalLimiter := ratelimit.New(1000, 500)
	return middleware.Chain(
		a.router,
		middleware.Recovery(a.log),
		middleware.BodyLimit(defaultBodyLimit),
		middleware.RequestLogger(a.log),
		middleware.SecureHeaders(),
		middleware.RateLimit(globalLimiter),
	)
}

// NewFromConfigPath loads config and wires the App.
// Pass empty string to auto-detect config location.
func NewFromConfigPath(configPath string, development bool, version string) (*App, error) {
	log := logger.New(logLevel(development), development)

	if configPath == "" {
		var err error
		configPath, err = findConfig()
		if err != nil {
			return nil, err
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config %q: %w", configPath, err)
	}

	return New(cfg, log, version)
}

func logLevel(dev bool) slog.Level {
	if dev {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// findConfig searches standard locations for onyx.toml.
func findConfig() (string, error) {
	candidates := []string{"onyx.toml", "/etc/onyx/onyx.toml"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "onyx", "onyx.toml"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no onyx.toml found -- run 'onyx setup' first")
}
