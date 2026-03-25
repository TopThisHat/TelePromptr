-- 000004_create_trace_spans_indexes.down.sql
-- Drops all trace_spans indexes added in the up migration.
-- The primary key index is not dropped here (managed by the table migration).

DROP INDEX IF EXISTS idx_spans_time_brin;
DROP INDEX IF EXISTS idx_spans_search;
DROP INDEX IF EXISTS idx_spans_prompt;
DROP INDEX IF EXISTS idx_spans_session;
DROP INDEX IF EXISTS idx_spans_model;
DROP INDEX IF EXISTS idx_spans_service;
DROP INDEX IF EXISTS idx_spans_trace;
DROP INDEX IF EXISTS idx_spans_project_time;
