package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// requestIDKey is the context key for storing the request ID.
type requestIDKey struct{}

// RequestIDHeader is the HTTP header used to propagate request IDs.
const RequestIDHeader = "X-Request-Id"

// requestIDByteLen is the number of random bytes used to generate a request ID,
// yielding a 32-character hex string.
const requestIDByteLen = 16

// RequestID is middleware that assigns a unique request ID to every request.
// If the incoming request already has an X-Request-Id header, that value is
// preserved. Otherwise a new random hex ID is generated. The ID is stored
// in the request context and set on the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = generateRequestID()
		}

		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context. Returns an empty
// string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// generateRequestID produces a cryptographically random hex string suitable
// for use as a request correlation ID.
func generateRequestID() string {
	b := make([]byte, requestIDByteLen)
	if _, err := rand.Read(b); err != nil {
		// Fallback: should never happen with crypto/rand.
		return "unknown"
	}
	return hex.EncodeToString(b)
}
