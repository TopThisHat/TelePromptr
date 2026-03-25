## 1. Monorepo Scaffolding

- [ ] 1.1 Create root directory structure with `apps/api/` and `apps/web/`, root `go.work`, and `.gitignore`
- [ ] 1.2 Create `Justfile` with targets: `dev`, `build`, `test`, `lint`, `migrate`, `docker-up`, `docker-down`
- [ ] 1.3 Create `docker-compose.yml` with PostgreSQL 17 service (telepromptr db/user), mapped ports, persistent volume, and environment variables
- [ ] 1.4 Initialize `apps/api` as a Go 1.25 module with clean architecture directory skeleton (`cmd/server`, `internal/{domain,application,infrastructure,interfaces}`, `pkg/`)
- [ ] 1.5 Initialize `apps/web` as a SvelteKit project with Svelte 5, Tailwind CSS 4, TypeScript, and `adapter-node` for SSR

## 2. Go Backend Foundation

- [ ] 2.1 Implement configuration loading (env vars + `.env` file): DB URL, OTLP ports, admin token, encryption key, buffer sizes, flush intervals
- [ ] 2.2 Implement structured logging using `slog` with JSON output and request correlation via context
- [ ] 2.3 Implement HTTP router (chi or stdlib) with middleware: CORS, request ID, logging, panic recovery
- [ ] 2.4 Implement dual auth middleware: API key auth via SHA-256 lookup (Bearer token) + admin token auth via `X-Admin-Token` header with `X-Project-Id` for project scoping
- [ ] 2.5 Implement standard JSON response helpers: success with pagination wrapper, error with code/message/details
- [ ] 2.6 Implement request validation middleware with structured error details
- [ ] 2.7 Wire up manual dependency injection in `cmd/server/main.go`: config → DB pool → repos → services → handlers → router → server

## 3. PostgreSQL Database Layer

- [ ] 3.1 Set up PostgreSQL connection pool (`pgxpool`) and migration framework (`golang-migrate`) with `just migrate` target
- [ ] 3.2 Create migration: `projects` table (id UUID PK, name, description, settings JSONB, created_at, updated_at)
- [ ] 3.3 Create migration: `api_keys` table (id UUID PK, project_id FK, key_hash BYTEA for SHA-256, key_prefix TEXT, created_at, last_used_at, revoked_at) with unique index on key_hash
- [ ] 3.4 Create migration: `trace_spans` table with monthly date partitioning — columns: trace_id, span_id, parent_span_id, name, span_kind, status_code, status_message, start_time, end_time, duration_ms, attributes JSONB, events JSONB, resource_attributes JSONB, service_name, service_version, deployment_environment, model, provider, input_tokens, output_tokens, total_tokens, cost, input_content TEXT, output_content TEXT, prompt_id, prompt_version, prompt_execution_id, session_id, project_id FK, search_text TSVECTOR (generated)
- [ ] 3.5 Create indexes on `trace_spans`: (project_id, start_time), (trace_id), (model), (service_name), (session_id), (prompt_id, prompt_version), GIN on search_text, BRIN on start_time
- [ ] 3.6 Create migration: `analytics_rollups` table (project_id, hour_bucket, model, service_name, span_count, error_count, total_input_tokens, total_output_tokens, total_cost, latency percentiles, updated_at) with composite PK + `rollup_watermark` tracking table
- [ ] 3.7 Create migration: `prompts` table (id, project_id FK, name, description, created_at), `prompt_versions` table (id, prompt_id FK, version INT, body TEXT, status TEXT, created_at), `prompt_tags` table (prompt_id FK, tag TEXT) with unique constraint on (project_id, name)
- [ ] 3.8 Create migration: `llm_providers` table (id, project_id FK, provider, display_name, api_key_encrypted BYTEA, base_url, models JSONB, created_at) — schema only for Phase 1, UI in Phase 2

## 4. Project Management & Auth (Phase 1)

