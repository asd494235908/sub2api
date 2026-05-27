ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS daily_sale_starts_at VARCHAR(5),
    ADD COLUMN IF NOT EXISTS daily_sale_ends_at VARCHAR(5);

ALTER TABLE subscription_plans
    ADD CONSTRAINT chk_subscription_plans_daily_sale_starts_at_format
        CHECK (daily_sale_starts_at IS NULL OR daily_sale_starts_at ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    ADD CONSTRAINT chk_subscription_plans_daily_sale_ends_at_format
        CHECK (daily_sale_ends_at IS NULL OR daily_sale_ends_at ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    ADD CONSTRAINT chk_subscription_plans_daily_sale_window_pair
        CHECK ((daily_sale_starts_at IS NULL AND daily_sale_ends_at IS NULL) OR (daily_sale_starts_at IS NOT NULL AND daily_sale_ends_at IS NOT NULL));
