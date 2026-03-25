-- 000006_create_prompts.down.sql
-- Drops the prompt management tables in dependency order.

DROP TABLE IF EXISTS prompt_tags;
DROP TABLE IF EXISTS prompt_versions;
DROP TABLE IF EXISTS prompts;