- [ ] 4.1 Implement `Project` domain entity and `ProjectRepository` interface
- [ ] 4.2 Implement `ProjectRepository` PostgreSQL adapter
- [ ] 4.3 Implement `ProjectService` application service (create, list, get, update, update settings)
- [ ] 4.4 Implement API key generation (`tp_proj_` + 32 random hex bytes), SHA-256 hashing, prefix extraction
- [ ] 4.5 Implement `APIKeyRepository` PostgreSQL adapter (create, list by project, find by SHA-256 hash, revoke, update last_used)
- [ ] 4.6 Implement REST handlers: `POST /api/v1/projects`, `GET /api/v1/projects`, `GET /api/v1/projects/:id`, `PUT /api/v1/projects/:id`
- [ ] 4.7 Implement REST handlers: `POST /api/v1/projects/:id/api-keys`, `GET /api/v1/projects/:id/api-keys`, `DELETE /api/v1/projects/:id/api-keys/:keyId`
- [ ] 4.8 Implement project settings endpoints: cost-per-token rates CRUD, model display names CRUD, rate limit configuration

## 5. Trace Ingestion (Phase 1)

- [ ] 5.1 Add `opentelemetry-proto` Go protobuf definitions as a dependency; generate Go types for `TraceService`
- [ ] 5.2 Implement OTLP gRPC server: `TraceService.Export` handler on configurable port (default 4317) with API key extraction from gRPC metadata → SHA-256 auth
- [ ] 5.3 Implement OTLP HTTP handler: `POST /v1/traces` with JSON + protobuf content type support and Bearer token auth
- [ ] 5.4 Implement resource attribute extractor: extract `service.name`, `service.version`, `deployment.environment` from request-level resource; store full resource as JSONB
- [ ] 5.5 Implement span event extractor: store full events array as JSONB; extract `input_content` and `output_content` from `gen_ai.content.prompt`/`gen_ai.content.completion` events (with fallback to span attributes)
- [ ] 5.6 Implement LLM metadata extractor: extract `model`, `provider` from `gen_ai.*` attributes; compute `total_tokens` and `cost` from token counts + project cost rates
- [ ] 5.7 Implement TelePromptr attribute extractor: extract `prompt_id`, `prompt_version`, `prompt_execution_id`, `session_id` from `telepromptr.*` span attributes
- [ ] 5.8 Implement span kind and status extraction: map OTel `SpanKind` enum to string, extract `status_code` and `status_message`
- [ ] 5.9 Implement tsvector population: generate `search_text` from span name + service_name + model + status_message
- [ ] 5.10 Implement bounded in-memory span buffer (configurable max, default 100k spans) with backpressure: return gRPC `RESOURCE_EXHAUSTED` / HTTP 429 with `Retry-After` when full
- [ ] 5.11 Implement per-project rate limiter: token bucket per project_id, configurable via project settings, returns throttling response when exceeded
- [ ] 5.12 Implement PostgreSQL batch span writer: flush on batch size (default 1000) or interval (default 5s), whichever comes first; update rollup watermark after flush
- [ ] 5.13 Write integration tests: send OTLP traces via gRPC and HTTP, verify all extracted fields (resource attrs, events, content, LLM metadata, span kind/status, search text, session, prompt linkage, backpressure behavior)

## 6. Trace Exploration API (Phase 1)

- [ ] 6.1 Implement `TraceRepository` interface and PostgreSQL adapter with: time range filtering (partition-aware), service_name filter, model filter, status filter, session_id filter, prompt_id/version filter, full-text search via tsvector, duration range filter, pagination
- [ ] 6.2 Implement trace tree reconstruction: query spans by trace_id, build parent-child hierarchy, compute trace summary (duration, span count, error status, service list, models used)
- [ ] 6.3 Implement session query: list distinct sessions (with trace count, time range), get all traces in a session ordered by start_time
- [ ] 6.4 Implement `TraceService` application service (list, get detail, search, list sessions, get session traces, export)
- [ ] 6.5 Implement REST handlers: `GET /api/v1/traces` (list with all filters), `GET /api/v1/traces/:traceId` (full detail with spans, events, content)
- [ ] 6.6 Implement REST handlers: `GET /api/v1/sessions` (list sessions), `GET /api/v1/sessions/:sessionId/traces` (traces in session)
- [ ] 6.7 Implement trace export: `GET /api/v1/traces/:traceId?format=otlp` returning OTLP-compatible JSON

## 7. Analytics API (Phase 1)

