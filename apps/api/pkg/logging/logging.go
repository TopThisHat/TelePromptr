// Package logging provides structured logging for the TelePromptr API server
// using Go's standard log/slog package with JSON output. It supports request
// correlation by embedding a logger with a request ID into the context.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// contextKey is an unexported type used for context value keys in this package,
// preventing collisions with keys from other packages.
type contextKey int

const (
	// loggerKey is the context key for storing a *slog.Logger.
	loggerKey contextKey = iota
)

// New creates a new JSON-formatted slog.Logger writing to the given writer
// at the specified level. If w is nil, os.Stdout is used.
func New(w io.Writer, level slog.Level) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	}))
}

// NewDefault creates a new JSON-formatted slog.Logger writing to os.Stdout
// at slog.LevelInfo.
func NewDefault() *slog.Logger {
	return New(os.Stdout, slog.LevelInfo)
}

// WithLogger returns a new context derived from ctx that carries the given
// logger. Use FromContext to retrieve it later.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts the logger from the context. If no logger is present,
// it returns the default slog logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// WithRequestID returns a child logger that includes the given request ID
// as a structured attribute. The returned logger should typically be stored
// back into the context via WithLogger.
func WithRequestID(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With(slog.String("request_id", requestID))
}
