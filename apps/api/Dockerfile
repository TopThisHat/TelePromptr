# =============================================================================
# Multi-stage build for TelePromptr Go API
# =============================================================================
# Build:  docker build -t telepromptr-api ./apps/api
# Run:    docker run -p 8080:8080 -p 4317:4317 -p 4318:4318 telepromptr-api

# ---------------------------------------------------------------------------
# Stage 1 - Build the Go binary
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile a static, stripped binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/server

# ---------------------------------------------------------------------------
# Stage 2 - Minimal runtime image
# ---------------------------------------------------------------------------
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /api /usr/local/bin/api

# REST API, OTLP gRPC, OTLP HTTP
EXPOSE 8080 4317 4318

ENTRYPOINT ["api"]
