package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
)

const (
	SettingKeyFirstRechargeEnabled = "first_recharge_enabled"
	SettingKeyFirstRechargeConfig  = "first_recharge_config"

	RedeemTypeFirstRechargeBonus = "first_recharge_bonus"
)

type FirstRechargeConfig struct {
	Tiers []FirstRechargeTier `json:"tiers"`
}

type FirstRechargeTier struct {
	ID          string  `json:"id"`
	PayAmount   float64 `json:"pay_amount"`
	BonusAmount float64 `json:"bonus_amount"`
	Enabled     bool    `json:"enabled"`
	SortOrder   int     `json:"sort_order"`
}

type FirstRechargeConfigSummary struct {
	Enabled bool                `json:"enabled"`
	Config  FirstRechargeConfig `json:"config"`
}

type FirstRechargeGrantResult struct {
	UserID         int64     `json:"user_id"`
	TierID         string    `json:"tier_id"`
	GrantedChances int       `json:"granted_chances"`
	Available      int       `json:"available"`
	CreatedAt      time.Time `json:"created_at"`
}

type FirstRechargeBulkChanceMode string

const (
	FirstRechargeBulkChanceModeGrant FirstRechargeBulkChanceMode = "grant"
	FirstRechargeBulkChanceModeReset FirstRechargeBulkChanceMode = "reset"
)

type FirstRechargeBulkChanceResult struct {
	TierID        string                      `json:"tier_id"`
	Mode          FirstRechargeBulkChanceMode `json:"mode"`
	Chances       int                         `json:"chances"`
	AffectedUsers int                         `json:"affected_users"`
	CreatedAt     time.Time                   `json:"created_at"`
}

func (s *PaymentService) firstRechargeEnabled(ctx context.Context) bool {
	if s == nil || s.configService == nil {
		return false
	}
	value, err := s.configService.settingRepo.GetValue(ctx, SettingKeyFirstRechargeEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *PaymentService) GetFirstRechargeConfig(ctx context.Context) (*FirstRechargeConfig, bool, error) {
	enabled := s.firstRechargeEnabled(ctx)
	if s == nil || s.configService == nil {
		return nil, enabled, fmt.Errorf("payment service is not initialized")
	}
	raw, err := s.configService.settingRepo.GetValue(ctx, SettingKeyFirstRechargeConfig)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, enabled, fmt.Errorf("get first recharge config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg, normErr := normalizeFirstRechargeConfig(nil)
		return cfg, enabled, normErr
	}
	var cfg FirstRechargeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, enabled, infraerrors.InternalServer("FIRST_RECHARGE_CONFIG_INVALID", "first recharge config is invalid")
	}
	normalized, normErr := normalizeFirstRechargeConfig(&cfg)
	if normErr != nil {
		fallback, fallbackErr := normalizeFirstRechargeConfig(nil)
		return fallback, enabled, fallbackErr
	}
	return normalized, enabled, nil
}

func (s *PaymentService) UpdateFirstRechargeConfig(ctx context.Context, enabled bool, cfg *FirstRechargeConfig) (*FirstRechargeConfig, error) {
	if s == nil || s.configService == nil {
		return nil, fmt.Errorf("payment service is not initialized")
	}
	normalized, err := normalizeFirstRechargeConfig(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal first recharge config: %w", err)
	}
	updates := map[string]string{
		SettingKeyFirstRechargeEnabled: strconv.FormatBool(enabled),
		SettingKeyFirstRechargeConfig:  string(raw),
	}
	if err := s.configService.settingRepo.SetMultiple(ctx, updates); err != nil {
		return nil, fmt.Errorf("save first recharge config: %w", err)
	}
	return normalized, nil
}

func (s *PaymentService) GrantFirstRechargeChance(ctx context.Context, userID int64, tierID string, chances int, operator string, note string) (*FirstRechargeGrantResult, error) {
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "user id is required")
	}
	tierID = strings.TrimSpace(tierID)
	if tierID == "" {
		return nil, infraerrors.BadRequest("INVALID_TIER", "tier id is required")
	}
	if chances <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CHANCES", "chances must be greater than 0")
	}
	if err := s.ensureFirstRechargeTables(ctx); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := firstRechargeAddChance(ctx, s.entClient, userID, tierID, chances, strings.TrimSpace(operator), strings.TrimSpace(note), now); err != nil {
		return nil, err
	}
	state, err := firstRechargeLoadChance(ctx, s.entClient, userID, tierID)
	if err != nil {
		return nil, err
	}
	return &FirstRechargeGrantResult{
		UserID:         userID,
		TierID:         tierID,
		GrantedChances: chances,
		Available:      state.Available,
		CreatedAt:      now,
	}, nil
}

