package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS middleware configuration.
type CORSConfig struct {
	// AllowedOrigins is the list of origins that are allowed. Use "*" to allow all.
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods allowed for cross-origin requests.
	AllowedMethods []string

	// AllowedHeaders is the list of HTTP headers allowed in cross-origin requests.
	AllowedHeaders []string

	// AllowCredentials indicates whether the response to the request can be
	// exposed when the credentials flag is true.
	AllowCredentials bool

	// MaxAge is the max age (in seconds) for preflight cache.
	MaxAge string
}

// DefaultCORSConfig returns a CORSConfig with sensible defaults for
// development and typical API usage.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-Id", "X-Admin-Token", "X-Project-Id"},
		AllowCredentials: false,
		MaxAge:           "86400",
	}
}

// CORS returns middleware that handles Cross-Origin Resource Sharing headers
// according to the provided configuration. It automatically handles preflight
// OPTIONS requests by returning 204 No Content with appropriate headers.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	allowOrigin := strings.Join(cfg.AllowedOrigins, ", ")
	allowMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			if cfg.MaxAge != "" {
				w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
			}
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
