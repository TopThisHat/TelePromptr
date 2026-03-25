## Context

TelePromptr is a greenfield LLM telemetry platform. There is no existing codebase — we are building from scratch as a monorepo with a Go 1.25 backend API and Svelte 5 + Tailwind CSS 4 frontend.

The platform targets engineering teams building LLM-powered applications who need observability, prompt management, and evaluation in a single self-hosted tool. The primary integration point is OpenTelemetry — teams instrument their apps with standard OTel SDKs and point traces at TelePromptr.

**Constraints:**
- Go 1.25 backend, Svelte 5 frontend, Tailwind CSS 4
- Clean architecture with dependency injection
- Must handle trace ingestion without blocking the REST API
- Single-instance self-hosted deployment for v1
- Phased delivery: MVP first (traces + prompts + analytics + projects), then evaluation, then advanced features
- The prompt-to-trace linkage must ship in MVP — this is the core differentiator

## Goals / Non-Goals

**Goals:**
- Ingest OpenTelemetry traces via OTLP gRPC and HTTP, including span events and resource attributes
- Store and display actual LLM input/output content (prompt text, completion text)
- Provide prompt versioning linked to traces via span attributes — in the MVP
- Enable systematic LLM output evaluation with curated datasets (Phase 2)
- Deliver real-time analytics with incremental rollups and webhook alerting
- Support project-level data isolation with admin-mode UI access
- Run locally with a single `docker compose up`

**Non-Goals:**
- Building a custom OpenTelemetry SDK or collector — we receive standard OTLP data
- Real-time streaming/WebSocket push for dashboards (polling is acceptable for v1)
- Fine-tuning or model hosting — this is observability only
- SSO/SAML/multi-user authentication (single-user admin for v1)
- Horizontal auto-scaling — single-instance deployment for v1
- Mobile clients
- Columnar store (ClickHouse) — deferred until PostgreSQL query performance becomes a bottleneck
- Log or metric ingestion in Phase 1 (traces only; logs planned for Phase 3)

## Decisions

### 1. Monorepo Structure with Justfile

```
/
├── apps/
│   ├── api/          # Go 1.25 backend
│   └── web/          # Svelte 5 + Tailwind CSS 4 frontend (SvelteKit)
├── Justfile          # Language-agnostic task runner
├── docker-compose.yml
└── go.work          # Go workspace (for future multi-module)
```

No `packages/` directory. All frontend code lives in `apps/web/src/lib/` until a second consumer exists.

**Rationale:** `just` is a language-agnostic task runner that works equally well for Go and JS targets. Turborepo was considered but rejected — it only understands JS packages and provides no value for Go. With one Go app and one Svelte app, there is no JS package dependency graph to optimize.

**Alternatives considered:**
- Turborepo: only benefits JS packages; would need a separate Makefile for Go anyway
- Makefile: viable, but `just` has better syntax, per-recipe documentation, and cross-platform support

### 2. Go Backend Architecture — Clean Architecture with Manual DI

```
apps/api/
├── cmd/server/          # Entry point, manual DI wiring (~60 lines)
├── internal/
│   ├── domain/          # Domain entities and repository interfaces (no external deps)
│   │   ├── trace/
│   │   ├── prompt/
│   │   ├── evaluation/
│   │   └── project/
│   ├── application/     # Use cases / service layer
│   │   ├── trace/
│   │   ├── prompt/
│   │   ├── evaluation/
│   │   └── project/
│   ├── infrastructure/  # External adapters (DB, OTLP, etc.)
│   │   ├── postgres/
│   │   ├── otlp/
│   │   └── config/
│   └── interfaces/      # HTTP/gRPC handlers
│       ├── rest/
│       └── grpc/
├── pkg/                 # Shared utilities (logging, errors, middleware)
├── go.mod
└── go.sum
```

**Rationale:** Clean architecture enforces dependency inversion — domain logic has zero knowledge of databases or HTTP. All dependencies flow inward.

