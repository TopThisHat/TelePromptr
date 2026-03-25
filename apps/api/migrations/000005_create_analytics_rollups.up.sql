-- 000005_create_analytics_rollups.up.sql
-- Creates pre-aggregated analytics tables for fast dashboard queries.
--
-- analytics_rollups: Hourly bucketed summaries per project/model/service.
-- rollup_watermark: Singleton row tracking the last processed timestamp
--                   to support incremental rollup computation.

CREATE TABLE IF NOT EXISTS analytics_rollups (
    project_id          UUID            NOT NULL,
    hour_bucket         TIMESTAMPTZ     NOT NULL,
    model               TEXT            NOT NULL,
    service_name        TEXT            NOT NULL,
    span_count          BIGINT          DEFAULT 0,
    error_count         BIGINT          DEFAULT 0,
    total_input_tokens  BIGINT          DEFAULT 0,
    total_output_tokens BIGINT          DEFAULT 0,
    total_cost          NUMERIC(12,6)   DEFAULT 0,
    latency_sum_ms      BIGINT          DEFAULT 0,
    latency_p50_ms      DOUBLE PRECISION,
    latency_p90_ms      DOUBLE PRECISION,
    latency_p95_ms      DOUBLE PRECISION,
    latency_p99_ms      DOUBLE PRECISION,
    updated_at          TIMESTAMPTZ     DEFAULT NOW(),
    PRIMARY KEY (project_id, hour_bucket, model, service_name)
);

CREATE TABLE IF NOT EXISTS rollup_watermark (
    id                  INTEGER     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_processed_at   TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01',
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Seed the singleton watermark row.
INSERT INTO rollup_watermark (last_processed_at)
VALUES ('1970-01-01')
ON CONFLICT (id) DO NOTHING;