- [ ] 7.1 Implement analytics rollup background goroutine: runs every 60s, queries new spans since watermark, computes aggregates per (project, hour, model, service_name), upserts into `analytics_rollups`, updates watermark
- [ ] 7.2 Implement percentile computation in rollup: use PostgreSQL `percentile_cont` on each batch for p50/p90/p95/p99 latency
- [ ] 7.3 Implement `AnalyticsRepository` interface and PostgreSQL adapter querying `analytics_rollups` with time range, granularity, model/service grouping
- [ ] 7.4 Implement cost analytics endpoint: `GET /api/v1/analytics/cost` with time bucketing and model + service breakdown
- [ ] 7.5 Implement latency analytics endpoint: `GET /api/v1/analytics/latency` with rollup percentiles (fast) and `?exact=true` fallback to trace_spans query
- [ ] 7.6 Implement token analytics endpoint: `GET /api/v1/analytics/tokens` with input/output/total by time bucket
- [ ] 7.7 Implement error analytics endpoint: `GET /api/v1/analytics/errors` with error rate + count by time bucket, plus recent error spans with status_message and trace links

## 8. Prompt Management API (Phase 1)

- [ ] 8.1 Implement `Prompt` and `PromptVersion` domain entities with lifecycle state machine (draft → active → archived) and `{{variable}}` template parsing
- [ ] 8.2 Implement `PromptRepository` PostgreSQL adapter (CRUD, versioning, tag queries, find active version, find by project + name uniqueness)
- [ ] 8.3 Implement `PromptService` application service: create prompt, create version, activate version (auto-archive previous), archive version, reject edits to non-draft, tag management
- [ ] 8.4 Implement prompt template rendering: parse `{{variable}}` placeholders, substitute with provided values, validate all variables present, generate `prompt_execution_id` UUID on render
- [ ] 8.5 Implement REST handlers: `POST /api/v1/prompts`, `GET /api/v1/prompts`, `GET /api/v1/prompts/:id`
- [ ] 8.6 Implement REST handlers: `POST /api/v1/prompts/:id/versions`, `GET /api/v1/prompts/:id/versions/:version`, `PATCH /api/v1/prompts/:id/versions/:version` (activate/archive)
- [ ] 8.7 Implement REST handler: `POST /api/v1/prompts/:id/render` and `POST /api/v1/prompts/:id/versions/:version/render` — return rendered text + prompt_execution_id + prompt_id + version
- [ ] 8.8 Implement prompt tagging endpoints: add/remove tags, tag-filtered listing

## 9. Frontend Foundation (Phase 1)

- [ ] 9.1 Set up SvelteKit routing with `(app)` layout group, `+layout.server.ts` that loads project list via admin token API
- [ ] 9.2 Implement `$lib/server/api.ts`: server-only module that makes authenticated requests to Go API using `ADMIN_TOKEN` env var and `X-Admin-Token` + `X-Project-Id` headers
- [ ] 9.3 Implement app shell layout: sidebar nav (Traces, Sessions, Prompts, Analytics, Settings), header with project selector dropdown, content area
- [ ] 9.4 Set up Tailwind CSS 4 theme: color palette, typography, spacing, dark mode support
- [ ] 9.5 Implement shared UI components: data table with sortable columns and pagination, loading skeleton, error alert, empty state, time range picker (presets + custom), search input with debounce

## 10. Trace Exploration UI (Phase 1)

- [ ] 10.1 Implement traces list page (`/traces`): data table with columns (trace ID, root span, service, duration, time, status, model), pagination, loading states
- [ ] 10.2 Implement trace filter bar: time range selector, service dropdown, model dropdown, status toggle, duration range, session filter, prompt filter, full-text search input
- [ ] 10.3 Implement trace detail page (`/traces/[id]`): span waterfall visualization with timing bars, parent-child nesting, span kind indicators, color-coded status
- [ ] 10.4 Implement span detail panel: expandable side panel showing attributes table, events list, status (code + message), timing, resource info (service, version, environment)
- [ ] 10.5 Implement LLM span section: dedicated panel showing model, provider, input/output tokens, cost, latency, actual prompt text (input_content), actual completion text (output_content) with syntax highlighting, prompt linkage info with link to prompt version
- [ ] 10.6 Implement session view (`/sessions`): session list with trace count and time range; session detail showing all traces in chronological order with conversation flow
- [ ] 10.7 Implement trace comparison: select two traces, display waterfalls side by side with duration/span count diff highlighting
- [ ] 10.8 Implement trace export: download button for selected traces as JSON

