package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setRequiredEnv sets all required environment variables with valid test values.
func setRequiredEnv(t *testing.T) {
	t.Helper()
	envs := map[string]string{
		"DATABASE_URL":               "postgres://user:pass@localhost:5432/testdb",
		"TELEPROMPTR_ADMIN_TOKEN":    "test-admin-token-value",
		"TELEPROMPTR_ENCRYPTION_KEY": "this-key-is-at-least-32-bytes-long!!",
	}
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

// clearOptionalEnv ensures optional env vars are unset so defaults are used.
func clearOptionalEnv(t *testing.T) {
	t.Helper()
	optionals := []string{
		"HTTP_PORT", "GRPC_PORT", "OTLP_HTTP_PORT",
		"BUFFER_SIZE", "BATCH_SIZE", "FLUSH_INTERVAL_MS",
	}
	for _, k := range optionals {
		os.Unsetenv(k)
	}
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)
	clearOptionalEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPPort != DefaultHTTPPort {
		t.Errorf("HTTPPort = %d, want %d", cfg.HTTPPort, DefaultHTTPPort)
	}
	if cfg.GRPCPort != DefaultGRPCPort {
		t.Errorf("GRPCPort = %d, want %d", cfg.GRPCPort, DefaultGRPCPort)
	}
	if cfg.OTLPHTTPPort != DefaultOTLPHTTP {
		t.Errorf("OTLPHTTPPort = %d, want %d", cfg.OTLPHTTPPort, DefaultOTLPHTTP)
	}
	if cfg.BufferSize != DefaultBufferSize {
		t.Errorf("BufferSize = %d, want %d", cfg.BufferSize, DefaultBufferSize)
	}
	if cfg.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, DefaultBatchSize)
	}
	if cfg.FlushIntervalMS != DefaultFlushMS {
		t.Errorf("FlushIntervalMS = %d, want %d", cfg.FlushIntervalMS, DefaultFlushMS)
	}
}

func TestLoad_CustomPorts(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("GRPC_PORT", "50051")
	t.Setenv("OTLP_HTTP_PORT", "4319")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPPort != 9090 {
		t.Errorf("HTTPPort = %d, want 9090", cfg.HTTPPort)
	}
	if cfg.GRPCPort != 50051 {
		t.Errorf("GRPCPort = %d, want 50051", cfg.GRPCPort)
	}
	if cfg.OTLPHTTPPort != 4319 {
		t.Errorf("OTLPHTTPPort = %d, want 4319", cfg.OTLPHTTPPort)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		unset   string
		wantMsg string
	}{
		{
			name:    "missing DATABASE_URL",
			unset:   "DATABASE_URL",
			wantMsg: "DATABASE_URL is required",
		},
		{
			name:    "missing TELEPROMPTR_ADMIN_TOKEN",
			unset:   "TELEPROMPTR_ADMIN_TOKEN",
			wantMsg: "TELEPROMPTR_ADMIN_TOKEN is required",
		},
		{
			name:    "missing TELEPROMPTR_ENCRYPTION_KEY",
			unset:   "TELEPROMPTR_ENCRYPTION_KEY",
			wantMsg: "TELEPROMPTR_ENCRYPTION_KEY is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			clearOptionalEnv(t)
			os.Unsetenv(tt.unset)

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); !strings.Contains(got, tt.wantMsg) {
				t.Errorf("error = %q, want to contain %q", got, tt.wantMsg)
			}
		})
	}
}

func TestLoad_EncryptionKeyTooShort(t *testing.T) {
	setRequiredEnv(t)
	clearOptionalEnv(t)
	t.Setenv("TELEPROMPTR_ENCRYPTION_KEY", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short encryption key")
	}
	if got := err.Error(); !strings.Contains(got, "at least 32 bytes") {
		t.Errorf("error = %q, want to mention 32 bytes", got)
	}
}

