// Command migrate is a CLI tool for running TelePromptr database migrations.
//
// Usage:
//
//	go run ./cmd/migrate -dir migrations -url postgres://user:pass@host:5432/db
//	go run ./cmd/migrate -dir migrations -url postgres://... -down
//	go run ./cmd/migrate -dir migrations -url postgres://... -version
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	// pgx/v5 database driver for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	// file source driver for loading SQL migration files.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("dir", "migrations", "path to migration files")
	url := flag.String("url", "", "database connection URL (required)")
	down := flag.Bool("down", false, "roll back all migrations")
	version := flag.Bool("version", false, "print current migration version")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "error: -url flag is required")
		flag.Usage()
		os.Exit(1)
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("resolving migrations dir: %v", err)
	}

	sourceURL := "file://" + absDir
	dbURL := toPgx5URL(*url)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatalf("creating migrator: %v", err)
	}
	defer m.Close()

	switch {
	case *version:
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("getting version: %v", err)
		}
		dirtyStr := ""
		if dirty {
			dirtyStr = " (dirty)"
		}
		fmt.Printf("version: %d%s\n", v, dirtyStr)

	case *down:
		fmt.Println("Rolling back all migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("Done.")

	default:
		fmt.Println("Applying migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("Done.")
	}
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
