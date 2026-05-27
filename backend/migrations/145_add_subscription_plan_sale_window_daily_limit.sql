ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS sale_starts_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sale_ends_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS daily_purchase_limit INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_subscription_plans_sale_starts_at ON subscription_plans(sale_starts_at);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sale_ends_at ON subscription_plans(sale_ends_at);
