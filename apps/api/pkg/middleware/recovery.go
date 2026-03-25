package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/ralphlozano/telepromptr/apps/api/pkg/logging"
)

// Recovery is middleware that recovers from panics in downstream handlers,
// logs the panic with a stack trace, and returns a 500 Internal Server Error
// response. It prevents a single panicking request from crashing the server.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger := logging.FromContext(r.Context())
				logger.Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, `{"error":{"code":"internal_error","message":"internal server error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
