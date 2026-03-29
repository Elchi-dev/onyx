// Package app is the dependency wiring layer for Onyx.
// It owns the lifecycle of all major components and connects them together.
// CLI commands only need to know about app.App — not individual internals.
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
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
	tlsmgr "github.com/Elchi-dev/onyx/internal/tls"
)

const defaultBodyLimit = 10 << 20

// App holds all wired components and orchestrates startup and shutdown.
type App struct {
	cfg     *config.Config
	log     *slog.Logger
	db      *database.DB
	dash    *dashboard.Dashboard
	router  *proxy.Router
	tlsMgr  *tlsmgr.Manager
	version string
}

// New wires all components from the given config and logger.
func New(cfg *config.Config, log *slog.Logger, version string) (*App, error) {
	db, err := database.Open(cfg.DBPath())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	var httpsHosts []string
	router := proxy.New(log, nil)

	for _, r := range cfg.Routes {
		if r.Enabled {
			if err := router.AddRoute(r.Host, r.Target); err != nil {
				log.Warn("skipping invalid config route", "host", r.Host, "error", err)
				continue
			}
			if r.HTTPS {
				httpsHosts = append(httpsHosts, r.Host)
			}
		}
	}

	dbRoutes, err := db.ListRoutes()
	if err != nil {
		return nil, fmt.Errorf("loading db routes: %w", err)
	}
	for _, r := range dbRoutes {
		if r.Enabled {
			_ = router.AddRoute(r.Host, r.Target)
		}
		if r.HTTPS {
			httpsHosts = append(httpsHosts, r.Host)
		}
	}

	log.Info("routes loaded",
		"config", len(cfg.Routes),
		"database", len(dbRoutes),
	)

	tm := tlsmgr.New(cfg.Server.DataDir, httpsHosts, log)
	dash := dashboard.New(log, db, tm)
	dash.SetVersion(version)
	router.SetEventHandler(dash.BroadcastRequest)

	return &App{
		cfg: cfg, log: log, db: db,
		dash: dash, router: router, tlsMgr: tm, version: version,
	}, nil
}

// Run starts all servers and blocks until SIGTERM or SIGINT.
func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	proxyAddr := fmt.Sprintf(":%d", a.cfg.Server.HTTPPort)
	httpsAddr := fmt.Sprintf(":%d", a.cfg.Server.HTTPSPort)
	dashAddr := fmt.Sprintf(":%d", a.cfg.Dashboard.Port)
	hasHTTPS := a.hasHTTPSRoutes()

	a.log.Info("Onyx starting",
		slog.String("proxy", proxyAddr),
		slog.String("dashboard", dashAddr),
		slog.String("version", a.version),
		slog.Bool("https", hasHTTPS),
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpHandler := a.proxyHandler()
		if hasHTTPS {
			redirect := tlsmgr.RedirectToHTTPS(a.cfg.Server.HTTPSPort)
			httpHandler = a.tlsMgr.HTTPHandler(redirect)
		}
		if err := proxy.StartServer(ctx, proxyAddr, httpHandler, a.log); err != nil {
			errCh <- err
		}
	}()

	if hasHTTPS {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.startHTTPSServer(ctx, httpsAddr); err != nil {
				errCh <- err
			}
		}()
	}

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

func (a *App) startHTTPSServer(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:      addr,
		Handler:   a.proxyHandler(),
		TLSConfig: a.tlsMgr.TLSConfig(),
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("https server: %w", err)
	}
	tlsLn := tls.NewListener(ln, a.tlsMgr.TLSConfig())
	errCh := make(chan error, 1)
	go func() {
		a.log.Info("https proxy listening", "addr", addr)
		if err := srv.Serve(tlsLn); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("https server: %w", err)
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return srv.Close()
	}
}

func (a *App) hasHTTPSRoutes() bool {
	for _, r := range a.cfg.Routes {
		if r.HTTPS && r.Enabled {
			return true
		}
	}
	routes, _ := a.db.ListRoutes()
	for _, r := range routes {
		if r.HTTPS && r.Enabled {
			return true
		}
	}
	return false
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *App) DB() *database.DB { return a.db }

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
	return "", fmt.Errorf("no onyx.toml found — run 'onyx setup' first")
}
