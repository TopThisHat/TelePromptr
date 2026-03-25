## ADDED Requirements

### Requirement: OTLP gRPC trace ingestion
The system SHALL accept OpenTelemetry trace data via OTLP gRPC protocol on a configurable port (default 4317), implemented directly from the OpenTelemetry protobuf `TraceService.Export` definition.

#### Scenario: Receive traces via gRPC
- **WHEN** an instrumented application sends an `ExportTraceServiceRequest` to the gRPC endpoint with a valid API key in the `authorization` metadata field
- **THEN** the system accepts the spans, extracts resource attributes, span events, LLM-specific attributes, and input/output content, associates them with the API key's project, and persists them to PostgreSQL

#### Scenario: Reject malformed gRPC payloads
- **WHEN** a client sends a malformed or invalid protobuf payload to the gRPC endpoint
- **THEN** the system returns gRPC status `INVALID_ARGUMENT` and does not persist any data

#### Scenario: Reject unauthenticated gRPC requests
- **WHEN** a client sends traces without an API key in the gRPC metadata
- **THEN** the system returns gRPC status `UNAUTHENTICATED`

### Requirement: OTLP HTTP trace ingestion
The system SHALL accept OpenTelemetry trace data via OTLP HTTP protocol on a configurable port (default 4318) at `POST /v1/traces`.

#### Scenario: Receive traces via HTTP with JSON
- **WHEN** a client POSTs OTLP trace data with `Content-Type: application/json` and a valid `Authorization: Bearer <key>` header
- **THEN** the system deserializes the JSON payload and processes spans identically to gRPC-received spans

#### Scenario: Receive traces via HTTP with protobuf
- **WHEN** a client POSTs OTLP trace data with `Content-Type: application/x-protobuf` and a valid API key
- **THEN** the system deserializes the protobuf payload and processes spans identically

#### Scenario: Reject unauthenticated HTTP requests
- **WHEN** a client POSTs traces without an `Authorization` header or with an invalid key
- **THEN** the system returns HTTP 401 Unauthorized

### Requirement: API key verification via SHA-256
The system SHALL verify API keys using SHA-256 hashing with indexed lookup.

#### Scenario: Verify API key in O(1)
- **WHEN** a trace request includes an API key
- **THEN** the system computes `sha256(key)`, performs an indexed lookup in the `api_keys` table, and resolves the associated project in constant time

#### Scenario: Reject revoked API keys
- **WHEN** a trace request includes an API key that has been revoked
- **THEN** the system rejects the request with an authentication error

### Requirement: Resource attribute extraction
The system SHALL extract and denormalize OpenTelemetry resource attributes from the `ExportTraceServiceRequest` onto each span record.

#### Scenario: Extract service identity
- **WHEN** a trace request includes resource attributes `service.name`, `service.version`, and `deployment.environment`
- **THEN** the system stores these as indexed columns (`service_name`, `service_version`, `deployment_environment`) on every span in the batch

#### Scenario: Store full resource attributes
- **WHEN** a trace request includes arbitrary resource attributes
- **THEN** the system stores the complete resource attribute set as a JSONB column on each span

### Requirement: Span event ingestion
The system SHALL ingest and store all span events, including timestamps, event names, and event attributes.

#### Scenario: Store span events
- **WHEN** a span includes events (e.g., exception events, LLM content events)
- **THEN** the system stores the full event array as a JSONB column on the span record

#### Scenario: Spans without events
- **WHEN** a span has no events
- **THEN** the system stores the span with a null or empty events field

### Requirement: LLM content extraction
The system SHALL extract the actual input and output text from LLM span events and store them as dedicated text columns.

#### Scenario: Extract prompt content from span events
- **WHEN** a span contains an event named `gen_ai.content.prompt` (or known equivalent conventions)
- **THEN** the system extracts the text content and stores it in the `input_content` TEXT column

#### Scenario: Extract completion content from span events
- **WHEN** a span contains an event named `gen_ai.content.completion` (or known equivalent conventions)
- **THEN** the system extracts the text content and stores it in the `output_content` TEXT column

#### Scenario: Fallback content extraction from attributes
- **WHEN** span events do not contain content but span attributes contain `gen_ai.prompt` or `gen_ai.completion`
- **THEN** the system falls back to extracting content from attributes

#### Scenario: No content available
- **WHEN** neither span events nor attributes contain LLM content
- **THEN** the system stores the span with null `input_content` and `output_content`

