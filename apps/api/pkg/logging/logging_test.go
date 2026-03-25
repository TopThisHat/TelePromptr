package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo)

	logger.Info("test message", slog.String("key", "value"))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log entry: %v", err)
	}

	if msg, ok := entry["msg"].(string); !ok || msg != "test message" {
		t.Errorf("msg = %v, want %q", entry["msg"], "test message")
	}
	if val, ok := entry["key"].(string); !ok || val != "value" {
		t.Errorf("key = %v, want %q", entry["key"], "value")
	}
}

func TestNew_NilWriter(t *testing.T) {
	t.Parallel()

	// Should not panic when w is nil; defaults to os.Stdout.
	logger := New(nil, slog.LevelInfo)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, slog.LevelWarn)

	// Info should be filtered out at Warn level.
	logger.Info("should not appear")
	if buf.Len() > 0 {
		t.Error("info message should have been filtered at Warn level")
	}

	logger.Warn("should appear")
	if buf.Len() == 0 {
		t.Error("warn message should have been logged at Warn level")
	}
}

func TestNewDefault(t *testing.T) {
	t.Parallel()

	logger := NewDefault()
	if logger == nil {
		t.Fatal("expected non-nil logger from NewDefault")
	}
}

func TestWithLogger_FromContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo)

	ctx := WithLogger(context.Background(), logger)
	recovered := FromContext(ctx)

	recovered.Info("recovered message")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log entry: %v", err)
	}
	if msg, ok := entry["msg"].(string); !ok || msg != "recovered message" {
		t.Errorf("msg = %v, want %q", entry["msg"], "recovered message")
	}
}

func TestFromContext_NoLogger(t *testing.T) {
	t.Parallel()

	// When no logger is in the context, FromContext should return the default.
	logger := FromContext(context.Background())
	if logger == nil {
		t.Fatal("expected non-nil logger from FromContext with empty context")
	}
}

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo)

	reqLogger := WithRequestID(logger, "req-abc-123")
	reqLogger.Info("request log")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log entry: %v", err)
	}
	if rid, ok := entry["request_id"].(string); !ok || rid != "req-abc-123" {
		t.Errorf("request_id = %v, want %q", entry["request_id"], "req-abc-123")
	}
}

func TestWithLogger_NilValue(t *testing.T) {
	t.Parallel()

	// Storing a nil logger should cause FromContext to fall back to default.
	ctx := context.WithValue(context.Background(), loggerKey, (*slog.Logger)(nil))
	logger := FromContext(ctx)
	if logger == nil {
		t.Fatal("expected non-nil logger even when nil was stored in context")
	}
}
