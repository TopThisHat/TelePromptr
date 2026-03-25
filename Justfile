# =============================================================================
# TelePromptr Justfile
# =============================================================================
# Language-agnostic task runner for the TelePromptr monorepo.
# Chosen over Turborepo because `just` works equally well for Go and JS targets.
#
# Install: https://github.com/casey/just
#   brew install just        (macOS)
#   cargo install just       (Rust)
#
# Usage:
#   just              List all available targets
#   just dev          Run both API and web dev servers
#   just test         Run all tests
# =============================================================================

# List available targets when `just` is invoked with no arguments.
default:
    @just --list --unsorted

# -----------------------------------------------------------------------------
# Development
# -----------------------------------------------------------------------------

# Run Go API and Svelte dev server concurrently.
dev:
    just dev-api & just dev-web & wait

# Run the Go API in development mode with live reload.
dev-api:
    cd apps/api && go run .

# Run the SvelteKit frontend dev server.
dev-web:
    cd apps/web && npm run dev

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------

# Build both the Go API binary and Svelte web assets.
build: build-api build-web

# Build the Go API binary.
build-api:
    cd apps/api && go build -o ../../dist/api .

# Build the SvelteKit frontend for production.
build-web:
    cd apps/web && npm run build

# -----------------------------------------------------------------------------
# Test
# -----------------------------------------------------------------------------

# Run tests for both Go API and Svelte web.
test: test-api test-web

# Run Go tests with race detection and coverage.
test-api:
    cd apps/api && go test -race -cover ./...

# Run Svelte/JS tests.
test-web:
    cd apps/web && npm test

# -----------------------------------------------------------------------------
# Lint & Format
# -----------------------------------------------------------------------------

# Run linters for both Go and Svelte/JS.
lint: lint-api lint-web

# Run go vet on the API.
lint-api:
    cd apps/api && go vet ./...

# Run ESLint on the web frontend.
lint-web:
    cd apps/web && npm run lint

# Format Go source files.
fmt:
    cd apps/api && go fmt ./...

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------

# Run database migrations (placeholder -- tooling TBD).
migrate:
    @echo "TODO: run database migrations (goose, atlas, or similar)"

# -----------------------------------------------------------------------------
# Docker
# -----------------------------------------------------------------------------

# Start all docker compose services in the background.
docker-up:
    docker compose up -d

# Stop all docker compose services.
docker-down:
    docker compose down

# Stop all services and remove volumes (fresh start).
docker-nuke:
    docker compose down -v

# Rebuild and restart all docker compose services.
docker-rebuild:
    docker compose up -d --build
