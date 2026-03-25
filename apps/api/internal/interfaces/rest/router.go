// Package rest provides the HTTP REST interface for the TelePromptr API.
// It configures routing using Go 1.22+ pattern matching and applies the
// standard middleware chain.
package rest

import (
	"log/slog"
	"net/http"

	"github.com/ralphlozano/telepromptr/apps/api/pkg/middleware"
)

// NewRouter creates a new http.Handler with the standard middleware chain
// applied: CORS, request ID, logging, and panic recovery. The returned mux
// can be used to register additional routes before serving.
func NewRouter(logger *slog.Logger) *http.ServeMux {
	return http.NewServeMux()
}

// WithMiddleware wraps the given handler with the standard TelePromptr
// middleware chain in the correct order:
//
//  1. Panic recovery (outermost - catches panics from everything below)
//  2. CORS (handles preflight before other processing)
//  3. Request ID (generates/propagates correlation IDs)
//  4. Logging (logs with request ID in context)
func WithMiddleware(handler http.Handler, logger *slog.Logger) http.Handler {
	chain := middleware.Chain(
		middleware.Recovery,
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.RequestID,
		middleware.Logging(logger),
	)
	return chain(handler)
}
