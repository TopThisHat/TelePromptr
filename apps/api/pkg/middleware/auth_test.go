package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testAdminToken = "test-admin-secret"
	testProjectID  = "proj-123"
	testAPIKey     = "tp_proj_test_key_abc123"
)

// testLookup returns a mock APIKeyLookupFunc that recognizes the test API key.
func testLookup() APIKeyLookupFunc {
	expectedHash := HashAPIKey(testAPIKey)
	return func(ctx context.Context, keyHash string) (string, error) {
		if keyHash == expectedHash {
			return testProjectID, nil
		}
		return "", nil
	}
}

// errorLookup returns a mock APIKeyLookupFunc that always returns an error.
func errorLookup() APIKeyLookupFunc {
	return func(ctx context.Context, keyHash string) (string, error) {
		return "", fmt.Errorf("database connection failed")
	}
}

func newAuthMiddleware(lookup APIKeyLookupFunc) func(http.Handler) http.Handler {
	return Auth(AuthConfig{
		AdminToken:   testAdminToken,
		LookupAPIKey: lookup,
	})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_AdminToken_Valid(t *testing.T) {
	t.Parallel()

	var gotPC *ProjectContext
	handler := newAuthMiddleware(testLookup())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPC = GetProjectContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AdminTokenHeader, testAdminToken)
	req.Header.Set(ProjectIDHeader, testProjectID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotPC == nil {
		t.Fatal("expected ProjectContext in context")
	}
	if !gotPC.IsAdmin {
		t.Error("expected IsAdmin = true for admin auth")
	}
	if gotPC.ProjectID != testProjectID {
		t.Errorf("ProjectID = %q, want %q", gotPC.ProjectID, testProjectID)
	}
}

func TestAuth_AdminToken_Invalid(t *testing.T) {
	t.Parallel()

	handler := newAuthMiddleware(testLookup())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AdminTokenHeader, "wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var errResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	errObj, ok := errResp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if code := errObj["code"]; code != "unauthorized" {
		t.Errorf("error code = %v, want %q", code, "unauthorized")
	}
}

func TestAuth_AdminToken_NoProjectID(t *testing.T) {
	t.Parallel()

	var gotPC *ProjectContext
	handler := newAuthMiddleware(testLookup())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPC = GetProjectContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AdminTokenHeader, testAdminToken)
	// No X-Project-Id header.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotPC == nil {
		t.Fatal("expected ProjectContext")
	}
	if gotPC.ProjectID != "" {
		t.Errorf("ProjectID = %q, want empty", gotPC.ProjectID)
	}
}

func TestAuth_APIKey_Valid(t *testing.T) {
	t.Parallel()

	var gotPC *ProjectContext
	handler := newAuthMiddleware(testLookup())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPC = GetProjectContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotPC == nil {
		t.Fatal("expected ProjectContext")
	}
	if gotPC.IsAdmin {
		t.Error("expected IsAdmin = false for API key auth")
	}
	if gotPC.ProjectID != testProjectID {
		t.Errorf("ProjectID = %q, want %q", gotPC.ProjectID, testProjectID)
	}
}

func TestAuth_APIKey_NotFound(t *testing.T) {
	t.Parallel()

	handler := newAuthMiddleware(testLookup())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tp_proj_unknown_key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_APIKey_WrongPrefix(t *testing.T) {
	t.Parallel()

	handler := newAuthMiddleware(testLookup())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sk_wrong_prefix_key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_APIKey_LookupError(t *testing.T) {
	t.Parallel()

	handler := newAuthMiddleware(errorLookup())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAuth_NoAuth(t *testing.T) {
	t.Parallel()

	handler := newAuthMiddleware(testLookup())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidBearerFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
	}{
		{name: "no bearer prefix", header: "Basic dXNlcjpwYXNz"},
		{name: "bearer only", header: "Bearer "},
		{name: "empty bearer", header: "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := newAuthMiddleware(testLookup())(okHandler())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestGetProjectContext_EmptyContext(t *testing.T) {
	t.Parallel()

	pc := GetProjectContext(context.Background())
	if pc != nil {
		t.Errorf("expected nil ProjectContext from empty context, got %+v", pc)
	}
}

func TestHashAPIKey(t *testing.T) {
	t.Parallel()

	hash1 := HashAPIKey("tp_proj_test_key")
	hash2 := HashAPIKey("tp_proj_test_key")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64 (SHA-256 hex)", len(hash1))
	}

	// Different input should produce different hash.
	hash3 := HashAPIKey("tp_proj_different_key")
	if hash1 == hash3 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestExtractBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{name: "valid", header: "Bearer mytoken", wantToken: "mytoken", wantOK: true},
		{name: "with spaces", header: "Bearer  mytoken ", wantToken: "mytoken", wantOK: true},
		{name: "empty token", header: "Bearer ", wantOK: false},
		{name: "no space", header: "Bearer", wantOK: false},
		{name: "wrong scheme", header: "Basic abc", wantOK: false},
		{name: "empty", header: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token, ok := extractBearerToken(tt.header)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && token != tt.wantToken {
				t.Errorf("token = %q, want %q", token, tt.wantToken)
			}
		})
	}
}