**DI approach:** Manual constructor injection in `cmd/server/main.go`. At v1, the dependency graph is ~25 constructors and ~60 lines of wiring. Explicit, readable, zero codegen. Wire can be adopted later — the constructor-based design is already Wire-compatible.

### 3. PostgreSQL as Sole Data Store

PostgreSQL handles all data:
- **Relational data:** Projects, API keys, prompts, prompt versions, evaluations, evaluation datasets, LLM providers
- **Trace data:** `trace_spans` table with native table partitioning by date (monthly partitions)
- **Span events:** `span_events` JSONB column on trace_spans for events including LLM input/output content
- **Search:** `search_text tsvector` generated column with GIN index for full-text search across span names and key attributes
- **Analytics:** Incremental rollup table (`analytics_rollups`) updated by a background goroutine — not materialized views

**Why not materialized views:** PostgreSQL `REFRESH MATERIALIZED VIEW` takes an exclusive lock (blocks reads), recomputes the entire result set (not incremental), and gets slower linearly as data grows. Within weeks of production use, refresh times become prohibitive. An incremental rollup table processes only new spans since the last watermark, making it O(new_data) always.

**When to add ClickHouse:** When analytical queries on trace data consistently exceed acceptable latency thresholds (e.g., dashboard queries > 2 seconds). The `TraceRepository` interface abstracts storage — ClickHouse can be added as a one-adapter change.

**Alternatives considered:**
- PostgreSQL + ClickHouse dual store: rejected for v1 — doubles operational complexity for theoretical scale
- Materialized views: rejected — exclusive locks and O(all_data) refresh are unacceptable
- TimescaleDB: viable but adds extension dependency; native partitioning + rollup table is sufficient

### 4. Trace Data Model — Full OTel Support

The `trace_spans` table stores the complete OpenTelemetry span structure:

**Core OTel fields:**
- `trace_id`, `span_id`, `parent_span_id` — trace correlation
- `name` — span name
- `span_kind` — CLIENT, SERVER, INTERNAL, PRODUCER, CONSUMER
- `status_code` — OK (0), ERROR (2)
- `status_message` — error description text
- `start_time`, `end_time`, `duration_ms` — timing
- `attributes` — JSONB for arbitrary span attributes
- `events` — JSONB array of span events (timestamps, names, attributes)

**Resource attributes (denormalized):**
- `service_name` — from `service.name` resource attribute
- `service_version` — from `service.version`
- `deployment_environment` — from `deployment.environment`
- `resource_attributes` — JSONB for all resource attributes

**LLM-specific extracted columns (indexed):**
- `model`, `provider` — from `gen_ai.request.model`, `gen_ai.system`
- `input_tokens`, `output_tokens`, `total_tokens` — from `gen_ai.usage.*`
- `cost` — computed from token counts + project cost-per-token rates
- `input_content`, `output_content` — TEXT, extracted from span events `gen_ai.content.prompt` and `gen_ai.content.completion`

**TelePromptr-specific columns (indexed):**
- `prompt_id`, `prompt_version`, `prompt_execution_id` — from `telepromptr.prompt.*` attributes
- `session_id` — from `telepromptr.session.id` attribute

**Search:**
- `search_text tsvector` — generated column concatenating span name, service_name, model, status_message. GIN indexed for full-text search via `@@` operator.

**Why denormalize resource attributes:** Resource attributes describe the service, not the span. In pure OTel, they live at the top of the `ExportTraceServiceRequest`. Denormalizing `service_name`, `service_version`, `deployment_environment` onto each span record avoids a join and enables direct filtering. The full resource set is stored in `resource_attributes` JSONB for completeness.

**Why extract input/output content:** LLM span events contain the actual prompt and completion text. Without extracting these, TelePromptr can show metadata (model, tokens, cost) but not what was actually sent to or received from the LLM. This makes trace exploration useless for debugging and evaluation impossible (scorers need text to score). Extracting to TEXT columns enables display, search, and evaluation.