func (s *PaymentService) BulkUpdateFirstRechargeChances(ctx context.Context, tierID string, chances int, mode FirstRechargeBulkChanceMode, operator string, note string) (*FirstRechargeBulkChanceResult, error) {
	tierID = strings.TrimSpace(tierID)
	if tierID == "" {
		return nil, infraerrors.BadRequest("INVALID_TIER", "tier id is required")
	}
	if chances < 0 {
		return nil, infraerrors.BadRequest("INVALID_CHANCES", "chances must be >= 0")
	}
	if mode != FirstRechargeBulkChanceModeGrant && mode != FirstRechargeBulkChanceModeReset {
		return nil, infraerrors.BadRequest("INVALID_MODE", "mode must be grant or reset")
	}
	if mode == FirstRechargeBulkChanceModeGrant && chances <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CHANCES", "grant chances must be greater than 0")
	}
	if err := s.ensureFirstRechargeTables(ctx); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	affected, err := firstRechargeBulkUpdateChances(ctx, s.entClient, tierID, chances, mode, strings.TrimSpace(operator), strings.TrimSpace(note), now)
	if err != nil {
		return nil, err
	}
	return &FirstRechargeBulkChanceResult{
		TierID:        tierID,
		Mode:          mode,
		Chances:       chances,
		AffectedUsers: affected,
		CreatedAt:     now,
	}, nil
}

