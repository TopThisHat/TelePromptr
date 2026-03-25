// Package main is the entry point for the TelePromptr API server.
// It wires together all dependencies and starts the HTTP server with
// graceful shutdown support.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ralphlozano/telepromptr/apps/api/internal/infrastructure/config"
	"github.com/ralphlozano/telepromptr/apps/api/internal/interfaces/rest"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/httputil"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/logging"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/middleware"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/version"
)

// shutdownTimeout is the maximum duration to wait for in-flight requests
// to complete during graceful shutdown.
const shutdownTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// run contains the actual server lifecycle. Returning an error from here
// causes main to print it and exit with code 1.
func run() error {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Set up structured logging.
	logger := logging.New(os.Stdout, slog.LevelInfo)
	logger.Info("starting TelePromptr API server",
		slog.String("version", version.Version),
		slog.String("git_commit", version.GitCommit),
		slog.Int("http_port", cfg.HTTPPort),
		slog.Int("grpc_port", cfg.GRPCPort),
	)

	// Build the HTTP router and register routes.
	mux := rest.NewRouter(logger)
	registerRoutes(mux, cfg, logger)

	// Apply middleware chain.
	handler := rest.WithMiddleware(mux, logger)

	// Create the HTTP server.
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return logging.WithLogger(context.Background(), logger)
		},
	}

	// Start the server in a goroutine.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
		close(errCh)
	}()

	// Wait for interrupt signal or server error.
	return waitForShutdown(srv, logger, errCh)
}

// registerRoutes wires up all HTTP routes on the given mux.
func registerRoutes(mux *http.ServeMux, cfg *config.Config, logger *slog.Logger) {
	// Health check endpoint - unauthenticated.
	mux.HandleFunc("GET /healthz", handleHealthz)

	// Ready check endpoint - unauthenticated.
	mux.HandleFunc("GET /readyz", handleReadyz)

	// Version endpoint - unauthenticated.
	mux.HandleFunc("GET /version", handleVersion)

	// Protected API routes will be registered here as domain services are
	// implemented. They will use the Auth middleware with the appropriate
	// APIKeyLookupFunc backed by the database layer.
	_ = cfg    // will be used for auth config wiring
	_ = logger // will be used for route-specific logging
}

// handleHealthz returns 200 OK if the server is running. This is suitable
// for liveness probes (e.g., Kubernetes).
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	httputil.Success(w, map[string]string{"status": "ok"})
}

// handleReadyz returns 200 OK if the server is ready to accept traffic.
// In the future this will check database connectivity and other dependencies.
func handleReadyz(w http.ResponseWriter, r *http.Request) {
	httputil.Success(w, map[string]string{"status": "ready"})
}

// handleVersion returns the current server version information.
func handleVersion(w http.ResponseWriter, r *http.Request) {
	httputil.Success(w, map[string]string{
		"version":    version.Version,
		"git_commit": version.GitCommit,
		"build_time": version.BuildTime,
	})
}

// waitForShutdown blocks until a SIGTERM or SIGINT is received, then
// gracefully shuts down the HTTP server. If the server returns an error
// before a signal is received, that error is returned immediately.
func waitForShutdown(srv *http.Server, logger *slog.Logger, errCh <-chan error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	// Begin graceful shutdown.
	logger.Info("shutting down HTTP server", slog.Duration("timeout", shutdownTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	logger.Info("HTTP server stopped gracefully")
	return nil
}

// stubAPIKeyLookup is a placeholder APIKeyLookupFunc that always returns empty.
// It will be replaced with a real database-backed implementation when the
// persistence layer is built.
var _ middleware.APIKeyLookupFunc = stubAPIKeyLookup

func stubAPIKeyLookup(ctx context.Context, keyHash string) (string, error) {
	return "", nil
}
