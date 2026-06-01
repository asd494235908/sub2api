CREATE TABLE IF NOT EXISTS user_affiliate_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    identity_type VARCHAR(32) NOT NULL,
    rate_multiplier DECIMAL(10,4) NOT NULL,
    source_inviter_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    qualification_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_identities_user_type
ON user_affiliate_identities(user_id, identity_type);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_identities_active
ON user_affiliate_identities(user_id, status, expires_at);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_identities_source
ON user_affiliate_identities(source_inviter_id, identity_type, status);

CREATE TABLE IF NOT EXISTS user_signup_fingerprints (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    composite_hash VARCHAR(128) NOT NULL DEFAULT '',
    canvas_hash VARCHAR(128) NOT NULL DEFAULT '',
    webgl_hash VARCHAR(128) NOT NULL DEFAULT '',
    components JSONB NOT NULL DEFAULT '{}'::jsonb,
    duplicate_count INTEGER NOT NULL DEFAULT 0,
    risk_flagged BOOLEAN NOT NULL DEFAULT false,
    risk_reason VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_signup_fingerprints_composite
ON user_signup_fingerprints(composite_hash)
WHERE composite_hash <> '';

CREATE INDEX IF NOT EXISTS idx_user_signup_fingerprints_risk
ON user_signup_fingerprints(risk_flagged);

INSERT INTO settings (key, value, updated_at)
VALUES
  ('affiliate_identity_enabled', 'false', NOW()),
  ('affiliate_identity_config', '{"inviter_rate_multiplier":1.5,"invitee_rate_multiplier":1.4,"duration_hours":720,"qualified_invitee_count":0,"qualified_pay_amount":50,"eligible_order_types":["balance","subscription"],"fingerprint_enforcement_enabled":true,"max_accounts_per_fingerprint_hash":3}', NOW())
ON CONFLICT (key) DO NOTHING;
