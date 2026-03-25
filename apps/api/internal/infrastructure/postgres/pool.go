// Package postgres provides PostgreSQL connectivity for the TelePromptr API.
// It manages connection pooling via pgxpool and database migrations via
// golang-migrate.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the settings required to establish a PostgreSQL connection pool.
type Config struct {
	// DatabaseURL is a PostgreSQL connection string
	// (e.g. "postgres://user:pass@host:5432/db?sslmode=disable").
	DatabaseURL string

	// MinConns is the minimum number of connections kept open in the pool.
	// A sensible default for development is 2.
	MinConns int32

	// MaxConns is the maximum number of connections the pool will open.
	// A sensible default for development is 10.
	MaxConns int32

	// HealthCheckPeriod controls how often idle connections are checked.
	// Defaults to 30 seconds if zero.
	HealthCheckPeriod time.Duration
}

// defaults fills in zero-valued fields with sensible development defaults.
func (c *Config) defaults() {
	if c.MinConns == 0 {
		c.MinConns = 2
	}
	if c.MaxConns == 0 {
		c.MaxConns = 10
	}
	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = 30 * time.Second
	}
}

// New creates a new pgxpool.Pool from the provided Config.
// It validates the configuration, applies defaults, and pings the database
// before returning the pool.
func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("postgres: database URL must not be empty")
	}

	cfg.defaults()

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: parsing database URL: %w", err)
	}

	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: creating connection pool: %w", err)
	}

	// Verify connectivity before returning.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: pinging database: %w", err)
	}

	return pool, nil
}
