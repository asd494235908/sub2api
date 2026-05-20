ALTER TABLE sms_broadcast_campaigns
    ADD COLUMN IF NOT EXISTS template_var_rows JSONB NOT NULL DEFAULT '[]'::jsonb;

