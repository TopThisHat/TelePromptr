package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ralphlozano/telepromptr/apps/api/pkg/logging"
)

// Logging is middleware that logs each HTTP request and response with
// structured attributes including method, path, status code, duration,
// and request ID (if present in context). It injects a child logger with
// the request ID into the request context for downstream handlers.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Attach request ID to the logger if available.
			reqID := GetRequestID(r.Context())
			reqLogger := logger
			if reqID != "" {
				reqLogger = logging.WithRequestID(logger, reqID)
			}

			// Store the enriched logger in the request context.
			ctx := logging.WithLogger(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			// Wrap the response writer to capture the status code.
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			reqLogger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the response status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wroteHeader {
		sw.status = code
		sw.wroteHeader = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

// Write delegates to the underlying writer, recording a 200 status if
// WriteHeader has not yet been called.
func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.wroteHeader = true
	}
	return sw.ResponseWriter.Write(b)
}
