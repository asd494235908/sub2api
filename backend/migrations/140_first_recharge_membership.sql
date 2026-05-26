CREATE TABLE IF NOT EXISTS first_recharge_chances (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  available INTEGER NOT NULL DEFAULT 1,
  consumed INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, tier_id)
);

CREATE INDEX IF NOT EXISTS idx_first_recharge_chances_user ON first_recharge_chances(user_id);

CREATE TABLE IF NOT EXISTS first_recharge_grant_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  chances INTEGER NOT NULL,
  operator VARCHAR(64) NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_first_recharge_grant_logs_user ON first_recharge_grant_logs(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS first_recharge_consumption_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  order_id BIGINT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_first_recharge_consumption_logs_user ON first_recharge_consumption_logs(user_id, created_at DESC);
