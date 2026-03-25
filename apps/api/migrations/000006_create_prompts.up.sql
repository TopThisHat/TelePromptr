-- 000006_create_prompts.up.sql
-- Creates the prompt management tables for TelePromptr's prompt registry.
--
-- prompts:          Top-level prompt templates, unique per (project_id, name).
-- prompt_versions:  Immutable versioned snapshots of a prompt's template text.
-- prompt_tags:      Free-form tags for organizing and filtering prompts.

CREATE TABLE IF NOT EXISTS prompts (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT        DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS prompt_versions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id       UUID        NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    version         INTEGER     NOT NULL,
    template_text   TEXT        NOT NULL,
    variables       JSONB       NOT NULL DEFAULT '[]',
    status          TEXT        NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'active', 'archived')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(prompt_id, version)
);

CREATE TABLE IF NOT EXISTS prompt_tags (
    prompt_id   UUID    NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    tag         TEXT    NOT NULL,
    PRIMARY KEY (prompt_id, tag)
);
