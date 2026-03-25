-- 000004_create_trace_spans_indexes.up.sql
-- Adds performance indexes to the trace_spans partitioned table.
-- These indexes are created on the parent table and automatically
-- propagated to all existing and future partitions.

-- Composite indexes for common query patterns

-- Primary query: list spans for a project ordered by time (dashboard view)
CREATE INDEX IF NOT EXISTS idx_spans_project_time
    ON trace_spans(project_id, start_time DESC);

-- Trace assembly: find all spans belonging to a trace
CREATE INDEX IF NOT EXISTS idx_spans_trace
    ON trace_spans(trace_id);

-- Service filtering: spans for a specific service within a project
CREATE INDEX IF NOT EXISTS idx_spans_service
    ON trace_spans(project_id, service_name, start_time DESC);

-- Model filtering: spans for a specific LLM model within a project
CREATE INDEX IF NOT EXISTS idx_spans_model
    ON trace_spans(project_id, model, start_time DESC);

-- Session lookup: find spans belonging to a user session (partial index)
CREATE INDEX IF NOT EXISTS idx_spans_session
    ON trace_spans(session_id)
    WHERE session_id IS NOT NULL;

-- Prompt lookup: find spans linked to a prompt template (partial index)
CREATE INDEX IF NOT EXISTS idx_spans_prompt
    ON trace_spans(prompt_id)
    WHERE prompt_id IS NOT NULL;

-- Full-text search using GIN on the generated tsvector column
CREATE INDEX IF NOT EXISTS idx_spans_search
    ON trace_spans USING GIN(search_text);

-- BRIN index for efficient time-range scans on partitioned data.
-- BRIN is ideal here because start_time is naturally ordered within partitions.
CREATE INDEX IF NOT EXISTS idx_spans_time_brin
    ON trace_spans USING BRIN(start_time);
