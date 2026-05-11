CREATE SCHEMA IF NOT EXISTS prompt_archive;

CREATE TABLE IF NOT EXISTS prompt_archive.archive_settings (
    id BIGSERIAL PRIMARY KEY,
    singleton_key SMALLINT NOT NULL UNIQUE DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    all_groups BOOLEAN NOT NULL DEFAULT FALSE,
    group_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    bucket TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id BIGINT NOT NULL DEFAULT 0
);

INSERT INTO prompt_archive.archive_settings (singleton_key, enabled, all_groups, group_ids, bucket, updated_at, updated_by_user_id)
VALUES (1, FALSE, FALSE, '[]'::jsonb, '', NOW(), 0)
ON CONFLICT (singleton_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS prompt_archive.ai_design_records (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    client_request_id TEXT NOT NULL DEFAULT '',
    session_id TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    username_snapshot TEXT NOT NULL DEFAULT '',
    email_snapshot TEXT NOT NULL DEFAULT '',
    api_key_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    protocol TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT '',
    system_prompt TEXT NOT NULL DEFAULT '',
    user_prompt_text TEXT NOT NULL DEFAULT '',
    prompt_summary TEXT NOT NULL DEFAULT '',
    object_key TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    error_message TEXT NOT NULL DEFAULT '',
    attachments_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prompt_archive_records_created_at
    ON prompt_archive.ai_design_records (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prompt_archive_records_session_group
    ON prompt_archive.ai_design_records (session_id, group_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prompt_archive_records_user_group
    ON prompt_archive.ai_design_records (user_id, group_id, created_at DESC);
