-- Record the platform balance amount credited by affiliate cash transfers.
-- The existing amount column remains the cash rebate amount for rebate/withdrawal accounting.

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS platform_amount DECIMAL(20,8) NULL;

COMMENT ON COLUMN user_affiliate_ledger.platform_amount IS '返利转平台余额实际到账的平台额度；仅 action=transfer 使用，历史无法可靠回填时为 NULL';
