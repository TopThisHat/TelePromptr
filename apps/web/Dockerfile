# =============================================================================
# Multi-stage build for TelePromptr SvelteKit frontend (adapter-node)
# =============================================================================
# Build:  docker build -t telepromptr-web ./apps/web
# Run:    docker run -p 3000:3000 telepromptr-web

# ---------------------------------------------------------------------------
# Stage 1 - Install deps and build the SvelteKit app
# ---------------------------------------------------------------------------
FROM node:22-alpine AS builder

WORKDIR /app

# Cache dependency installation in a separate layer
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build with adapter-node
COPY . .
RUN npm run build

# ---------------------------------------------------------------------------
# Stage 2 - Minimal runtime image
# ---------------------------------------------------------------------------
FROM node:22-alpine

WORKDIR /app

# Copy the adapter-node output and package manifest
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
COPY --from=builder /app/package-lock.json ./

# Install only production dependencies
RUN npm ci --omit=dev

EXPOSE 3000
ENV PORT=3000

CMD ["node", "build"]
