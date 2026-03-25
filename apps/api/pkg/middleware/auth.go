package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// projectContextKey is an unexported type for storing the authenticated project
// context in request contexts.
type projectContextKey struct{}

// ProjectContext holds the authenticated project information resolved from
// either API key or admin authentication.
type ProjectContext struct {
	// ProjectID is the identifier of the project this request is scoped to.
	ProjectID string

	// IsAdmin indicates whether the request was authenticated via admin token.
	IsAdmin bool
}

// APIKeyLookupFunc is a function that looks up an API key by its SHA-256 hash
// and returns the associated project ID. It returns an empty string if the key
// is not found.
type APIKeyLookupFunc func(ctx context.Context, keyHash string) (projectID string, err error)

// AuthConfig holds configuration for the dual authentication middleware.
type AuthConfig struct {
	// AdminToken is the expected value of the X-Admin-Token header for admin auth.
	AdminToken string

	// LookupAPIKey is called to resolve a hashed API key to a project ID.
	LookupAPIKey APIKeyLookupFunc
}

// AdminTokenHeader is the HTTP header used for admin authentication.
const AdminTokenHeader = "X-Admin-Token"

// ProjectIDHeader is the HTTP header used by admin requests to specify
// which project to operate on.
const ProjectIDHeader = "X-Project-Id"

// apiKeyPrefix is the expected prefix for API keys in the Authorization header.
const apiKeyPrefix = "tp_proj_"

// Auth returns middleware that implements dual authentication:
//
//  1. API Key auth: Authorization header with "Bearer tp_proj_..." token.
//     The raw key is SHA-256 hashed and looked up via the configured
//     APIKeyLookupFunc to resolve the project ID.
//
//  2. Admin auth: X-Admin-Token header matching the configured admin token.
//     The project is identified via the X-Project-Id header.
//
// On success, a ProjectContext is stored in the request context. On failure,
// a 401 Unauthorized response is returned.
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try admin auth first.
			if adminToken := r.Header.Get(AdminTokenHeader); adminToken != "" {
				if subtle.ConstantTimeCompare([]byte(adminToken), []byte(cfg.AdminToken)) == 1 {
					projectID := r.Header.Get(ProjectIDHeader)
					pc := &ProjectContext{
						ProjectID: projectID,
						IsAdmin:   true,
					}
					ctx := context.WithValue(r.Context(), projectContextKey{}, pc)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				writeAuthError(w, "invalid admin token")
				return
			}

			// Try API key auth.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, "missing authentication")
				return
			}

			rawKey, ok := extractBearerToken(authHeader)
			if !ok {
				writeAuthError(w, "invalid authorization header format")
				return
			}

			if !strings.HasPrefix(rawKey, apiKeyPrefix) {
				writeAuthError(w, "invalid API key format")
				return
			}

			keyHash := HashAPIKey(rawKey)
			projectID, err := cfg.LookupAPIKey(r.Context(), keyHash)
			if err != nil {
				http.Error(w, `{"error":{"code":"internal_error","message":"authentication lookup failed"}}`, http.StatusInternalServerError)
				return
			}
			if projectID == "" {
				writeAuthError(w, "invalid API key")
				return
			}

			pc := &ProjectContext{
				ProjectID: projectID,
				IsAdmin:   false,
			}
			ctx := context.WithValue(r.Context(), projectContextKey{}, pc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetProjectContext extracts the authenticated ProjectContext from the request
// context. Returns nil if the request has not been authenticated.
func GetProjectContext(ctx context.Context) *ProjectContext {
	pc, _ := ctx.Value(projectContextKey{}).(*ProjectContext)
	return pc
}

// HashAPIKey computes the SHA-256 hex digest of the given raw API key.
// This is the value stored in the database and used for lookups.
func HashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

// extractBearerToken parses a "Bearer <token>" Authorization header value.
// It returns the token and true, or an empty string and false if the header
// is not in the expected format.
func extractBearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

// writeAuthError writes a 401 Unauthorized JSON error response.
// It uses json.Marshal to safely encode the message string, preventing
// JSON injection if a caller ever passes dynamic content.
func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	resp := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	resp.Error.Code = "unauthorized"
	resp.Error.Message = message
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