### Requirement: LLM metadata extraction
The system SHALL extract and index LLM-specific metadata from span attributes following OpenTelemetry GenAI semantic conventions.

#### Scenario: Extract model and provider
- **WHEN** a span contains `gen_ai.system`, `gen_ai.request.model`, or `gen_ai.response.model`
- **THEN** the system stores these as indexed columns (`model`, `provider`)

#### Scenario: Extract token usage
- **WHEN** a span contains `gen_ai.usage.input_tokens` and `gen_ai.usage.output_tokens`
- **THEN** the system stores token counts and computes `total_tokens` as the sum

#### Scenario: Compute cost from project settings
- **WHEN** a span has token usage and the project has cost-per-token rates configured for the model
- **THEN** the system computes and stores the estimated cost

### Requirement: Span kind and status storage
The system SHALL store the full OTel span kind and status information.

#### Scenario: Store span kind
- **WHEN** a span has a `SpanKind` value (CLIENT, SERVER, INTERNAL, PRODUCER, CONSUMER)
- **THEN** the system stores it as a `span_kind` column on the span record

#### Scenario: Store status with message
- **WHEN** a span has a status code (OK or ERROR) and an optional status message
- **THEN** the system stores `status_code` (integer) and `status_message` (text) as separate columns

### Requirement: Prompt linkage attribute extraction
The system SHALL extract TelePromptr prompt linkage attributes from spans.

#### Scenario: Extract prompt tracking attributes
- **WHEN** a span contains `telepromptr.prompt.id`, `telepromptr.prompt.version`, and/or `telepromptr.prompt.execution_id`
- **THEN** the system stores these as indexed columns

#### Scenario: Spans without prompt attributes
- **WHEN** a span does not contain any `telepromptr.prompt.*` attributes
- **THEN** the system stores the span with null prompt linkage fields

### Requirement: Session attribute extraction
The system SHALL extract session identifiers from spans for conversation grouping.

#### Scenario: Extract session ID
- **WHEN** a span contains a `telepromptr.session.id` attribute
- **THEN** the system stores it as an indexed `session_id` column

#### Scenario: Spans without session ID
- **WHEN** a span does not contain a `telepromptr.session.id` attribute
- **THEN** the system stores the span with a null `session_id`

### Requirement: Trace correlation and storage
The system SHALL maintain the full trace structure and store spans in a date-partitioned PostgreSQL table.

#### Scenario: Reconstruct trace tree
- **WHEN** multiple spans arrive with the same trace ID but different span/parent span IDs
- **THEN** the system stores them for full parent-child hierarchy reconstruction

#### Scenario: Out-of-order span arrival
- **WHEN** child spans arrive before their parent span
- **THEN** the system stores all spans and correctly reconstructs the tree once all arrive

### Requirement: Batch insert buffering with backpressure
The system SHALL buffer incoming spans in a bounded in-memory buffer and flush to PostgreSQL in batches, with backpressure when the buffer is full.

#### Scenario: Normal throughput batch insert
- **WHEN** spans arrive at normal throughput
- **THEN** the system buffers them and flushes to PostgreSQL when batch size (default 1000) or flush interval (default 5s) is reached

#### Scenario: Buffer full — backpressure
- **WHEN** the in-memory buffer reaches its configured maximum size (default 100,000 spans)
- **THEN** the system returns gRPC `RESOURCE_EXHAUSTED` or HTTP 429 with `Retry-After` header, and does not accept more spans until buffer space is available

#### Scenario: Per-project rate limiting
- **WHEN** a project has a configured rate limit (spans/second) and the incoming rate exceeds it
- **THEN** the system returns a throttling response for that project's requests while continuing to accept spans from other projects

### Requirement: Full-text search indexing
The system SHALL maintain a tsvector search index on span records for efficient full-text search.

#### Scenario: Index span for search
- **WHEN** a span is persisted
- **THEN** the system populates a `search_text` tsvector column by concatenating span name, service_name, model, and status_message, and indexes it with a GIN index

### Requirement: Project-scoped ingestion
The system SHALL associate all ingested traces with the project identified by the API key.

#### Scenario: Route traces to correct project
- **WHEN** traces are sent with a valid API key
- **THEN** all spans are stored with the `project_id` associated with that key

### Requirement: Analytics rollup update
The system SHALL update the incremental analytics rollup table after each batch flush.

#### Scenario: Rollup on flush
- **WHEN** a batch of spans is flushed to PostgreSQL
- **THEN** the system updates the rollup watermark so the background rollup goroutine processes the new spans in its next cycle
