ALTER TABLE sms_broadcast_campaigns
    ADD COLUMN IF NOT EXISTS template_id VARCHAR(64) NOT NULL DEFAULT '';
