-- 000005_create_analytics_rollups.down.sql
-- Drops the analytics rollup tables.

DROP TABLE IF EXISTS rollup_watermark;
DROP TABLE IF EXISTS analytics_rollups;