### 5. OTLP Ingestion from Protobuf Definitions

Implement the OTLP gRPC and HTTP services directly from the OpenTelemetry protobuf definitions (`opentelemetry-proto`), not by embedding the OTel Collector's receiver libraries.

The OTLP spec is intentionally small:
- One gRPC service: `TraceService` with one method `Export(ExportTraceServiceRequest) returns (ExportTraceServiceResponse)`
- One HTTP endpoint: `POST /v1/traces` accepting JSON or protobuf

**Rationale:** The OTel Collector's receiver libraries are designed for the Collector binary — massive transitive dependency tree, Collector-specific lifecycle management, semi-internal APIs. Implementing from protobuf gives us ~200 lines of gRPC handler code, full control, and minimal dependencies.

### 6. Ingestion Pipeline with Backpressure

```
OTLP Request → Auth (SHA-256 key lookup) → Extract (resource attrs, LLM attrs, events, content) → Buffer (bounded, in-memory) → Batch Writer (PostgreSQL)
```

**Bounded buffer:**
- Configurable max size (default: 100,000 spans)
- When full: return OTLP throttling response (gRPC `RESOURCE_EXHAUSTED` / HTTP 429 with `Retry-After` header)
- Per-project rate limits configurable in project settings (spans/second, default: unlimited)

**Batch writer:**
- Configurable batch size (default: 1,000 spans) and flush interval (default: 5 seconds)
- Whichever threshold is reached first triggers a flush
- After each flush, updates the analytics rollup watermark

**Why backpressure matters:** Without a bounded buffer, if trace volume exceeds the flush rate, the buffer grows unbounded until OOM. Returning 429 lets the OTel SDK's retry mechanism handle backoff correctly.

### 7. API Key Authentication — SHA-256

API keys are high-entropy random values (32 bytes, hex-encoded). Unlike passwords, they don't benefit from bcrypt's intentional slowness. Using bcrypt for API key verification requires O(n) sequential comparisons (~100ms each) across all stored keys.

**Approach:**
- Generate key: `tp_proj_` + 32 random bytes hex-encoded (e.g., `tp_proj_a1b2c3d4...`)
- Store: `sha256(full_key)` as the lookup hash, plus `key_prefix` (first 8 chars after prefix) for display
- Verify: compute `sha256(presented_key)`, look up by hash (indexed, O(1))
- Key is shown once at creation time, never retrievable after

**Why SHA-256 over bcrypt:** API keys are 256-bit random values — brute force is infeasible regardless of hash speed. bcrypt's slowness protects weak passwords from offline dictionary attacks, which doesn't apply here. SHA-256 is collision-resistant and enables O(1) indexed lookup instead of O(n) sequential comparison.

### 8. Dual Auth Mode — Admin + API Key

The web UI needs to call the backend API but has no API key (single-user admin, no login). The REST API needs API key auth for programmatic access. These are contradictory without a dual-auth design.

**Approach:**
- **API key auth:** `Authorization: Bearer tp_proj_...` header → scopes to key's project
- **Admin mode:** Requests from the SvelteKit server (via server-side `load` functions) include an `X-Admin-Token` header with a shared secret configured via environment variable
- Admin mode allows access to all projects, determined by `X-Project-Id` header or route params
- Admin mode is how the web UI accesses data — the browser talks to SvelteKit, SvelteKit talks to the Go API server-to-server

**Why SvelteKit server-side:** The browser never calls the Go API directly. SvelteKit's `+page.server.ts` load functions fetch data from the Go API using the admin token (set from environment config). This means:
- No auth secrets in the browser
- The Go API can restrict admin endpoints to internal network / localhost
- SvelteKit handles SSR and can pre-render data before sending HTML to the browser
- API key auth remains clean for external programmatic clients

