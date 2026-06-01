CREATE TABLE IF NOT EXISTS user_affiliate_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_review',
    payout_method VARCHAR(32) NOT NULL,
    payout_account_note TEXT NOT NULL DEFAULT '',
    admin_note TEXT NOT NULL DEFAULT '',
    payout_channel VARCHAR(64) NOT NULL DEFAULT '',
    payout_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    reject_reason TEXT NOT NULL DEFAULT '',
    failure_reason TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ NULL,
    paid_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_affiliate_withdrawals_amount_positive CHECK (amount > 0),
    CONSTRAINT user_affiliate_withdrawals_status_check CHECK (
        status IN ('pending_review', 'approved', 'paid', 'rejected', 'failed', 'cancelled')
    )
);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_user_id
    ON user_affiliate_withdrawals(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_status
    ON user_affiliate_withdrawals(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_withdrawals_created_at
    ON user_affiliate_withdrawals(created_at DESC);

COMMENT ON TABLE user_affiliate_withdrawals IS '邀请返利现金提现申请';
COMMENT ON COLUMN user_affiliate_withdrawals.amount IS '提现金额，从 user_affiliates.aff_quota 扣减，不复用 aff_frozen_quota';
COMMENT ON COLUMN user_affiliate_withdrawals.status IS 'pending_review|approved|paid|rejected|failed|cancelled';
COMMENT ON COLUMN user_affiliate_withdrawals.payout_method IS '用户选择或填写的收款方式，如 wechat_manual';
COMMENT ON COLUMN user_affiliate_withdrawals.payout_account_note IS '用户填写的收款说明，不保存证件等高敏信息';

INSERT INTO settings (key, value, updated_at)
VALUES
  ('affiliate_withdraw_enabled', 'false', NOW()),
  ('affiliate_withdraw_min_amount', '1.00000000', NOW()),
  ('affiliate_withdraw_max_amount', '0.00000000', NOW()),
  ('affiliate_withdraw_daily_request_limit', '3', NOW()),
  ('affiliate_withdraw_help_text', '', NOW())
ON CONFLICT (key) DO NOTHING;
