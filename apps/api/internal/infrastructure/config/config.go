// Package config provides configuration loading for the TelePromptr API server.
// It reads values from environment variables and optional .env files, then
// validates that all required settings are present and within acceptable ranges.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Default values for optional configuration settings.
const (
	DefaultHTTPPort    = 8080
	DefaultGRPCPort    = 4317
	DefaultOTLPHTTP    = 4318
	DefaultBufferSize  = 1024
	DefaultBatchSize   = 100
	DefaultFlushMS     = 5000
	MinEncryptionKeyLen = 32
)

// Config holds all configuration values for the TelePromptr API server.
// Required fields are validated at load time; optional fields carry defaults.
type Config struct {
	// DatabaseURL is the PostgreSQL connection string. Required.
	DatabaseURL string

	// HTTPPort is the port for the REST API server.
	HTTPPort int

	// GRPCPort is the port for the OTLP gRPC ingest server.
	GRPCPort int

	// OTLPHTTPPort is the port for the OTLP HTTP ingest server.
	OTLPHTTPPort int

	// AdminToken is the shared secret used for admin authentication
	// via the X-Admin-Token header. Required.
	AdminToken string

	// EncryptionKey is used to encrypt sensitive data at rest.
	// Must be at least 32 bytes. Required.
	EncryptionKey string

	// BufferSize is the capacity of in-memory trace/span buffers.
	BufferSize int

	// BatchSize controls how many items are flushed to the database at once.
	BatchSize int

	// FlushIntervalMS is the maximum time in milliseconds between buffer flushes.
	FlushIntervalMS int
}

// Load reads configuration from the environment (and an optional .env file)
// and returns a validated Config. It returns an error describing all
// validation failures if any required fields are missing or invalid.
func Load() (*Config, error) {
	return LoadFromEnvFile("")
}

// LoadFromEnvFile reads configuration from the specified .env file path (if
// non-empty and the file exists), overlays environment variables on top, then
// validates the result. Environment variables always take precedence over
// values found in the .env file.
func LoadFromEnvFile(envFilePath string) (*Config, error) {
	if envFilePath != "" {
		if err := loadDotEnv(envFilePath); err != nil {
			return nil, fmt.Errorf("loading .env file %q: %w", envFilePath, err)
		}
	}

	cfg := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AdminToken:      os.Getenv("TELEPROMPTR_ADMIN_TOKEN"),
		EncryptionKey:   os.Getenv("TELEPROMPTR_ENCRYPTION_KEY"),
		HTTPPort:        intEnvOrDefault("HTTP_PORT", DefaultHTTPPort),
		GRPCPort:        intEnvOrDefault("GRPC_PORT", DefaultGRPCPort),
		OTLPHTTPPort:    intEnvOrDefault("OTLP_HTTP_PORT", DefaultOTLPHTTP),
		BufferSize:      intEnvOrDefault("BUFFER_SIZE", DefaultBufferSize),
		BatchSize:       intEnvOrDefault("BATCH_SIZE", DefaultBatchSize),
		FlushIntervalMS: intEnvOrDefault("FLUSH_INTERVAL_MS", DefaultFlushMS),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration values are present and within
// acceptable ranges. It accumulates all errors and returns them joined.
func (c *Config) validate() error {
	var errs []string

	if c.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if c.AdminToken == "" {
		errs = append(errs, "TELEPROMPTR_ADMIN_TOKEN is required")
	}
	if c.EncryptionKey == "" {
		errs = append(errs, "TELEPROMPTR_ENCRYPTION_KEY is required")
	} else if len(c.EncryptionKey) < MinEncryptionKeyLen {
		errs = append(errs, fmt.Sprintf("TELEPROMPTR_ENCRYPTION_KEY must be at least %d bytes (got %d)", MinEncryptionKeyLen, len(c.EncryptionKey)))
	}
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		errs = append(errs, fmt.Sprintf("HTTP_PORT must be between 1 and 65535 (got %d)", c.HTTPPort))
	}
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		errs = append(errs, fmt.Sprintf("GRPC_PORT must be between 1 and 65535 (got %d)", c.GRPCPort))
	}
	if c.OTLPHTTPPort < 1 || c.OTLPHTTPPort > 65535 {
		errs = append(errs, fmt.Sprintf("OTLP_HTTP_PORT must be between 1 and 65535 (got %d)", c.OTLPHTTPPort))
	}
	if c.BufferSize < 1 {
		errs = append(errs, fmt.Sprintf("BUFFER_SIZE must be positive (got %d)", c.BufferSize))
	}
	if c.BatchSize < 1 {
		errs = append(errs, fmt.Sprintf("BATCH_SIZE must be positive (got %d)", c.BatchSize))
	}
	if c.FlushIntervalMS < 1 {
		errs = append(errs, fmt.Sprintf("FLUSH_INTERVAL_MS must be positive (got %d)", c.FlushIntervalMS))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// intEnvOrDefault reads an integer from the named environment variable. If the
// variable is unset or cannot be parsed as an integer, the provided default is
// returned.
func intEnvOrDefault(key string, defaultVal int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return val
}

// loadDotEnv reads a .env file and sets environment variables for any keys that
// are not already present in the environment. This ensures that real environment
// variables always take precedence. Lines starting with # and blank lines are
// ignored. Supports KEY=VALUE and KEY="VALUE" forms.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening .env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := parseDotEnvLine(line)
		if !ok {
			return fmt.Errorf("invalid .env syntax on line %d: %q", lineNum, line)
		}

		// Only set if not already present in the real environment.
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// parseDotEnvLine parses a single KEY=VALUE line from a .env file.
// It handles optional quoting of the value with double or single quotes.
// Returns the key, value, and whether parsing succeeded.
func parseDotEnvLine(line string) (key, value string, ok bool) {
	// Split on first '=' only.
	idx := strings.IndexByte(line, '=')
	if idx < 1 {
		return "", "", false
	}

	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])

	// Strip matching quotes.
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true
}
