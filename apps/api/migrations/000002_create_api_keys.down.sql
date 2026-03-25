-- 000002_create_api_keys.down.sql
-- Drops the api_keys table and its indexes.

DROP INDEX IF EXISTS idx_api_keys_project;
DROP INDEX IF EXISTS idx_api_keys_hash;
DROP TABLE IF EXISTS api_keys;
