-- 000003_create_trace_spans.up.sql
-- Creates the trace_spans table, the core telemetry storage for TelePromptr.
-- This is a range-partitioned table on start_time (monthly partitions).
--
-- Columns are organized into logical groups:
--   - Core OTel fields (trace_id, span_id, etc.)
--   - Flexible attribute storage (JSONB)
--   - Resource identification (service_name, version, env)
--   - LLM-specific fields (model, tokens, cost)
--   - Content capture (input/output text)
--   - TelePromptr metadata (prompt_id, session_id)
--   - Full-text search (generated tsvector column)

CREATE TABLE IF NOT EXISTS trace_spans (
    -- Core OpenTelemetry fields
    trace_id            TEXT        NOT NULL,
    span_id             TEXT        NOT NULL,
    parent_span_id      TEXT,
    name                TEXT        NOT NULL,
    span_kind           TEXT        NOT NULL,
    status_code         TEXT        NOT NULL DEFAULT 'UNSET',
    status_message      TEXT,
    start_time          TIMESTAMPTZ NOT NULL,
    end_time            TIMESTAMPTZ,
    duration_ms         DOUBLE PRECISION,

    -- Flexible attribute storage
    attributes          JSONB       NOT NULL DEFAULT '{}',
    events              JSONB       NOT NULL DEFAULT '[]',
    resource_attributes JSONB       NOT NULL DEFAULT '{}',

    -- Resource identification
    service_name        TEXT        NOT NULL DEFAULT '',
    service_version     TEXT        NOT NULL DEFAULT '',
    deployment_environment TEXT     NOT NULL DEFAULT '',

    -- LLM-specific fields
    model               TEXT,
    provider            TEXT,
    input_tokens        BIGINT,
    output_tokens       BIGINT,
    total_tokens        BIGINT,
    cost                NUMERIC(12,6),

    -- Content capture
    input_content       TEXT,
    output_content      TEXT,

    -- TelePromptr metadata
    prompt_id           UUID,
    prompt_version      INTEGER,
    prompt_execution_id UUID,
    session_id          UUID,

    -- Meta
    project_id          UUID        NOT NULL,

    -- Full-text search: generated column combining searchable text fields
    search_text         TSVECTOR    GENERATED ALWAYS AS (
        to_tsvector('english',
            coalesce(name, '') || ' ' ||
            coalesce(service_name, '') || ' ' ||
            coalesce(model, '') || ' ' ||
            coalesce(status_message, '') || ' ' ||
            coalesce(input_content, '') || ' ' ||
            coalesce(output_content, '')
        )
    ) STORED,

    -- Composite primary key includes partition key (start_time)
    PRIMARY KEY (project_id, span_id, start_time)
) PARTITION BY RANGE (start_time);

-- Create initial monthly partitions: current month + next 2 months.
-- Partition naming: trace_spans_YYYYMM
-- Using date_trunc ensures boundaries align to month starts.

DO $$
DECLARE
    partition_start DATE;
    partition_end   DATE;
    partition_name  TEXT;
BEGIN
    FOR i IN 0..2 LOOP
        partition_start := date_trunc('month', CURRENT_DATE) + (i || ' months')::INTERVAL;
        partition_end   := partition_start + INTERVAL '1 month';
        partition_name  := 'trace_spans_' || to_char(partition_start, 'YYYYMM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF trace_spans
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            partition_start,
            partition_end
        );
    END LOOP;
END $$;
