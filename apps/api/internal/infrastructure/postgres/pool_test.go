package postgres

import (
	"context"
	"testing"
	"time"
)

func TestConfig_defaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   Config
		wantMin int32
		wantMax int32
		wantHP  time.Duration
	}{
		{
			name:    "all zero values get defaults",
			input:   Config{DatabaseURL: "postgres://localhost/test"},
			wantMin: 2,
			wantMax: 10,
			wantHP:  30 * time.Second,
		},
		{
			name:    "explicit values are preserved",
			input:   Config{DatabaseURL: "postgres://localhost/test", MinConns: 5, MaxConns: 20, HealthCheckPeriod: 1 * time.Minute},
			wantMin: 5,
			wantMax: 20,
			wantHP:  1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := tt.input
			cfg.defaults()

			if cfg.MinConns != tt.wantMin {
				t.Errorf("MinConns = %d, want %d", cfg.MinConns, tt.wantMin)
			}
			if cfg.MaxConns != tt.wantMax {
				t.Errorf("MaxConns = %d, want %d", cfg.MaxConns, tt.wantMax)
			}
			if cfg.HealthCheckPeriod != tt.wantHP {
				t.Errorf("HealthCheckPeriod = %v, want %v", cfg.HealthCheckPeriod, tt.wantHP)
			}
		})
	}
}

func TestNew_emptyURL(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected error for empty database URL, got nil")
	}
}

func TestToPgx5URL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "postgres:// prefix",
			input: "postgres://user:pass@localhost:5432/db?sslmode=disable",
			want:  "pgx5://user:pass@localhost:5432/db?sslmode=disable",
		},
		{
			name:  "postgresql:// prefix",
			input: "postgresql://user:pass@localhost:5432/db",
			want:  "pgx5://user:pass@localhost:5432/db",
		},
		{
			name:  "already pgx5:// is unchanged",
			input: "pgx5://user:pass@localhost:5432/db",
			want:  "pgx5://user:pass@localhost:5432/db",
		},
		{
			name:  "unknown scheme is unchanged",
			input: "mysql://user:pass@localhost:3306/db",
			want:  "mysql://user:pass@localhost:3306/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := toPgx5URL(tt.input)
			if got != tt.want {
				t.Errorf("toPgx5URL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