## 11. Prompt Management UI (Phase 1)

- [ ] 11.1 Implement prompt list page (`/prompts`): table with name, active version, tags, trace count; search and tag filter controls
- [ ] 11.2 Implement prompt creation form: name, description, template editor with `{{variable}}` syntax highlighting and variable extraction preview
- [ ] 11.3 Implement prompt detail page (`/prompts/[id]`): version history timeline, diff viewer between versions, lifecycle status badges (draft/active/archived)
- [ ] 11.4 Implement lifecycle controls: activate/archive buttons with confirmation dialogs, version status indicators
- [ ] 11.5 Implement prompt render preview: variable input form, rendered output display, copy-to-clipboard for span attribute integration code (Python/TypeScript/Go snippets)
- [ ] 11.6 Implement prompt-linked traces: "View Traces" link on each version that navigates to trace explorer pre-filtered by prompt_id + version, with trace count badge

## 12. Analytics Dashboard UI (Phase 1)

- [ ] 12.1 Implement dashboard page (`/analytics`) with unified time range controls (presets + custom range picker)
- [ ] 12.2 Implement cost analytics panel: time-series chart, model breakdown stacked chart, service breakdown table, summary cards (total cost, cost/day trend)
- [ ] 12.3 Implement latency analytics panel: percentile display cards (p50/p90/p95/p99), histogram, model comparison table
- [ ] 12.4 Implement token usage panel: time-series chart (input/output/total lines), model breakdown table
- [ ] 12.5 Implement error rate panel: time-series chart with error rate %, error list with status_message, service, model, and trace links

## 13. Project Management UI (Phase 1)

- [ ] 13.1 Implement project list and creation page (`/settings/projects`)
- [ ] 13.2 Implement project settings page: cost-per-token rate editor (per model), model display name configuration, rate limit setting
- [ ] 13.3 Implement API key management: generate (show full key once with copy button), list (prefix + dates + status), revoke with confirmation dialog

## 14. Phase 1 Integration & Polish

- [ ] 14.1 Add `Dockerfile` for Go API (multi-stage build) and `Dockerfile` for SvelteKit web (node-based SSR), update `docker-compose.yml` with all three services
- [ ] 14.2 Write end-to-end integration tests: create project → generate API key → ingest traces (with events, resources, LLM attrs, prompt linkage, session) → query traces → verify content display → check analytics rollups → verify prompt CRUD and render
- [ ] 14.3 Add health check (`GET /health`) and readiness check (`GET /ready`) with DB connectivity verification
- [ ] 14.4 Write documentation: getting started guide (docker compose up → create project → configure OTel SDK → view traces), span attribute conventions for prompt linkage and sessions

---

## 15. Evaluation Framework API (Phase 2)

- [ ] 15.1 Create PostgreSQL migrations: `evaluations` table, `evaluation_criteria` (scorer type, config JSONB), `evaluation_runs` (status, progress, timestamps), `evaluation_results` (run_id, trace_id/case_id, per-criterion scores JSONB, aggregate score)
- [ ] 15.2 Create PostgreSQL migrations: `datasets` table (id, project_id, name, description, created_at), `dataset_cases` table (id, dataset_id, input TEXT, expected_output TEXT, metadata JSONB, created_at)
- [ ] 15.3 Implement evaluation domain entities: `Evaluation`, `Criterion`, `Scorer` interface, `Run`, `Result`, `Dataset`, `DatasetCase`
- [ ] 15.4 Implement rule-based scorers: regex match, contains/not-contains, JSON schema validation — all implementing `Scorer` interface, operating on `output_content` text
- [ ] 15.5 Implement LLM-as-judge scorer: load provider config from `llm_providers`, send output + rubric to judge LLM, parse numeric score — implementing `Scorer` interface
- [ ] 15.6 Implement `DatasetRepository` PostgreSQL adapter (CRUD datasets, add/remove cases, import from traces)
- [ ] 15.7 Implement `EvaluationRepository` PostgreSQL adapter (CRUD evaluations, create/query runs, store results)
- [ ] 15.8 Implement `EvaluationService`: create evaluation, run against traces (filter → fetch output_content → score), run against dataset (iterate cases → call LLM → score output → compare expected), aggregate results
- [ ] 15.9 Implement REST handlers: `POST/GET /api/v1/evaluations`, `POST /api/v1/evaluations/:id/runs`, `GET /api/v1/evaluations/:id/runs/:runId`
- [ ] 15.10 Implement REST handlers: `POST/GET /api/v1/datasets`, `POST /api/v1/datasets/:id/cases`, `GET /api/v1/datasets/:id`

