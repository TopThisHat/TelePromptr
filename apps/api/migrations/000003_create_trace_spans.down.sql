-- 000003_create_trace_spans.down.sql
-- Drops the trace_spans partitioned table and all its partitions.
-- Dropping the parent table cascades to all child partitions.

DROP TABLE IF EXISTS trace_spans CASCADE;
