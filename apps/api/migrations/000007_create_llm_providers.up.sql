-- 000007_create_llm_providers.up.sql
-- Creates the llm_providers table for storing project-scoped LLM provider
-- configurations. API keys are stored encrypted (BYTEA) and decrypted at
-- runtime by the application layer.

CREATE TABLE IF NOT EXISTS llm_providers (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider            TEXT        NOT NULL,
    display_name        TEXT,
    api_key_encrypted   BYTEA       NOT NULL,
    base_url            TEXT,
    models              JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, provider)
);
