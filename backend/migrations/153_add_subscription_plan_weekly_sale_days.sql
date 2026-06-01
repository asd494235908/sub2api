ALTER TABLE subscription_plans
  ADD COLUMN IF NOT EXISTS weekly_sale_days JSONB NOT NULL DEFAULT '[]'::jsonb;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_subscription_plans_weekly_sale_days_array'
      AND conrelid = 'subscription_plans'::regclass
  ) THEN
    ALTER TABLE subscription_plans
      ADD CONSTRAINT chk_subscription_plans_weekly_sale_days_array
        CHECK (jsonb_typeof(weekly_sale_days) = 'array');
  END IF;
END $$;