func (s *PaymentService) ApplyFirstRechargeBonusForOrder(ctx context.Context, order *dbent.PaymentOrder) error {
	if order == nil || order.OrderType != payment.OrderTypeBalance {
		return nil
	}
	if s.redeemService == nil {
		return fmt.Errorf("redeem service is not initialized")
	}
	if err := s.ensureFirstRechargeTables(ctx); err != nil {
		return err
	}
	claimedTierID, alreadyClaimed, err := firstRechargeClaimedTier(ctx, s.entClient, order.ID)
	if err != nil {
		return err
	}
	if alreadyClaimed {
		return s.redeemClaimedFirstRechargeBonus(ctx, order, claimedTierID)
	}
	if s.configService == nil {
		return nil
	}

	cfg, enabled, err := s.GetFirstRechargeConfig(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	tier := firstRechargeMatchTier(cfg, order.PayAmount)
	if tier == nil {
		return nil
	}
	code := firstRechargeBonusCode(order.ID, tier.ID)
	existing, lookupErr := s.redeemService.GetByCode(ctx, code)
	if lookupErr != nil && !errors.Is(lookupErr, ErrRedeemCodeNotFound) {
		return lookupErr
	}
	codeExists := lookupErr == nil && existing != nil

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin first recharge reward claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	claimed, err := firstRechargeClaimChance(txCtx, tx.Client(), order.UserID, tier.ID, order.ID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !claimed {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit unclaimed first recharge reward: %w", err)
		}
		return nil
	}
	if !codeExists {
		rc := &RedeemCode{
			Code:   code,
			Type:   RedeemTypeFirstRechargeBonus,
			Value:  tier.BonusAmount,
			Status: StatusUnused,
			Notes:  fmt.Sprintf("first recharge bonus tier=%s pay_amount=%.2f order_id=%d", tier.ID, order.PayAmount, order.ID),
		}
		if err := s.redeemService.CreateCode(txCtx, rc); err != nil {
			return fmt.Errorf("create first recharge bonus redeem code: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit first recharge reward claim: %w", err)
	}
	return s.redeemClaimedFirstRechargeBonus(ctx, order, tier.ID)
}

func (s *PaymentService) redeemClaimedFirstRechargeBonus(ctx context.Context, order *dbent.PaymentOrder, tierID string) error {
	code := firstRechargeBonusCode(order.ID, tierID)
	existing, err := s.redeemService.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("load claimed first recharge bonus code %q: %w", code, err)
	}
	if existing == nil {
		return fmt.Errorf("load claimed first recharge bonus code %q: empty result", code)
	}
	if existing.IsUsed() {
		return nil
	}
	if _, err := s.redeemService.Redeem(ContextSkipRedeemAffiliate(ctx), order.UserID, code); err != nil {
		return fmt.Errorf("redeem first recharge bonus: %w", err)
	}
	return nil
}

func firstRechargeBonusCode(orderID int64, tierID string) string {
	return fmt.Sprintf("FIRST-%d-%s", orderID, sanitizeFirstRechargeCodePart(tierID))
}

func firstRechargeClaimedTier(ctx context.Context, client *dbent.Client, orderID int64) (string, bool, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return "", false, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT tier_id FROM first_recharge_consumption_logs WHERE order_id = ? LIMIT 1`)
	rows, err := execCtx.QueryContext(ctx, query, orderID)
	if err != nil {
		return "", false, fmt.Errorf("load first recharge consumption: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return "", false, fmt.Errorf("iterate first recharge consumption: %w", err)
		}
		return "", false, nil
	}
	var tierID string
	if err := rows.Scan(&tierID); err != nil {
		return "", false, fmt.Errorf("scan first recharge consumption: %w", err)
	}
	return tierID, true, nil
}

func normalizeFirstRechargeConfig(cfg *FirstRechargeConfig) (*FirstRechargeConfig, error) {
	if cfg == nil {
		return &FirstRechargeConfig{Tiers: []FirstRechargeTier{}}, nil
	}
	out := &FirstRechargeConfig{Tiers: make([]FirstRechargeTier, 0, len(cfg.Tiers))}
	seen := map[string]struct{}{}
	for _, tier := range cfg.Tiers {
		tier.ID = strings.TrimSpace(tier.ID)
		if tier.ID == "" {
			return nil, infraerrors.BadRequest("FIRST_RECHARGE_TIER_INVALID", "first recharge tier id is required")
		}
		if _, ok := seen[tier.ID]; ok {
			return nil, infraerrors.BadRequest("FIRST_RECHARGE_TIER_INVALID", "first recharge tier id must be unique")
		}
		seen[tier.ID] = struct{}{}
		if !isFinitePositive(tier.PayAmount) {
			return nil, infraerrors.BadRequest("FIRST_RECHARGE_TIER_INVALID", "first recharge tier pay amount must be > 0")
		}
		if math.IsNaN(tier.BonusAmount) || math.IsInf(tier.BonusAmount, 0) || tier.BonusAmount <= 0 {
			return nil, infraerrors.BadRequest("FIRST_RECHARGE_TIER_INVALID", "first recharge tier bonus amount must be > 0")
		}
		out.Tiers = append(out.Tiers, tier)
	}
	sort.SliceStable(out.Tiers, func(i, j int) bool {
		if out.Tiers[i].SortOrder == out.Tiers[j].SortOrder {
			return out.Tiers[i].PayAmount < out.Tiers[j].PayAmount
		}
		return out.Tiers[i].SortOrder < out.Tiers[j].SortOrder
	})
	return out, nil
}

func firstRechargeMatchTier(cfg *FirstRechargeConfig, payAmount float64) *FirstRechargeTier {
	if cfg == nil {
		return nil
	}
	for i := range cfg.Tiers {
		tier := cfg.Tiers[i]
		if !tier.Enabled || math.Abs(tier.PayAmount-payAmount) > amountToleranceCNY {
			continue
		}
		copy := tier
		return &copy
	}
	return nil
}

func (s *PaymentService) ResolveMemberMultiplier(ctx context.Context, user *User, apiKey *APIKey, subscription *UserSubscription) float64 {
	if s == nil || user == nil || !s.isRechargeKeyBilling(apiKey, subscription) {
		return 1
	}
	state, err := s.ResolveMemberLevel(ctx, user)
	if err != nil || state == nil || state.RateMultiplier <= 0 || math.IsNaN(state.RateMultiplier) || math.IsInf(state.RateMultiplier, 0) {
		return s.resolveAffiliateIdentityMultiplier(ctx, user.ID, 1)
	}
	return s.resolveAffiliateIdentityMultiplier(ctx, user.ID, state.RateMultiplier)
}

func (s *PaymentService) resolveAffiliateIdentityMultiplier(ctx context.Context, userID int64, currentMultiplier float64) float64 {
	if s == nil || s.affiliateService == nil || userID <= 0 {
		return currentMultiplier
	}
	return s.affiliateService.ResolveAffiliateIdentityMultiplier(ctx, userID, currentMultiplier)
}

func (s *PaymentService) ensureFirstRechargeTables(ctx context.Context) error {
	if s == nil || s.entClient == nil {
		return fmt.Errorf("payment service is not initialized")
	}
	execCtx, err := luckyWheelExecContext(ctx, s.entClient)
	if err != nil {
		return err
	}
	for _, statement := range firstRechargeDDLForDialect(s.entClient.Driver().Dialect()) {
		if _, err := execCtx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure first recharge tables: %w", err)
		}
	}
	return nil
}

func firstRechargeDDLForDialect(dialectName string) []string {
	idType := "INTEGER PRIMARY KEY AUTOINCREMENT"
	timestampType := "TIMESTAMP"
	if dialectName == dialect.Postgres {
		idType = "BIGSERIAL PRIMARY KEY"
		timestampType = "TIMESTAMPTZ"
	}
	return []string{
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS first_recharge_chances (
  id %s,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  available INTEGER NOT NULL DEFAULT 1,
  consumed INTEGER NOT NULL DEFAULT 0,
  created_at %s NOT NULL,
  updated_at %s NOT NULL,
  UNIQUE(user_id, tier_id)
)`, idType, timestampType, timestampType),
		`CREATE INDEX IF NOT EXISTS idx_first_recharge_chances_user ON first_recharge_chances(user_id)`,
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS first_recharge_grant_logs (
  id %s,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  chances INTEGER NOT NULL,
  operator VARCHAR(64) NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_at %s NOT NULL
)`, idType, timestampType),
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS first_recharge_consumption_logs (
  id %s,
  user_id BIGINT NOT NULL,
  tier_id VARCHAR(64) NOT NULL,
  order_id BIGINT NOT NULL UNIQUE,
  created_at %s NOT NULL
)`, idType, timestampType),
	}
}

