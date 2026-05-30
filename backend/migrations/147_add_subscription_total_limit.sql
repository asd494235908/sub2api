ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS subscription_total_limit_usd DECIMAL(20,8);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS subscription_total_limit_usd DECIMAL(20,8);

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS total_limit_usd DECIMAL(20,8),
    ADD COLUMN IF NOT EXISTS total_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0;
