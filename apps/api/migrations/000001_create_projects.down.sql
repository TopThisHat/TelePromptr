-- 000001_create_projects.down.sql
-- Drops the projects table and its indexes.
-- WARNING: This will cascade-delete all dependent rows (api_keys, spans, etc.)
-- if those tables have ON DELETE CASCADE foreign keys.

DROP INDEX IF EXISTS idx_projects_name;
DROP TABLE IF EXISTS projects CASCADE;
