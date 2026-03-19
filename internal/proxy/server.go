package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// StartServer starts an HTTP server on addr with graceful shutdown support.
// It blocks until ctx is cancelled, then gives in-flight requests 10 seconds
// to complete before forcing a shutdown.
func StartServer(ctx context.Context, addr string, handler http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("proxy listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("proxy server: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Info("proxy shutting down gracefully")
		return srv.Shutdown(shutCtx)
	}
}