type firstRechargeChanceState struct {
	Available int
	Consumed  int
}

func firstRechargeLoadChance(ctx context.Context, client *dbent.Client, userID int64, tierID string) (*firstRechargeChanceState, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT available, consumed
FROM first_recharge_chances
WHERE user_id = ? AND tier_id = ?`)
	rows, err := execCtx.QueryContext(ctx, query, userID, tierID)
	if err != nil {
		return nil, fmt.Errorf("load first recharge chance: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return &firstRechargeChanceState{Available: 1, Consumed: 0}, nil
	}
	var state firstRechargeChanceState
	if err := rows.Scan(&state.Available, &state.Consumed); err != nil {
		return nil, fmt.Errorf("scan first recharge chance: %w", err)
	}
	return &state, nil
}

func firstRechargeAddChance(ctx context.Context, client *dbent.Client, userID int64, tierID string, chances int, operator string, note string, now time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	if client.Driver().Dialect() == dialect.Postgres {
		_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
VALUES ($1, $2, $3, 0, $4, $5)
ON CONFLICT (user_id, tier_id)
DO UPDATE SET available = first_recharge_chances.available + EXCLUDED.available, updated_at = EXCLUDED.updated_at`, userID, tierID, chances, now, now)
	} else {
		_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, ?)
ON CONFLICT(user_id, tier_id)
DO UPDATE SET available = available + excluded.available, updated_at = excluded.updated_at`, userID, tierID, chances, now, now)
	}
	if err != nil {
		return fmt.Errorf("add first recharge chance: %w", err)
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
INSERT INTO first_recharge_grant_logs (user_id, tier_id, chances, operator, note, created_at)
VALUES (?, ?, ?, ?, ?, ?)`)
	if _, err := execCtx.ExecContext(ctx, query, userID, tierID, chances, operator, note, now); err != nil {
		return fmt.Errorf("insert first recharge grant log: %w", err)
	}
	return nil
}

