## ADDED Requirements

### Requirement: Cost analytics
The system SHALL display cost analytics for LLM usage within a project via both UI and API, powered by the incremental rollup table.

#### Scenario: Get cost analytics via API
- **WHEN** a client sends `GET /api/v1/analytics/cost?start=...&end=...&granularity=hour`
- **THEN** the system queries the `analytics_rollups` table and returns time-bucketed cost data with total cost per bucket and optional model breakdown

#### Scenario: View total cost over time in UI
- **WHEN** a user navigates to the cost analytics view
- **THEN** the system displays a time-series chart of total LLM cost by time bucket (hour, day, week)

#### Scenario: Cost breakdown by model
- **WHEN** a user enables model breakdown
- **THEN** the system displays cost per model as a stacked chart and a table showing each model's total cost, percentage, and call count

#### Scenario: Cost breakdown by service
- **WHEN** a user enables service breakdown
- **THEN** the system displays cost per service_name, showing which applications are consuming the most LLM spend

### Requirement: Latency analytics
The system SHALL display latency distribution and percentile metrics for LLM calls.

#### Scenario: Get latency analytics via API
- **WHEN** a client sends `GET /api/v1/analytics/latency?start=...&end=...`
- **THEN** the system returns latency percentiles (p50, p90, p95, p99) from the rollup table for fast dashboard rendering

#### Scenario: Get exact latency percentiles via API
- **WHEN** a client sends `GET /api/v1/analytics/latency?start=...&end=...&exact=true`
- **THEN** the system queries `trace_spans` directly for exact percentile calculation (slower but precise)

#### Scenario: View latency percentiles in UI
- **WHEN** a user views the latency analytics panel
- **THEN** the system displays p50, p90, p95, p99 values and a histogram of LLM call latencies

#### Scenario: Latency by model comparison
- **WHEN** a user enables model comparison
- **THEN** the system displays latency percentiles per model in a comparison table

### Requirement: Token usage analytics
The system SHALL display token usage metrics for LLM calls.

#### Scenario: Get token analytics via API
- **WHEN** a client sends `GET /api/v1/analytics/tokens?start=...&end=...`
- **THEN** the system returns input, output, and total tokens aggregated by time bucket

#### Scenario: View token usage in UI
- **WHEN** a user views the token analytics panel
- **THEN** the system displays time-series charts for input, output, and total tokens

#### Scenario: Token usage by model
- **WHEN** a user enables model breakdown
- **THEN** the system displays per-model token usage in a sortable table

### Requirement: Error rate monitoring
The system SHALL display error rates and error details for LLM calls.

#### Scenario: Get error analytics via API
- **WHEN** a client sends `GET /api/v1/analytics/errors?start=...&end=...`
- **THEN** the system returns error rate (percentage of error spans) and error counts by time bucket

#### Scenario: View error rate in UI
- **WHEN** a user views the error rate panel
- **THEN** the system displays a time-series chart of error rate and total error count

#### Scenario: View error details
- **WHEN** a user clicks on an error data point
- **THEN** the system displays recent error spans with status_message, model, service_name, and a link to the full trace

### Requirement: Dashboard time range controls
The system SHALL provide unified time range controls that apply to all analytics panels.

#### Scenario: Select preset time range
- **WHEN** a user selects a preset (last 1h, 6h, 24h, 7d, 30d)
- **THEN** all analytics panels update to reflect the selected range

#### Scenario: Select custom time range
- **WHEN** a user specifies a custom start and end datetime
- **THEN** all analytics panels update to display data within the custom range

### Requirement: Webhook alerting (Phase 2)
The system SHALL support configurable webhook alerts when analytics metrics cross thresholds.

#### Scenario: Configure a cost alert
- **WHEN** an admin creates an alert rule with `{"metric": "cost_per_day", "threshold": 100.00, "webhook_url": "https://..."}`
- **THEN** the system stores the alert rule for the project

#### Scenario: Trigger a cost alert
- **WHEN** the analytics rollup job detects that a project's daily cost has exceeded its configured threshold
- **THEN** the system sends a POST request to the configured webhook URL with alert details (project, metric, current value, threshold)

#### Scenario: Configure an error rate alert
- **WHEN** an admin creates an alert rule with `{"metric": "error_rate", "threshold": 0.05, "window_minutes": 60, "webhook_url": "https://..."}`
- **THEN** the system monitors the error rate over the specified window and triggers the webhook when exceeded

#### Scenario: Alert deduplication
- **WHEN** an alert condition remains true across multiple rollup cycles
- **THEN** the system sends the alert webhook only once until the condition resolves, then re-arms
