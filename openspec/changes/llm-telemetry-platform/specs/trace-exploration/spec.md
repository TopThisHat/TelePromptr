## ADDED Requirements

### Requirement: Trace list with search and filtering
The system SHALL provide a paginated list of traces with search and multi-facet filtering via both the UI and REST API.

#### Scenario: List traces via API
- **WHEN** a client sends `GET /api/v1/traces?page=1&per_page=50`
- **THEN** the system returns a paginated response with trace summaries (trace ID, root span name, duration, timestamp, status, model, service name) and standard pagination metadata

#### Scenario: Filter traces by time range
- **WHEN** a client sends `GET /api/v1/traces?start=...&end=...`
- **THEN** the system returns only traces with root spans within the specified time range, leveraging partition pruning for performance

#### Scenario: Filter traces by service
- **WHEN** a client sends `GET /api/v1/traces?service_name=my-chatbot`
- **THEN** the system returns only traces from the specified service

#### Scenario: Filter traces by model and status
- **WHEN** a client applies filters for model name, status (error/ok), minimum duration, or span kind
- **THEN** the system returns only traces matching all applied filter criteria

#### Scenario: Filter traces by session
- **WHEN** a client sends `GET /api/v1/traces?session_id=abc123`
- **THEN** the system returns all traces in the specified session, ordered by start time

#### Scenario: Filter traces by prompt version
- **WHEN** a client sends `GET /api/v1/traces?prompt_id=xyz&prompt_version=3`
- **THEN** the system returns traces containing spans linked to the specified prompt version

#### Scenario: Full-text search
- **WHEN** a client sends `GET /api/v1/traces?q=search_term`
- **THEN** the system uses the tsvector index to return traces with spans whose name, service, model, or status message match the query

#### Scenario: List traces in the UI
- **WHEN** a user navigates to the traces page
- **THEN** the system displays a paginated table with trace ID, root span name, service, duration, timestamp, status, and model columns, with filter controls above

### Requirement: Trace detail view with span waterfall
The system SHALL display a detailed view of a single trace with a waterfall visualization showing full span content.

#### Scenario: Get trace detail via API
- **WHEN** a client sends `GET /api/v1/traces/:traceId`
- **THEN** the system returns the full trace including all spans with attributes, events, timing, parent-child relationships, resource attributes, and LLM content (input/output text)

#### Scenario: View trace waterfall in UI
- **WHEN** a user selects a trace from the list
- **THEN** the system displays a waterfall diagram showing all spans with relative timing, duration, parent-child nesting, and span kind indicators

#### Scenario: View span details with content
- **WHEN** a user clicks on a span in the waterfall
- **THEN** the system displays a detail panel with: attributes, events, status (code + message), timing, and resource info (service name, version, environment)

#### Scenario: View LLM span content
- **WHEN** a user views a span that contains LLM attributes
- **THEN** the system displays a dedicated LLM section showing: model, provider, input/output tokens, cost, latency, the actual prompt text (input_content), the actual completion text (output_content), and prompt linkage info (if present)

### Requirement: Session view
The system SHALL provide a session-oriented view that groups traces by conversation.

#### Scenario: View session timeline
- **WHEN** a user views a trace that has a session_id and clicks "View Session"
- **THEN** the system displays all traces in the same session, ordered chronologically, showing the multi-turn conversation flow

#### Scenario: Session list
- **WHEN** a user navigates to the sessions view or filters by session
- **THEN** the system displays distinct sessions with their trace count, total duration, and time range

### Requirement: Trace comparison
The system SHALL allow users to compare two traces side by side.

#### Scenario: Compare two traces
- **WHEN** a user selects two traces and chooses "Compare"
- **THEN** the system displays both waterfalls side by side with differences in duration, span count, and attributes highlighted

### Requirement: Trace export
The system SHALL allow exporting trace data.

#### Scenario: Export traces via UI
- **WHEN** a user selects one or more traces and clicks "Export"
- **THEN** the system downloads the selected traces as a JSON file

#### Scenario: Export traces via API
- **WHEN** a client sends `GET /api/v1/traces/:traceId?format=otlp`
- **THEN** the system returns the trace in OTLP-compatible JSON format
