CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_signup_bonus_uniq
ON user_affiliate_ledger (user_id, source_user_id)
WHERE action = 'signup_bonus' AND source_user_id IS NOT NULL;