func TestLoadFromEnvFile(t *testing.T) {
	// Write a temp .env file.
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `# Comment line
DATABASE_URL=postgres://file:pass@localhost/filedb
TELEPROMPTR_ADMIN_TOKEN=file-token
TELEPROMPTR_ENCRYPTION_KEY="this-is-a-long-key-for-file-test!!"
HTTP_PORT=3000
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing .env: %v", err)
	}

	// Clear everything so the .env file is the only source.
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("TELEPROMPTR_ADMIN_TOKEN")
	os.Unsetenv("TELEPROMPTR_ENCRYPTION_KEY")
	clearOptionalEnv(t)

	cfg, err := LoadFromEnvFile(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://file:pass@localhost/filedb" {
		t.Errorf("DatabaseURL = %q, want value from .env file", cfg.DatabaseURL)
	}
	if cfg.HTTPPort != 3000 {
		t.Errorf("HTTPPort = %d, want 3000", cfg.HTTPPort)
	}
}

func TestLoadFromEnvFile_EnvVarsTakePrecedence(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `DATABASE_URL=postgres://file:pass@localhost/filedb
TELEPROMPTR_ADMIN_TOKEN=file-token
TELEPROMPTR_ENCRYPTION_KEY=this-is-a-long-key-for-file-test!!
HTTP_PORT=3000
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing .env: %v", err)
	}

	// Set an env var that should override the .env file value.
	t.Setenv("HTTP_PORT", "9999")

	cfg, err := LoadFromEnvFile(envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPPort != 9999 {
		t.Errorf("HTTPPort = %d, want 9999 (env var should override .env file)", cfg.HTTPPort)
	}
}

func TestLoadFromEnvFile_FileNotFound(t *testing.T) {
	setRequiredEnv(t)

	_, err := LoadFromEnvFile("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing .env file")
	}
}

func TestParseDotEnvLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{
			name:      "simple key=value",
			line:      "FOO=bar",
			wantKey:   "FOO",
			wantValue: "bar",
			wantOK:    true,
		},
		{
			name:      "double quoted value",
			line:      `DB_URL="postgres://localhost/db"`,
			wantKey:   "DB_URL",
			wantValue: "postgres://localhost/db",
			wantOK:    true,
		},
		{
			name:      "single quoted value",
			line:      "SECRET='my secret'",
			wantKey:   "SECRET",
			wantValue: "my secret",
			wantOK:    true,
		},
		{
			name:      "value with equals sign",
			line:      "URL=postgres://user:pass@host/db?sslmode=disable",
			wantKey:   "URL",
			wantValue: "postgres://user:pass@host/db?sslmode=disable",
			wantOK:    true,
		},
		{
			name:      "empty value",
			line:      "EMPTY=",
			wantKey:   "EMPTY",
			wantValue: "",
			wantOK:    true,
		},
		{
			name:      "spaces around equals",
			line:      "KEY = value",
			wantKey:   "KEY",
			wantValue: "value",
			wantOK:    true,
		},
		{
			name:   "no equals sign",
			line:   "NOEQUALSSIGN",
			wantOK: false,
		},
		{
			name:   "equals at start",
			line:   "=value",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key, value, ok := parseDotEnvLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

func TestIntEnvOrDefault(t *testing.T) {
	tests := []struct {
		name       string
		envVal     string
		setEnv     bool
		defaultVal int
		want       int
	}{
		{name: "unset returns default", setEnv: false, defaultVal: 42, want: 42},
		{name: "empty returns default", envVal: "", setEnv: true, defaultVal: 42, want: 42},
		{name: "invalid returns default", envVal: "notanumber", setEnv: true, defaultVal: 42, want: 42},
		{name: "valid integer", envVal: "9090", setEnv: true, defaultVal: 42, want: 9090},
		{name: "zero is valid", envVal: "0", setEnv: true, defaultVal: 42, want: 0},
		{name: "negative is valid", envVal: "-1", setEnv: true, defaultVal: 42, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT_ENV_" + strings.ReplaceAll(tt.name, " ", "_")
			if tt.setEnv {
				t.Setenv(key, tt.envVal)
			}
			got := intEnvOrDefault(key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("intEnvOrDefault(%q, %d) = %d, want %d", key, tt.defaultVal, got, tt.want)
			}
		})
	}
}
