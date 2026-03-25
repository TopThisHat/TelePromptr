package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesID(t *testing.T) {
	t.Parallel()

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("expected request ID in context, got empty string")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	respID := rec.Header().Get(RequestIDHeader)
	if respID == "" {
		t.Error("expected X-Request-Id response header, got empty")
	}
	// Generated IDs should be 32 hex chars (16 bytes).
	if len(respID) != 32 {
		t.Errorf("request ID length = %d, want 32", len(respID))
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	t.Parallel()

	const existingID = "existing-id-12345"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id != existingID {
			t.Errorf("request ID = %q, want %q", id, existingID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != existingID {
		t.Errorf("response X-Request-Id = %q, want %q", got, existingID)
	}
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	t.Parallel()

	id := GetRequestID(context.Background())
	if id != "" {
		t.Errorf("expected empty string from empty context, got %q", id)
	}
}

func TestLogging_LogsRequest(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if method, ok := entry["method"].(string); !ok || method != "POST" {
		t.Errorf("method = %v, want POST", entry["method"])
	}
	if path, ok := entry["path"].(string); !ok || path != "/api/test" {
		t.Errorf("path = %v, want /api/test", entry["path"])
	}
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("status = %v, want %d", entry["status"], http.StatusCreated)
	}
}

func TestLogging_InjectsLoggerIntoContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Chain: RequestID -> Logging -> handler
	handler := RequestID(Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The logging middleware should have stored a logger with request_id.
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// The log entry should contain a request_id.
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}
	if _, ok := entry["request_id"]; !ok {
		t.Error("expected request_id in log entry when RequestID middleware is applied")
	}
}

func TestCORS_SetsHeaders(t *testing.T) {
	t.Parallel()

	cfg := DefaultCORSConfig()
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers header to be set")
	}
}

func TestCORS_Preflight(t *testing.T) {
	t.Parallel()

	cfg := DefaultCORSConfig()
	called := false
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if called {
		t.Error("preflight should not call the next handler")
	}
}

func TestCORS_AllowCredentials(t *testing.T) {
	t.Parallel()

	cfg := DefaultCORSConfig()
	cfg.AllowCredentials = true
	handler := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRecovery_WithPanic(t *testing.T) {
	t.Parallel()

	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Should not panic.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	// Should return a JSON error body.
	var errResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	errObj, ok := errResp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if code, ok := errObj["code"].(string); !ok || code != "internal_error" {
		t.Errorf("error code = %v, want %q", errObj["code"], "internal_error")
	}
}

func TestChain_Order(t *testing.T) {
	t.Parallel()

	var order []string

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := Chain(mw("A"), mw("B"), mw("C"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := []string{
		"A-before", "B-before", "C-before",
		"handler",
		"C-after", "B-after", "A-after",
	}
	if len(order) != len(expected) {
		t.Fatalf("chain order length = %d, want %d: %v", len(order), len(expected), order)
	}
	for i, want := range expected {
		if order[i] != want {
			t.Errorf("chain order[%d] = %q, want %q", i, order[i], want)
		}
	}
}

func TestStatusWriter_CapturesStatus(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: http.StatusOK}

	sw.WriteHeader(http.StatusNotFound)
	if sw.status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", sw.status, http.StatusNotFound)
	}

	// Second call should not change the captured status.
	sw.WriteHeader(http.StatusOK)
	if sw.status != http.StatusNotFound {
		t.Errorf("status after second WriteHeader = %d, want %d (should not change)", sw.status, http.StatusNotFound)
	}
}

func TestStatusWriter_WriteDefaultsStatus(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: http.StatusOK}

	_, err := sw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After Write without explicit WriteHeader, status should remain 200.
	if sw.status != http.StatusOK {
		t.Errorf("status = %d, want %d", sw.status, http.StatusOK)
	}
}
