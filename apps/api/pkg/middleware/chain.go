// Package middleware provides HTTP middleware for the TelePromptr API server
// including CORS, request ID generation, structured logging, and panic recovery.
package middleware

import "net/http"

// Chain composes multiple middleware functions into a single middleware.
// Middleware is applied in the order given, so Chain(A, B, C)(handler) means
// the request flows through A -> B -> C -> handler and the response flows back
// through C -> B -> A.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