**Admin token lifecycle:**
- Set via `TELEPROMPTR_ADMIN_TOKEN` environment variable
- If not set, auto-generated on first startup and printed to stdout
- Stored in SvelteKit's server-side environment (never sent to browser)

### 9. Frontend Architecture — Svelte 5 with SvelteKit SSR

```
apps/web/
├── src/
│   ├── lib/
│   │   ├── components/    # All reusable UI components
│   │   ├── server/        # Server-only: API client with admin token
│   │   ├── api/           # Client-side API helpers (form actions, etc.)
│   │   └── utils/         # Helpers, formatters, types
│   ├── routes/
│   │   ├── (app)/         # App layout (all pages)
│   │   │   ├── traces/
│   │   │   ├── prompts/
│   │   │   ├── analytics/
│   │   │   ├── evaluations/
│   │   │   └── settings/
│   │   └── +layout.svelte
│   └── app.html
├── static/
├── svelte.config.js
└── vite.config.ts
```

**Key patterns:**
- `$lib/server/api.ts` — server-only module that makes authenticated calls to the Go API using the admin token. Used in `+page.server.ts` load functions.
- `$lib/components/` — all UI components. No separate packages until a second consumer exists.
- Svelte 5 runes (`$state`, `$derived`, `$effect`) for reactive state. No stores directory.
- Tailwind CSS 4 for styling.

### 10. Session and Conversation Support

LLM applications are overwhelmingly conversational. TelePromptr supports sessions via a span attribute convention:

- Applications set `telepromptr.session.id` as a span attribute on LLM call spans
- The ingestion pipeline extracts this to an indexed `session_id` column
- Trace exploration supports filtering and grouping by session
- The trace detail view shows session context (other traces in the same session)

This is a lightweight convention (one span attribute) that enables powerful UX (see the full conversation flow).

### 11. Prompt-to-Trace Linkage

When an application uses a TelePromptr prompt via `POST /api/v1/prompts/:id/render`:
1. The response includes `prompt_execution_id` (UUID), `prompt_id`, and `version`
2. The application sets span attributes: `telepromptr.prompt.id`, `telepromptr.prompt.version`, `telepromptr.prompt.execution_id`
3. During ingestion, TelePromptr extracts these to indexed columns
4. Evaluations can filter traces by prompt + version

**Phase 2 SDKs** (Python + TypeScript) will automate steps 1-2 into a single function call. In Phase 1, applications integrate manually (3 span attributes).

**Passive prompt management** is also supported — users can create/version prompts in TelePromptr purely for record-keeping without using the render API. The linkage is optional, not required.

### 12. Analytics — Incremental Rollup Table

```sql
CREATE TABLE analytics_rollups (
    project_id UUID NOT NULL,
    hour_bucket TIMESTAMPTZ NOT NULL,
    model TEXT NOT NULL,
    service_name TEXT NOT NULL,
    span_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    total_input_tokens BIGINT DEFAULT 0,
    total_output_tokens BIGINT DEFAULT 0,
    total_cost NUMERIC(12,6) DEFAULT 0,
    latency_sum_ms BIGINT DEFAULT 0,
    latency_p50_ms DOUBLE PRECISION,
    latency_p90_ms DOUBLE PRECISION,
    latency_p95_ms DOUBLE PRECISION,
    latency_p99_ms DOUBLE PRECISION,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (project_id, hour_bucket, model, service_name)
);
```

A background goroutine runs on a configurable interval (default: 60 seconds):
1. Reads a `rollup_watermark` value (last processed span timestamp)
2. Queries new spans since the watermark
3. Computes aggregates and upserts into `analytics_rollups`
4. Updates the watermark

This is O(new_data) always, never locks, and supports concurrent reads from the analytics API.

**Percentile calculation:** For each rollup bucket, we compute approximate percentiles using PostgreSQL's `percentile_cont` on the batch of new spans. The rollup stores the latest computed percentiles. For exact percentiles across arbitrary time ranges, the analytics API falls back to querying `trace_spans` directly (slower but accurate).

