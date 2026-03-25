-- 000002_create_api_keys.up.sql
-- Creates the api_keys table for project-scoped authentication.
-- Keys are stored as SHA-256 hashes (key_hash) with a human-readable prefix
-- (key_prefix, e.g. "tp_abc12...") for identification in the UI.

CREATE TABLE IF NOT EXISTS api_keys (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key_hash    BYTEA       NOT NULL,
    key_prefix  TEXT        NOT NULL,
    name        TEXT        NOT NULL DEFAULT '',
    last_used_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at  TIMESTAMPTZ
);

-- Unique hash ensures no duplicate keys exist in the system.
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);

-- Fast lookup of all keys belonging to a project.
CREATE INDEX IF NOT EXISTS idx_api_keys_project ON api_keys(project_id);
