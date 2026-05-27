UPDATE redeem_codes
SET status = 'used'
WHERE status = 'unused'
  AND (used_by IS NOT NULL OR used_at IS NOT NULL);