### 13. LLM Provider Configuration

Both the evaluation framework (LLM-as-judge) and the prompt playground need to call LLMs. Provider configuration is stored per-project:

```sql
CREATE TABLE llm_providers (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id),
    provider TEXT NOT NULL,          -- 'openai', 'anthropic', 'custom'
    display_name TEXT,
    api_key_encrypted BYTEA NOT NULL,
    base_url TEXT,                   -- for custom/self-hosted providers
    models JSONB,                    -- available model list
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, provider)
);
```

**Encryption:** Provider API keys are encrypted at rest using AES-256-GCM with a key derived from the `TELEPROMPTR_ENCRYPTION_KEY` environment variable. This protects against database dump exposure.

The schema ships in Phase 1 (so migrations don't break later). The configuration UI and actual LLM calling logic ship in Phase 2.

### 14. Docker Compose for Local Development

```yaml
services:
  postgres:
    image: postgres:17
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: telepromptr
      POSTGRES_USER: telepromptr
      POSTGRES_PASSWORD: telepromptr
    volumes: [postgres_data:/var/lib/postgresql/data]

  api:
    build: ./apps/api
    ports: ["8080:8080", "4317:4317", "4318:4318"]
    environment:
      DATABASE_URL: postgres://telepromptr:telepromptr@postgres:5432/telepromptr
      TELEPROMPTR_ADMIN_TOKEN: dev-admin-token
      TELEPROMPTR_ENCRYPTION_KEY: dev-encryption-key-min-32-bytes!!
    depends_on: [postgres]

  web:
    build: ./apps/web
    ports: ["5173:5173"]
    environment:
      API_URL: http://api:8080
      ADMIN_TOKEN: dev-admin-token
    depends_on: [api]
```

`docker compose up` starts everything. For development, `just dev` runs Go API and Svelte dev server locally with hot reload, connecting to Dockerized PostgreSQL.

## Risks / Trade-offs

**[PostgreSQL trace scale ceiling]** → PostgreSQL with partitioning will hit analytical query limits at very high trace volumes. Mitigation: `TraceRepository` interface abstracts storage — ClickHouse is a one-adapter change. Monitor query latency and migrate when evidence justifies it.

**[Single-user admin security]** → The web UI relies on network-level access control + admin token. Mitigation: admin token prevents unauthenticated API access; document that the instance should be behind a firewall or VPN. Multi-user auth in a later phase doesn't require restructuring.

**[OTLP protobuf implementation maintenance]** → We own the deserialization code. Mitigation: OTLP spec is stable and backwards-compatible; changes are infrequent.

**[Input/output content storage size]** → Storing full prompt/completion text increases storage significantly. Mitigation: content columns are nullable; extraction is best-effort. Future retention policies can truncate or drop content after a TTL.

**[Phased SDK delivery]** → Phase 1 requires manual span attribute setup (3 attributes). This is friction. Mitigation: clear documentation with copy-paste code snippets for Python/TypeScript/Go. SDKs arrive in Phase 2.

**[Incremental rollup percentile approximation]** → Rolling up percentiles incrementally loses precision compared to computing over full datasets. Mitigation: for exact percentiles, fall back to direct `trace_spans` query. The rollup is for fast dashboard rendering; detailed analysis uses exact queries.

## Open Questions

- **Trace retention:** Configurable TTL per project with a background job that drops old partitions. Deferred to post-MVP.
- **Evaluation LLM provider defaults:** Whether to ship with default model configurations or require explicit setup. Likely require setup for security (no default API keys).
- **Tailwind CSS version:** User requested Tailwind 5. Latest stable is Tailwind CSS v4. Using v4. If v5 ships before implementation, migration should be straightforward.
- **Content extraction completeness:** Different LLM SDKs put prompt/completion in different span event names. The extractor should be configurable or use a priority list of known conventions (`gen_ai.content.prompt`, `llm.content.prompt`, event body text, etc.).
