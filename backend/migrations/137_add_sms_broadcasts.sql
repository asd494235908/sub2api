CREATE TABLE IF NOT EXISTS sms_broadcast_campaigns (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    mode VARCHAR(20) NOT NULL DEFAULT 'freeform',
    body TEXT NOT NULL,
    rendered_body TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    audience JSONB NOT NULL DEFAULT '{}'::jsonb,
    template_vars JSONB NOT NULL DEFAULT '{}'::jsonb,
    total_recipients BIGINT NOT NULL DEFAULT 0,
    sent_count BIGINT NOT NULL DEFAULT 0,
    failed_count BIGINT NOT NULL DEFAULT 0,
    skipped_count BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sms_broadcast_recipients (
    id BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES sms_broadcast_campaigns(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number VARCHAR(32) NOT NULL,
    raw_phone VARCHAR(64) NOT NULL DEFAULT '',
    rendered_body TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(campaign_id, phone_number)
);

CREATE INDEX IF NOT EXISTS idx_sms_broadcast_campaigns_status_created_at
    ON sms_broadcast_campaigns(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sms_broadcast_campaigns_created_at
    ON sms_broadcast_campaigns(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sms_broadcast_recipients_campaign_id
    ON sms_broadcast_recipients(campaign_id);

CREATE INDEX IF NOT EXISTS idx_sms_broadcast_recipients_status
    ON sms_broadcast_recipients(status);

COMMENT ON TABLE sms_broadcast_campaigns IS '管理员群发短信活动';
COMMENT ON TABLE sms_broadcast_recipients IS '群发短信收件人快照与发送明细';