## 16. Prompt Playground & LLM Provider UI (Phase 2)

- [ ] 16.1 Implement LLM provider configuration UI in project settings: add provider (OpenAI, Anthropic, custom), enter API key (masked), set available models, test connection
- [ ] 16.2 Implement `LLMProviderRepository` PostgreSQL adapter with AES-256-GCM encryption/decryption for API keys
- [ ] 16.3 Implement `LLMProviderService`: CRUD providers, call LLM (chat completion), abstract across OpenAI/Anthropic/custom providers
- [ ] 16.4 Implement REST handlers for LLM providers: `POST/GET /api/v1/projects/:id/llm-providers`, `DELETE /api/v1/projects/:id/llm-providers/:providerId`
- [ ] 16.5 Implement prompt playground UI: select prompt version → fill variables → select provider + model → "Run" button → display response with token count + cost + latency

## 17. Alerting (Phase 2)

- [ ] 17.1 Create PostgreSQL migration: `alert_rules` table (id, project_id, metric, threshold, window_minutes, webhook_url, enabled, last_triggered_at, created_at)
- [ ] 17.2 Implement alert evaluation in analytics rollup goroutine: after each rollup cycle, check alert rules against current values, fire webhook POST for newly triggered alerts, deduplicate (don't re-fire until condition resolves)
- [ ] 17.3 Implement REST handlers: `POST/GET /api/v1/projects/:id/alerts`, `PUT/DELETE /api/v1/projects/:id/alerts/:alertId`
- [ ] 17.4 Implement alerting UI in project settings: create/edit alert rules (metric selector, threshold input, webhook URL), alert history log

## 18. Python + TypeScript SDKs (Phase 2)

- [ ] 18.1 Create `sdks/python/` package: `telepromptr` Python SDK with `render_prompt()` function that calls render API + returns OTel-compatible span attributes dict, plus `TelePromptr` class for auto-instrumenting LLM calls with prompt linkage and session tracking
- [ ] 18.2 Create `sdks/typescript/` package: `@telepromptr/sdk` npm package with equivalent functionality to the Python SDK
- [ ] 18.3 Write SDK documentation with copy-paste integration examples for common LLM frameworks (LangChain, OpenAI SDK, Anthropic SDK)

## 19. Evaluation UI (Phase 2)

- [ ] 19.1 Implement evaluation list page (`/evaluations`): table with name, criteria count, last run date, last score
- [ ] 19.2 Implement evaluation creation form: name, description, criteria builder (add criterion → select scorer type → configure)
- [ ] 19.3 Implement dataset management page (`/evaluations/datasets`): list datasets, create dataset, add cases manually, import cases from traces
- [ ] 19.4 Implement evaluation run trigger: select source (traces with filters OR dataset), select prompt version for dataset runs, run button, progress indicator
- [ ] 19.5 Implement evaluation results page: score summary cards, per-criterion distribution charts, per-item results table with output text and scores
- [ ] 19.6 Implement prompt version comparison view: run same evaluation against two prompt versions, display side-by-side aggregate scores and per-case diffs

---

## 20. Advanced Features (Phase 3)

- [ ] 20.1 Implement OTLP log ingestion: add `LogsService.Export` gRPC + HTTP handlers, store logs in `trace_logs` table with trace_id correlation, display logs alongside trace spans in the UI
- [ ] 20.2 Implement trace replay view: chronological step-through of a trace showing each LLM call's input → output in sequence, with timing and cost per step — like "session replay for LLMs"
- [ ] 20.3 Implement human review workflow: assign traces/cases to review queue, scoring interface with rubric, annotation text, merge human scores with automated scores in evaluation results
- [ ] 20.4 Implement evaluation trend charts: score history across runs, regression detection, prompt version impact analysis with inline diff
- [ ] 20.5 Implement trace retention: configurable TTL per project, background job that drops old partitions, storage usage display in project settings
