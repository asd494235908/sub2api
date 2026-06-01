ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS subscription_rebate_base_amount DECIMAL(20,8);