func firstRechargeBulkUpdateChances(ctx context.Context, client *dbent.Client, tierID string, chances int, mode FirstRechargeBulkChanceMode, operator string, note string, now time.Time) (int, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	dialectName := client.Driver().Dialect()
	if dialectName == dialect.Postgres {
		if mode == FirstRechargeBulkChanceModeReset {
			_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
SELECT id, $1, $2, 0, $3, $4
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id, tier_id)
DO UPDATE SET available = EXCLUDED.available, consumed = 0, updated_at = EXCLUDED.updated_at`, tierID, chances, now, now)
		} else {
			_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
SELECT id, $1, $2, 0, $3, $4
FROM users
WHERE deleted_at IS NULL
ON CONFLICT (user_id, tier_id)
DO UPDATE SET available = first_recharge_chances.available + EXCLUDED.available, updated_at = EXCLUDED.updated_at`, tierID, chances, now, now)
		}
	} else {
		if mode == FirstRechargeBulkChanceModeReset {
			_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
SELECT id, ?, ?, 0, ?, ?
FROM users
WHERE deleted_at IS NULL
ON CONFLICT(user_id, tier_id)
DO UPDATE SET available = excluded.available, consumed = 0, updated_at = excluded.updated_at`, tierID, chances, now, now)
		} else {
			_, err = execCtx.ExecContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
SELECT id, ?, ?, 0, ?, ?
FROM users
WHERE deleted_at IS NULL
ON CONFLICT(user_id, tier_id)
DO UPDATE SET available = available + excluded.available, updated_at = excluded.updated_at`, tierID, chances, now, now)
		}
	}
	if err != nil {
		return 0, fmt.Errorf("bulk update first recharge chances: %w", err)
	}

	logNote := strings.TrimSpace(note)
	if logNote == "" {
		logNote = fmt.Sprintf("bulk %s first recharge chances", mode)
	}
	query := luckyWheelBindVars(dialectName, `
INSERT INTO first_recharge_grant_logs (user_id, tier_id, chances, operator, note, created_at)
SELECT id, ?, ?, ?, ?, ?
FROM users
WHERE deleted_at IS NULL`)
	if _, err := execCtx.ExecContext(ctx, query, tierID, chances, operator, logNote, now); err != nil {
		return 0, fmt.Errorf("insert bulk first recharge grant logs: %w", err)
	}
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	rows, err := execCtx.QueryContext(ctx, countQuery)
	if err != nil {
		return 0, fmt.Errorf("count bulk first recharge users: %w", err)
	}
	defer func() { _ = rows.Close() }()
	affected := 0
	if rows.Next() {
		if err := rows.Scan(&affected); err != nil {
			return 0, fmt.Errorf("scan bulk first recharge users: %w", err)
		}
	}
	return affected, nil
}

func firstRechargeClaimChance(ctx context.Context, client *dbent.Client, userID int64, tierID string, orderID int64, now time.Time) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("claim first recharge chance: ent client is nil")
	}
	if dbent.TxFromContext(ctx) != nil {
		return firstRechargeClaimChanceInTx(ctx, client, userID, tierID, orderID, now)
	}
	tx, err := client.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin first recharge claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	claimed, err := firstRechargeClaimChanceInTx(txCtx, tx.Client(), userID, tierID, orderID, now)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit first recharge claim: %w", err)
	}
	return claimed, nil
}

func firstRechargeClaimChanceInTx(ctx context.Context, client *dbent.Client, userID int64, tierID string, orderID int64, now time.Time) (bool, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return false, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT 1 FROM first_recharge_consumption_logs WHERE order_id = ?`)
	rows, err := execCtx.QueryContext(ctx, query, orderID)
	if err != nil {
		return false, fmt.Errorf("check first recharge consumption log: %w", err)
	}
	alreadyClaimed := rows.Next()
	if closeErr := rows.Close(); closeErr != nil {
		return false, fmt.Errorf("close first recharge consumption rows: %w", closeErr)
	}
	if alreadyClaimed {
		return true, nil
	}
	if client.Driver().Dialect() == dialect.Postgres {
		rows, err := execCtx.QueryContext(ctx, `
INSERT INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
VALUES ($1, $2, 1, 0, $3, $4)
ON CONFLICT (user_id, tier_id) DO NOTHING
RETURNING id`, userID, tierID, now, now)
		if err != nil {
			return false, fmt.Errorf("ensure first recharge chance: %w", err)
		}
		_ = rows.Close()
	} else {
		if _, err := execCtx.ExecContext(ctx, `
INSERT OR IGNORE INTO first_recharge_chances (user_id, tier_id, available, consumed, created_at, updated_at)
VALUES (?, ?, 1, 0, ?, ?)`, userID, tierID, now, now); err != nil {
			return false, fmt.Errorf("ensure first recharge chance: %w", err)
		}
	}
	query = luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE first_recharge_chances
SET available = available - 1, consumed = consumed + 1, updated_at = ?
WHERE user_id = ? AND tier_id = ? AND available > 0`)
	res, err := execCtx.ExecContext(ctx, query, now, userID, tierID)
	if err != nil {
		return false, fmt.Errorf("claim first recharge chance: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return false, nil
	}
	query = luckyWheelBindVars(client.Driver().Dialect(), `
INSERT INTO first_recharge_consumption_logs (user_id, tier_id, order_id, created_at)
VALUES (?, ?, ?, ?)`)
	if _, err := execCtx.ExecContext(ctx, query, userID, tierID, orderID, now); err != nil {
		return false, fmt.Errorf("insert first recharge consumption log: %w", err)
	}
	return true, nil
}

func sanitizeFirstRechargeCodePart(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			_, _ = b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "TIER"
	}
	return b.String()
}
