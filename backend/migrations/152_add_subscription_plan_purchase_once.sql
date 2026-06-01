ALTER TABLE subscription_plans
  ADD COLUMN IF NOT EXISTS purchase_once_per_active_subscription BOOLEAN NOT NULL DEFAULT FALSE;
