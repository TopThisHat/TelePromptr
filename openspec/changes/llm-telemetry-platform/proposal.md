## Why

Teams building LLM-powered applications need to trace requests through complex chains, manage and version prompts, evaluate output quality, and monitor costs. Tools like Langfuse, Arize Phoenix, and Helicone address parts of this — but they are either cloud-only, lack prompt-to-trace linkage, or treat evaluation as an afterthought. TelePromptr is a self-hosted LLM telemetry platform that unifies tracing, prompt management, and evaluation with first-class data linkage between all three concerns.

The key differentiator: prompts, traces, and evaluations are connected by design. When an LLM call is traced, TelePromptr knows which prompt version produced it, stores the actual input/output content, and can evaluate the output systematically — closing the feedback loop that other tools leave open. This linkage ships in the MVP, not as an afterthought.

## What Changes

- Establish a monorepo with a Go 1.25 backend and Svelte 5 + Tailwind CSS 4 frontend, using a Justfile for task orchestration
- Build an OTLP-compatible trace ingestion pipeline (gRPC + HTTP) with full span event and resource attribute support, LLM content extraction (input/output text), session grouping, and ingestion backpressure
- Create a prompt management system with versioning, tagging, deployment lifecycle, and trace linkage — shipped in MVP for day-one differentiation
- Provide a real-time analytics dashboard with incremental rollup aggregation and webhook-based cost/error alerting
- Build a project model with admin-mode UI access and SHA-256 API key authentication for ingestion
- Implement an evaluation framework with curated datasets, automated scoring, human review, and prompt version comparison
- Use PostgreSQL as the sole data store (with table partitioning and tsvector search), keeping the option to add a columnar store later behind repository interfaces

**Phased delivery:**
- **Phase 1 (MVP):** Trace ingestion (with events, resources, sessions, backpressure), trace exploration (with content display, service filtering, session grouping), analytics dashboard (with incremental rollups), prompt management (versioning, render with linkage, tags), project management, Docker Compose dev environment
- **Phase 2:** Evaluation framework (with datasets), prompt playground, cost/error alerting, LLM provider configuration, Python + TypeScript SDKs
- **Phase 3:** Log ingestion, trace replay, advanced evaluation (human review workflows, trend analysis), dataset management UI

## Capabilities

### New Capabilities

- `trace-ingestion`: Receives and stores OpenTelemetry traces via OTLP gRPC and HTTP with full span support — attributes, events (including LLM input/output content), resource attributes (service name, deployment), session grouping, and backpressure controls
- `trace-exploration`: Interactive UI for searching, filtering, and visualizing trace data with span-level detail, LLM content display (prompt/completion text), service filtering, session grouping, and waterfall visualization
- `prompt-management`: Create, version, tag, and manage prompt templates with variable interpolation, deployment lifecycle, trace linkage via render API, and prompt playground for testing
- `llm-evaluation`: Define evaluation criteria, run automated scoring against live traces or curated datasets, conduct human reviews, compare prompt versions, and track score trends over time
- `analytics-dashboard`: Real-time metrics for cost tracking, latency distribution, error rates, token usage, and model performance comparison — with configurable webhook alerts for threshold breaches
- `project-management`: Admin-mode UI access, project isolation, API key management, LLM provider configuration, and cost-per-token rate settings

### Modified Capabilities

(none — greenfield project)

## Impact

- **New codebase**: Monorepo with `apps/api` (Go), `apps/web` (Svelte), Justfile for orchestration
- **Infrastructure**: Requires PostgreSQL (single database for all data, partitioned trace tables)
- **Dependencies**: OpenTelemetry protobuf definitions (Go), Svelte 5, Tailwind CSS 4
- **APIs**: REST API with dual auth (admin mode for UI, API key for programmatic access); OTLP-compatible gRPC/HTTP ingestion
- **External integrations**: Receives standard OpenTelemetry traces; sends webhook alerts; calls LLM provider APIs for evaluation and playground
- **Client SDKs**: Python and TypeScript SDKs for prompt render + auto span attribute injection (Phase 2)
- **Local dev**: Docker Compose for PostgreSQL + Go API + Svelte dev server
