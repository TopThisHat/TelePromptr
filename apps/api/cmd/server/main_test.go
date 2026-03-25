package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ralphlozano/telepromptr/apps/api/internal/infrastructure/config"
	"github.com/ralphlozano/telepromptr/apps/api/internal/interfaces/rest"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/httputil"
	"github.com/ralphlozano/telepromptr/apps/api/pkg/version"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
}

func testConfig() *config.Config {
	return &config.Config{
		DatabaseURL:     "postgres://test:test@localhost/test",
		HTTPPort:        8080,
		GRPCPort:        4317,
		OTLPHTTPPort:    4318,
		AdminToken:      "test-admin-token",
		EncryptionKey:   "test-encryption-key-at-least-32-bytes!",
		BufferSize:      1024,
		BatchSize:       100,
		FlushIntervalMS: 5000,
	}
}

func TestHandleHealthz(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handleHealthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp httputil.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if status, ok := data["status"].(string); !ok || status != "ok" {
		t.Errorf("status = %v, want %q", data["status"], "ok")
	}
}

func TestHandleReadyz(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handleReadyz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp httputil.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if status, ok := data["status"].(string); !ok || status != "ready" {
		t.Errorf("status = %v, want %q", data["status"], "ready")
	}
}

func TestHandleVersion(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	handleVersion(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp httputil.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if v, ok := data["version"].(string); !ok || v != version.Version {
		t.Errorf("version = %v, want %q", data["version"], version.Version)
	}
	if gc, ok := data["git_commit"].(string); !ok || gc != version.GitCommit {
		t.Errorf("git_commit = %v, want %q", data["git_commit"], version.GitCommit)
	}
}

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	logger := testLogger()
	cfg := testConfig()
	mux := rest.NewRouter(logger)
	registerRoutes(mux, cfg, logger)

	// Test that the routes are registered and respond correctly.
	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{name: "healthz", method: http.MethodGet, path: "/healthz", want: http.StatusOK},
		{name: "readyz", method: http.MethodGet, path: "/readyz", want: http.StatusOK},
		{name: "version", method: http.MethodGet, path: "/version", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.want)
			}
		})
	}
}

func TestWithMiddleware_Integration(t *testing.T) {
	t.Parallel()

	logger := testLogger()
	cfg := testConfig()
	mux := rest.NewRouter(logger)
	registerRoutes(mux, cfg, logger)
	handler := rest.WithMiddleware(mux, logger)

	// The health endpoint should work through the full middleware stack.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should have X-Request-Id header from middleware.
	if rid := rec.Header().Get("X-Request-Id"); rid == "" {
		t.Error("expected X-Request-Id header from middleware")
	}

	// Should have CORS headers.
	if cors := rec.Header().Get("Access-Control-Allow-Origin"); cors == "" {
		t.Error("expected Access-Control-Allow-Origin header from middleware")
	}
}

func TestStubAPIKeyLookup(t *testing.T) {
	t.Parallel()

	projectID, err := stubAPIKeyLookup(nil, "any-hash")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if projectID != "" {
		t.Errorf("projectID = %q, want empty", projectID)
	}
}
