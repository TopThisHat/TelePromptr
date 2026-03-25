package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// pgx/v5 database driver for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	// file source driver for loading SQL migration files.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateUp applies all pending database migrations from the given directory.
// The migrationsDir must be an absolute or relative path to the folder
// containing sequentially numbered .up.sql / .down.sql pairs.
//
// The databaseURL must be a valid PostgreSQL connection string. The function
// uses the pgx/v5 driver prefix so the URL is passed through as-is.
func MigrateUp(databaseURL, migrationsDir string) error {
	// golang-migrate expects "pgx5://" scheme for the pgx/v5 driver.
	sourceURL := "file://" + migrationsDir
	dbURL := toPgx5URL(databaseURL)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("postgres: creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("postgres: running migrations up: %w", err)
	}

	return nil
}

// MigrateDown rolls back all migrations. Use with care -- this drops tables.
func MigrateDown(databaseURL, migrationsDir string) error {
	sourceURL := "file://" + migrationsDir
	dbURL := toPgx5URL(databaseURL)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("postgres: creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("postgres: running migrations down: %w", err)
	}

	return nil
}

// toPgx5URL converts a standard postgres:// URL to the pgx5:// scheme
// expected by golang-migrate's pgx/v5 database driver.
func toPgx5URL(databaseURL string) string {
	const (
		prefixPostgres  = "postgres://"
		prefixPostgreql = "postgresql://"
		prefixPgx5      = "pgx5://"
	)

	switch {
	case len(databaseURL) >= len(prefixPostgres) && databaseURL[:len(prefixPostgres)] == prefixPostgres:
		return prefixPgx5 + databaseURL[len(prefixPostgres):]
	case len(databaseURL) >= len(prefixPostgreql) && databaseURL[:len(prefixPostgreql)] == prefixPostgreql:
		return prefixPgx5 + databaseURL[len(prefixPostgreql):]
	default:
		return databaseURL
	}
}
