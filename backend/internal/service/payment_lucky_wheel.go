package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
)

const SettingKeyLuckyWheelConfig = "lucky_wheel_config"

const (
	LuckyWheelInviteBonusConsumeNextSessionOnce = "next_session_once"
)

var (
	luckyWheelNow         = func() time.Time { return time.Now().UTC() }
	luckyWheelRandFloat64 = rand.Float64
)

type LuckyWheelConfig struct {
	EligibleOrderTypes  []string                     `json:"eligible_order_types"`
	MultiplierStep      float64                      `json:"multiplier_step"`
	GlobalMaxMultiplier float64                      `json:"global_max_multiplier"`
	IntroText           string                       `json:"intro_text"`
	RulesTitle          string                       `json:"rules_title"`
	RulesItems          []string                     `json:"rules_items"`
	AmountTiers         []LuckyWheelAmountTier       `json:"amount_tiers"`
	InviteBonus         LuckyWheelInviteBonusConfig  `json:"invite_bonus"`
	GoldenWindow        LuckyWheelGoldenWindowConfig `json:"golden_window"`
}

type LuckyWheelAmountTier struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	MinAmount     float64  `json:"min_amount"`
	MaxAmount     *float64 `json:"max_amount,omitempty"`
	MinMultiplier float64  `json:"min_multiplier"`
	MaxMultiplier float64  `json:"max_multiplier"`
	DrawCount     int      `json:"draw_count"`
}

type LuckyWheelInviteBonusConfig struct {
	Enabled          bool    `json:"enabled"`
	QualifyingAmount float64 `json:"qualifying_amount"`
	BonusPerInvitee  float64 `json:"bonus_per_invitee"`
	MaxBonus         float64 `json:"max_bonus"`
	ConsumePolicy    string  `json:"consume_policy"`
}

type LuckyWheelGoldenWindowConfig struct {
	Enabled    bool    `json:"enabled"`
	Timezone   string  `json:"timezone"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	MinAmount  float64 `json:"min_amount"`
	ExtraDraws int     `json:"extra_draws"`
	DailyQuota int     `json:"daily_quota"`
}

type LuckyWheelSession struct {
	ID                     int64                  `json:"id"`
	UserID                 int64                  `json:"user_id"`
	SourceOrderID          int64                  `json:"source_order_id"`
	SourceOrderType        string                 `json:"source_order_type"`
	SourcePayAmount        float64                `json:"source_pay_amount"`
	MatchedTierID          string                 `json:"matched_tier_id"`
	MatchedTierName        string                 `json:"matched_tier_name"`
	MinMultiplier          float64                `json:"min_multiplier"`
	MaxMultiplier          float64                `json:"max_multiplier"`
	TotalDraws             int                    `json:"total_draws"`
	CompletedDraws         int                    `json:"completed_draws"`
	RemainingDraws         int                    `json:"remaining_draws"`
	BestMultiplier         float64                `json:"best_multiplier"`
	InviteBonusMultiplier  float64                `json:"invite_bonus_multiplier"`
	GoldenWindowExtraDraws int                    `json:"golden_window_extra_draws"`
	Settled                bool                   `json:"settled"`
	SettledBonusAmount     *float64               `json:"settled_bonus_amount,omitempty"`
	SettledAt              *time.Time             `json:"settled_at,omitempty"`
	DrawRecords            []LuckyWheelDrawRecord `json:"draw_records,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

type LuckyWheelDrawRecord struct {
	ID                    int64     `json:"id"`
	SessionID             int64     `json:"session_id"`
	UserID                int64     `json:"user_id"`
	DrawIndex             int       `json:"draw_index"`
	BaseMultiplier        float64   `json:"base_multiplier"`
	InviteBonusMultiplier float64   `json:"invite_bonus_multiplier"`
	FinalMultiplier       float64   `json:"final_multiplier"`
	IsBest                bool      `json:"is_best"`
	CreatedAt             time.Time `json:"created_at"`
}

type LuckyWheelSummary struct {
	Enabled         bool                `json:"enabled"`
	Config          LuckyWheelConfig    `json:"config"`
	ActiveSession   *LuckyWheelSession  `json:"active_session,omitempty"`
	PendingSessions []LuckyWheelSession `json:"pending_sessions"`
	HistorySessions []LuckyWheelSession `json:"history_sessions"`
}

type LuckyWheelDrawResult struct {
	SessionID          int64                `json:"session_id"`
	DrawRecord         LuckyWheelDrawRecord `json:"draw_record"`
	BestMultiplier     float64              `json:"best_multiplier"`
	RemainingDraws     int                  `json:"remaining_draws"`
	Settled            bool                 `json:"settled"`
	SettledBonusAmount *float64             `json:"settled_bonus_amount,omitempty"`
	Session            *LuckyWheelSession   `json:"session,omitempty"`
}

type LuckyWheelMultiplierStat struct {
	Multiplier float64 `json:"multiplier"`
	DrawCount  int64   `json:"draw_count"`
}

type LuckyWheelStats struct {
	Enabled                bool                       `json:"enabled"`
	TotalSessions          int64                      `json:"total_sessions"`
	PendingSessions        int64                      `json:"pending_sessions"`
	SettledSessions        int64                      `json:"settled_sessions"`
	TotalBonusAmount       float64                    `json:"total_bonus_amount"`
	RecentSessions         []LuckyWheelSession        `json:"recent_sessions"`
	MultiplierStats        []LuckyWheelMultiplierStat `json:"multiplier_stats"`
	GoldenWindowUsedToday  int                        `json:"golden_window_used_today"`
	GoldenWindowDailyQuota int                        `json:"golden_window_daily_quota"`
}

func (s *PaymentService) luckyWheelEnabled(ctx context.Context) bool {
	if s == nil || s.configService == nil {
		return false
	}
	if !s.configService.IsPaymentEnabled(ctx) {
		return false
	}
	value, err := s.configService.settingRepo.GetValue(ctx, SettingKeyLuckyWheelEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *PaymentService) ensureLuckyWheelTables(ctx context.Context) error {
	if s == nil || s.entClient == nil {
		return fmt.Errorf("payment service is not initialized")
	}
	execCtx, err := luckyWheelExecContext(ctx, s.entClient)
	if err != nil {
		return err
	}
	for _, statement := range luckyWheelDDLForDialect(s.entClient.Driver().Dialect()) {
		if _, err := execCtx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure lucky wheel tables: %w", err)
		}
	}
	return nil
}

func (s *PaymentService) GetLuckyWheelConfig(ctx context.Context) (*LuckyWheelConfig, bool, error) {
	enabled := s.luckyWheelEnabled(ctx)
	raw, err := s.configService.settingRepo.GetValue(ctx, SettingKeyLuckyWheelConfig)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, false, fmt.Errorf("get lucky wheel config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg, normErr := normalizeLuckyWheelConfig(nil)
		return cfg, enabled, normErr
	}
	var cfg LuckyWheelConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, enabled, infraerrors.InternalServer("LUCKY_WHEEL_CONFIG_INVALID", "lucky wheel config is invalid")
	}
	normalized, normErr := normalizeLuckyWheelConfig(&cfg)
	if normErr != nil {
		// Legacy prize/tier configs should not brick the new feature. Fall back to defaults.
		fallback, fallbackErr := normalizeLuckyWheelConfig(nil)
		return fallback, enabled, fallbackErr
	}
	return normalized, enabled, nil
}

func (s *PaymentService) UpdateLuckyWheelConfig(ctx context.Context, enabled bool, cfg *LuckyWheelConfig) (*LuckyWheelConfig, error) {
	normalized, err := normalizeLuckyWheelConfig(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal lucky wheel config: %w", err)
	}
	updates := map[string]string{
		SettingKeyLuckyWheelEnabled: strconvFormatBool(enabled),
		SettingKeyLuckyWheelConfig:  string(raw),
	}
	if err := s.configService.settingRepo.SetMultiple(ctx, updates); err != nil {
		return nil, fmt.Errorf("save lucky wheel config: %w", err)
	}
	return normalized, nil
}

func (s *PaymentService) GrantLuckyWheelChanceForOrder(ctx context.Context, order *dbent.PaymentOrder) error {
	if order == nil {
		return nil
	}
	if !s.luckyWheelEnabled(ctx) {
		return nil
	}
	if err := s.ensureLuckyWheelTables(ctx); err != nil {
		return err
	}
	cfg, _, err := s.GetLuckyWheelConfig(ctx)
	if err != nil {
		return err
	}
	if !luckyWheelOrderTypeEligible(cfg, order.OrderType) {
		return nil
	}
	tier, ok := matchLuckyWheelAmountTier(cfg.AmountTiers, order.PayAmount)
	if !ok {
		return nil
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin lucky wheel session tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	existing, err := luckyWheelGetSessionByOrderID(txCtx, tx.Client(), int64(order.ID))
	if err != nil {
		return err
	}
	if existing != nil {
		return tx.Commit()
	}

	now := luckyWheelNow()
	if err := luckyWheelAwardInviteBonusEvent(txCtx, tx.Client(), cfg, order, now); err != nil {
		return err
	}

	extraDraws, err := luckyWheelTryClaimGoldenWindowBonus(txCtx, tx.Client(), cfg, order, now)
	if err != nil {
		return err
	}

	inviteBonus, err := luckyWheelPendingInviteBonus(txCtx, tx.Client(), cfg, order.UserID)
	if err != nil {
		return err
	}

	session := &LuckyWheelSession{
		UserID:                 order.UserID,
		SourceOrderID:          int64(order.ID),
		SourceOrderType:        order.OrderType,
		SourcePayAmount:        order.PayAmount,
		MatchedTierID:          tier.ID,
		MatchedTierName:        tier.Name,
		MinMultiplier:          tier.MinMultiplier,
		MaxMultiplier:          tier.MaxMultiplier,
		TotalDraws:             tier.DrawCount + extraDraws,
		CompletedDraws:         0,
		BestMultiplier:         0,
		InviteBonusMultiplier:  inviteBonus,
		GoldenWindowExtraDraws: extraDraws,
		Settled:                false,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	sessionID, err := luckyWheelInsertSession(txCtx, tx.Client(), session)
	if err != nil {
		return err
	}
	if inviteBonus > 0 {
		if err := luckyWheelConsumeInviteBonuses(txCtx, tx.Client(), order.UserID, sessionID, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PaymentService) GetLuckyWheelSummary(ctx context.Context, userID int64) (*LuckyWheelSummary, error) {
	if err := s.ensureLuckyWheelTables(ctx); err != nil {
		return nil, err
	}
	cfg, enabled, err := s.GetLuckyWheelConfig(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := luckyWheelListSessionsBySettlement(ctx, s.entClient, userID, false, 20)
	if err != nil {
		return nil, err
	}
	history, err := luckyWheelListSessionsBySettlement(ctx, s.entClient, userID, true, 20)
	if err != nil {
		return nil, err
	}
	if err := luckyWheelAttachDrawRecords(ctx, s.entClient, pending); err != nil {
		return nil, err
	}
	if err := luckyWheelAttachDrawRecords(ctx, s.entClient, history); err != nil {
		return nil, err
	}

	var active *LuckyWheelSession
	if len(pending) > 0 {
		cloned := pending[0]
		active = &cloned
	}
	return &LuckyWheelSummary{
		Enabled:         enabled,
		Config:          *cfg,
		ActiveSession:   active,
		PendingSessions: pending,
		HistorySessions: history,
	}, nil
}

func (s *PaymentService) DrawLuckyWheel(ctx context.Context, userID, sessionID int64) (*LuckyWheelDrawResult, error) {
	if sessionID <= 0 {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_SESSION_REQUIRED", "session_id is required")
	}
	if err := s.ensureLuckyWheelTables(ctx); err != nil {
		return nil, err
	}
	cfg, enabled, err := s.GetLuckyWheelConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, infraerrors.Forbidden("LUCKY_WHEEL_DISABLED", "lucky wheel is disabled")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin lucky wheel draw tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	session, err := luckyWheelLockSession(txCtx, tx.Client(), userID, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Settled {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_SESSION_SETTLED", "lucky wheel session already settled")
	}
	if session.CompletedDraws >= session.TotalDraws {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_SESSION_EXHAUSTED", "lucky wheel session has no draws remaining")
	}

	base := drawLuckyWheelMultiplier(session.MinMultiplier, session.MaxMultiplier, cfg.MultiplierStep)
	effectiveMax := luckyWheelEffectiveSessionMaxMultiplier(session, cfg)
	final := math.Min(base+session.InviteBonusMultiplier, effectiveMax)
	drawIndex := session.CompletedDraws + 1
	best := session.BestMultiplier
	isBest := final >= best
	if isBest {
		best = final
	}
	now := luckyWheelNow()
	record := LuckyWheelDrawRecord{
		SessionID:             session.ID,
		UserID:                userID,
		DrawIndex:             drawIndex,
		BaseMultiplier:        base,
		InviteBonusMultiplier: session.InviteBonusMultiplier,
		FinalMultiplier:       final,
		IsBest:                isBest,
		CreatedAt:             now,
	}
	recordID, err := luckyWheelInsertDrawRecord(txCtx, tx.Client(), &record)
	if err != nil {
		return nil, err
	}
	record.ID = recordID

	settled := drawIndex >= session.TotalDraws
	var settledBonus *float64
	if settled {
		rewardBase, err := luckyWheelRewardBaseAmount(txCtx, tx.Client(), session)
		if err != nil {
			return nil, err
		}
		bonus := roundLuckyWheelAmount(rewardBase * best)
		settledBonus = &bonus
		if bonus > 0 {
			if err := s.userRepo.UpdateBalance(txCtx, userID, bonus); err != nil {
				return nil, fmt.Errorf("apply lucky wheel bonus: %w", err)
			}
		}
	}
	if err := luckyWheelUpdateSessionAfterDraw(txCtx, tx.Client(), session.ID, drawIndex, best, settled, settledBonus, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit lucky wheel draw tx: %w", err)
	}

	updated, err := luckyWheelLoadSession(ctx, s.entClient, session.ID, userID)
	if err != nil {
		return nil, err
	}
	if err := luckyWheelAttachDrawRecords(ctx, s.entClient, []LuckyWheelSession{*updated}); err != nil {
		return nil, err
	}
	return &LuckyWheelDrawResult{
		SessionID:          session.ID,
		DrawRecord:         record,
		BestMultiplier:     best,
		RemainingDraws:     luckyWheelMaxInt(session.TotalDraws-drawIndex, 0),
		Settled:            settled,
		SettledBonusAmount: settledBonus,
		Session:            updated,
	}, nil
}

func (s *PaymentService) GetLuckyWheelStats(ctx context.Context) (*LuckyWheelStats, error) {
	if err := s.ensureLuckyWheelTables(ctx); err != nil {
		return nil, err
	}
	cfg, enabled, err := s.GetLuckyWheelConfig(ctx)
	if err != nil {
		return nil, err
	}
	execCtx, err := luckyWheelExecContext(ctx, s.entClient)
	if err != nil {
		return nil, err
	}

	stats := &LuckyWheelStats{
		Enabled:                enabled,
		GoldenWindowDailyQuota: cfg.GoldenWindow.DailyQuota,
	}
	rows, err := execCtx.QueryContext(ctx, `
SELECT
  COUNT(*) AS total_count,
  SUM(CASE WHEN settled THEN 0 ELSE 1 END) AS pending_count,
  SUM(CASE WHEN settled THEN 1 ELSE 0 END) AS settled_count,
  COALESCE(SUM(settled_bonus_amount), 0) AS total_bonus_amount
FROM lucky_wheel_sessions`)
	if err != nil {
		return nil, fmt.Errorf("query lucky wheel stats: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var total, pending, settled sql.NullInt64
		var totalBonus sql.NullFloat64
		if err := rows.Scan(&total, &pending, &settled, &totalBonus); err != nil {
			return nil, fmt.Errorf("scan lucky wheel stats: %w", err)
		}
		stats.TotalSessions = total.Int64
		stats.PendingSessions = pending.Int64
		stats.SettledSessions = settled.Int64
		stats.TotalBonusAmount = totalBonus.Float64
	}

	stats.RecentSessions, err = luckyWheelListRecentSessions(ctx, s.entClient, 20)
	if err != nil {
		return nil, err
	}
	if err := luckyWheelAttachDrawRecords(ctx, s.entClient, stats.RecentSessions); err != nil {
		return nil, err
	}
	stats.MultiplierStats, err = luckyWheelListMultiplierStats(ctx, s.entClient)
	if err != nil {
		return nil, err
	}
	stats.GoldenWindowUsedToday, err = luckyWheelCountGoldenWindowClaims(ctx, s.entClient, cfg, luckyWheelNow())
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *PaymentService) luckyWheelBalanceRechargeMultiplier(ctx context.Context) (float64, error) {
	if s == nil || s.configService == nil {
		return defaultBalanceRechargeMultiplier, nil
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return 0, fmt.Errorf("get payment config for lucky wheel settlement: %w", err)
	}
	return normalizeBalanceRechargeMultiplier(cfg.BalanceRechargeMultiplier), nil
}

func normalizeLuckyWheelConfig(cfg *LuckyWheelConfig) (*LuckyWheelConfig, error) {
	if cfg == nil || luckyWheelConfigLooksEmptyOrLegacy(cfg) {
		return cloneDefaultLuckyWheelConfig(), nil
	}

	normalized := &LuckyWheelConfig{
		EligibleOrderTypes:  normalizeLuckyWheelOrderTypes(cfg.EligibleOrderTypes),
		MultiplierStep:      cfg.MultiplierStep,
		GlobalMaxMultiplier: cfg.GlobalMaxMultiplier,
		IntroText:           strings.TrimSpace(cfg.IntroText),
		RulesTitle:          strings.TrimSpace(cfg.RulesTitle),
		RulesItems:          make([]string, 0, len(cfg.RulesItems)),
		AmountTiers:         make([]LuckyWheelAmountTier, 0, len(cfg.AmountTiers)),
		InviteBonus:         cfg.InviteBonus,
		GoldenWindow:        cfg.GoldenWindow,
	}
	for _, item := range cfg.RulesItems {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		normalized.RulesItems = append(normalized.RulesItems, item)
	}
	if normalized.IntroText == "" {
		normalized.IntroText = defaultLuckyWheelIntroText
	}
	if normalized.RulesTitle == "" {
		normalized.RulesTitle = defaultLuckyWheelRulesTitle
	}
	if len(normalized.RulesItems) == 0 {
		normalized.RulesItems = append(normalized.RulesItems, defaultLuckyWheelRulesItems...)
	}
	if len(normalized.EligibleOrderTypes) == 0 {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_ORDER_TYPES_REQUIRED", "at least one eligible order type is required")
	}
	if !isFinitePositive(normalized.MultiplierStep) {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_STEP_INVALID", "multiplier step must be greater than 0")
	}
	if !isFinitePositive(normalized.GlobalMaxMultiplier) {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_GLOBAL_MAX_INVALID", "global max multiplier must be greater than 0")
	}
	seenTierIDs := make(map[string]struct{}, len(cfg.AmountTiers))
	for _, tier := range cfg.AmountTiers {
		tier.ID = strings.TrimSpace(tier.ID)
		tier.Name = strings.TrimSpace(tier.Name)
		if tier.ID == "" {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_ID_REQUIRED", "amount tier id is required")
		}
		if _, ok := seenTierIDs[tier.ID]; ok {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_DUPLICATE", "amount tier ids must be unique")
		}
		seenTierIDs[tier.ID] = struct{}{}
		if tier.Name == "" {
			tier.Name = tier.ID
		}
		if tier.MinAmount < 0 || math.IsNaN(tier.MinAmount) || math.IsInf(tier.MinAmount, 0) {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_MIN_INVALID", "tier min amount must be >= 0")
		}
		if tier.MaxAmount != nil {
			maxAmount := *tier.MaxAmount
			if maxAmount < tier.MinAmount || math.IsNaN(maxAmount) || math.IsInf(maxAmount, 0) {
				return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_MAX_INVALID", "tier max amount must be greater than or equal to min amount")
			}
		}
		if !isFinitePositive(tier.MinMultiplier) || !isFinitePositive(tier.MaxMultiplier) || tier.MaxMultiplier < tier.MinMultiplier {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_MULTIPLIER_INVALID", "tier multiplier range is invalid")
		}
		if tier.MaxMultiplier > normalized.GlobalMaxMultiplier {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_MULTIPLIER_INVALID", "tier max multiplier cannot exceed global max multiplier")
		}
		if tier.DrawCount <= 0 {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_DRAW_COUNT_INVALID", "tier draw count must be greater than 0")
		}
		normalized.AmountTiers = append(normalized.AmountTiers, tier)
	}
	if len(normalized.AmountTiers) == 0 {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIERS_REQUIRED", "at least one amount tier is required")
	}
	sort.Slice(normalized.AmountTiers, func(i, j int) bool {
		if normalized.AmountTiers[i].MinAmount == normalized.AmountTiers[j].MinAmount {
			return normalized.AmountTiers[i].ID < normalized.AmountTiers[j].ID
		}
		return normalized.AmountTiers[i].MinAmount < normalized.AmountTiers[j].MinAmount
	})
	if normalized.AmountTiers[0].MinAmount > 20 {
		return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_RANGE_INVALID", "first amount tier must start at or below 20")
	}
	prevMax := -1.0
	for i, tier := range normalized.AmountTiers {
		if i > 0 && tier.MinAmount <= prevMax {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_OVERLAP", "amount tiers must not overlap")
		}
		if tier.MaxAmount == nil {
			if i != len(normalized.AmountTiers)-1 {
				return nil, infraerrors.BadRequest("LUCKY_WHEEL_TIER_OPEN_ENDED", "only the last tier can be open ended")
			}
			prevMax = tier.MinAmount
			continue
		}
		prevMax = *tier.MaxAmount
	}
	if normalized.InviteBonus.Enabled {
		if !isFinitePositive(normalized.InviteBonus.QualifyingAmount) {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_INVITE_AMOUNT_INVALID", "invite bonus qualifying amount must be greater than 0")
		}
		if !isFinitePositive(normalized.InviteBonus.BonusPerInvitee) {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_INVITE_BONUS_INVALID", "invite bonus per invitee must be greater than 0")
		}
		if !isFinitePositive(normalized.InviteBonus.MaxBonus) {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_INVITE_MAX_INVALID", "invite bonus max must be greater than 0")
		}
		if strings.TrimSpace(normalized.InviteBonus.ConsumePolicy) == "" {
			normalized.InviteBonus.ConsumePolicy = LuckyWheelInviteBonusConsumeNextSessionOnce
		}
		if normalized.InviteBonus.ConsumePolicy != LuckyWheelInviteBonusConsumeNextSessionOnce {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_INVITE_CONSUME_POLICY_INVALID", "unsupported invite bonus consume policy")
		}
	}
	if normalized.GoldenWindow.Enabled {
		if strings.TrimSpace(normalized.GoldenWindow.Timezone) == "" {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_TZ_REQUIRED", "golden window timezone is required")
		}
		if _, err := time.LoadLocation(normalized.GoldenWindow.Timezone); err != nil {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_TZ_INVALID", "golden window timezone is invalid")
		}
		if _, err := parseLuckyWheelClock(normalized.GoldenWindow.StartTime); err != nil {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_START_INVALID", "golden window start time must be HH:MM")
		}
		if _, err := parseLuckyWheelClock(normalized.GoldenWindow.EndTime); err != nil {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_END_INVALID", "golden window end time must be HH:MM")
		}
		if !isFinitePositive(normalized.GoldenWindow.MinAmount) {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_AMOUNT_INVALID", "golden window min amount must be greater than 0")
		}
		if normalized.GoldenWindow.ExtraDraws <= 0 {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_EXTRA_DRAWS_INVALID", "golden window extra draws must be greater than 0")
		}
		if normalized.GoldenWindow.DailyQuota <= 0 {
			return nil, infraerrors.BadRequest("LUCKY_WHEEL_GOLDEN_DAILY_QUOTA_INVALID", "golden window daily quota must be greater than 0")
		}
	}
	return normalized, nil
}

func luckyWheelConfigLooksEmptyOrLegacy(cfg *LuckyWheelConfig) bool {
	return cfg == nil || (len(cfg.AmountTiers) == 0 && cfg.MultiplierStep == 0 && cfg.GlobalMaxMultiplier == 0)
}

func cloneDefaultLuckyWheelConfig() *LuckyWheelConfig {
	max50 := 50.0
	return &LuckyWheelConfig{
		EligibleOrderTypes:  []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		MultiplierStep:      0.1,
		GlobalMaxMultiplier: 3.0,
		IntroText:           defaultLuckyWheelIntroText,
		RulesTitle:          defaultLuckyWheelRulesTitle,
		RulesItems:          append([]string(nil), defaultLuckyWheelRulesItems...),
		AmountTiers: []LuckyWheelAmountTier{
			{ID: "tier_20_50", Name: "20-50", MinAmount: 20, MaxAmount: &max50, MinMultiplier: 1.1, MaxMultiplier: 2.0, DrawCount: 2},
			{ID: "tier_51_plus", Name: "51+", MinAmount: 51, MaxAmount: nil, MinMultiplier: 1.2, MaxMultiplier: 3.0, DrawCount: 3},
		},
		InviteBonus: LuckyWheelInviteBonusConfig{
			Enabled:          true,
			QualifyingAmount: 20,
			BonusPerInvitee:  0.2,
			MaxBonus:         1.0,
			ConsumePolicy:    LuckyWheelInviteBonusConsumeNextSessionOnce,
		},
		GoldenWindow: LuckyWheelGoldenWindowConfig{
			Enabled:    true,
			Timezone:   "Asia/Shanghai",
			StartTime:  "20:00",
			EndTime:    "22:00",
			MinAmount:  51,
			ExtraDraws: 1,
			DailyQuota: 5,
		},
	}
}

const defaultLuckyWheelIntroText = "按用户实付人民币进入不同倍率档位，多次抽奖取最高倍率；充值按到账平台金额结算奖励，订阅按有效天数 × 日额度结算奖励。"
const defaultLuckyWheelRulesTitle = "活动规则"

var defaultLuckyWheelRulesItems = []string{
	"实付 20-50 元：1.1x-2.0x，保底 1.1x，最多 2 次。",
	"实付 51 元及以上：1.2x-3.0x，保底 1.2x，最多 3 次，黄金窗口可额外 +1 次。",
	"同一笔支付多次抽奖取最高倍率；充值奖励按到账平台金额计算，订阅奖励按有效天数 × 日额度计算。",
	"每邀请 1 位新用户完成满 20 元实付订单，下次会话倍率加成 +0.2x，累计上限 +1.0x。",
	"北京时间 20:00-22:00 的 51 元以上实付订单，前 5 名可额外获得 1 次机会。",
}

func normalizeLuckyWheelOrderTypes(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value != payment.OrderTypeBalance && value != payment.OrderTypeSubscription {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func luckyWheelOrderTypeEligible(cfg *LuckyWheelConfig, orderType string) bool {
	for _, candidate := range cfg.EligibleOrderTypes {
		if candidate == orderType {
			return true
		}
	}
	return false
}

func matchLuckyWheelAmountTier(tiers []LuckyWheelAmountTier, amount float64) (LuckyWheelAmountTier, bool) {
	for _, tier := range tiers {
		if amount < tier.MinAmount {
			continue
		}
		if tier.MaxAmount != nil && amount > *tier.MaxAmount {
			continue
		}
		return tier, true
	}
	return LuckyWheelAmountTier{}, false
}

func luckyWheelRewardBaseAmount(ctx context.Context, client *dbent.Client, session *LuckyWheelSession) (float64, error) {
	if session == nil {
		return 0, nil
	}
	order, err := client.PaymentOrder.Get(ctx, session.SourceOrderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return session.SourcePayAmount, nil
		}
		return 0, fmt.Errorf("load lucky wheel source order: %w", err)
	}
	switch order.OrderType {
	case payment.OrderTypeBalance:
		if order.Amount <= 0 {
			return 0, nil
		}
		return order.Amount, nil
	case payment.OrderTypeSubscription:
		if order.SubscriptionDays == nil || *order.SubscriptionDays <= 0 || order.SubscriptionGroupID == nil || *order.SubscriptionGroupID <= 0 {
			return 0, nil
		}
		g, err := client.Group.Query().
			Where(group.IDEQ(*order.SubscriptionGroupID)).
			Only(ctx)
		if err != nil {
			if dbent.IsNotFound(err) {
				return 0, nil
			}
			return 0, fmt.Errorf("load lucky wheel subscription group: %w", err)
		}
		if g.DailyLimitUsd == nil || *g.DailyLimitUsd <= 0 {
			return 0, nil
		}
		return float64(*order.SubscriptionDays) * *g.DailyLimitUsd, nil
	default:
		if order.Amount <= 0 {
			return 0, nil
		}
		return order.Amount, nil
	}
}

func drawLuckyWheelMultiplier(minMultiplier, maxMultiplier, step float64) float64 {
	if maxMultiplier <= minMultiplier || step <= 0 {
		return roundLuckyWheelMultiplier(minMultiplier)
	}
	steps := int(math.Round((maxMultiplier - minMultiplier) / step))
	if steps <= 0 {
		return roundLuckyWheelMultiplier(minMultiplier)
	}
	index := int(math.Floor(luckyWheelRandFloat64() * float64(steps+1)))
	if index > steps {
		index = steps
	}
	return roundLuckyWheelMultiplier(minMultiplier + (float64(index) * step))
}

func luckyWheelEffectiveSessionMaxMultiplier(session *LuckyWheelSession, cfg *LuckyWheelConfig) float64 {
	if session == nil {
		return cfg.GlobalMaxMultiplier
	}
	effectiveMax := session.MaxMultiplier
	if cfg != nil && len(cfg.AmountTiers) > 0 {
		if tier, ok := matchLuckyWheelAmountTier(cfg.AmountTiers, session.SourcePayAmount); ok {
			effectiveMax = tier.MaxMultiplier
		}
	}
	if cfg != nil && cfg.GlobalMaxMultiplier > 0 {
		effectiveMax = math.Min(effectiveMax, cfg.GlobalMaxMultiplier)
	}
	if effectiveMax <= 0 {
		effectiveMax = session.MaxMultiplier
	}
	return effectiveMax
}

func roundLuckyWheelMultiplier(value float64) float64 {
	return math.Round(value*100) / 100
}

func roundLuckyWheelAmount(value float64) float64 {
	return math.Round(value*100) / 100
}

func parseLuckyWheelClock(value string) (int, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid clock")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("invalid hour")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid minute")
	}
	return hour*60 + minute, nil
}

func luckyWheelGoldenWindowState(cfg LuckyWheelGoldenWindowConfig, now time.Time) (bool, string, error) {
	if !cfg.Enabled {
		return false, "", nil
	}
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return false, "", err
	}
	local := now.In(loc)
	start, err := parseLuckyWheelClock(cfg.StartTime)
	if err != nil {
		return false, "", err
	}
	end, err := parseLuckyWheelClock(cfg.EndTime)
	if err != nil {
		return false, "", err
	}
	current := local.Hour()*60 + local.Minute()
	inWindow := false
	switch {
	case end > start:
		inWindow = current >= start && current < end
	case end < start:
		inWindow = current >= start || current < end
	default:
		inWindow = true
	}
	return inWindow, local.Format("2006-01-02"), nil
}

func luckyWheelAwardInviteBonusEvent(ctx context.Context, client *dbent.Client, cfg *LuckyWheelConfig, order *dbent.PaymentOrder, now time.Time) error {
	if !cfg.InviteBonus.Enabled || order == nil {
		return nil
	}
	if order.PayAmount < cfg.InviteBonus.QualifyingAmount {
		return nil
	}
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT inviter_id
FROM user_affiliates
WHERE user_id = ?
LIMIT 1`)
	rows, err := execCtx.QueryContext(ctx, query, order.UserID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") || strings.Contains(strings.ToLower(err.Error()), "does not exist") {
			return nil
		}
		return fmt.Errorf("query invite bonus inviter: %w", err)
	}
	defer rows.Close()

	var inviterID sql.NullInt64
	if !rows.Next() {
		return nil
	}
	if err := rows.Scan(&inviterID); err != nil {
		return fmt.Errorf("scan invite bonus inviter: %w", err)
	}
	if !inviterID.Valid || inviterID.Int64 <= 0 {
		return nil
	}
	return luckyWheelInsertInviteBonusEvent(ctx, client, inviterID.Int64, order.UserID, int64(order.ID), cfg.InviteBonus.BonusPerInvitee, now)
}

func luckyWheelPendingInviteBonus(ctx context.Context, client *dbent.Client, cfg *LuckyWheelConfig, userID int64) (float64, error) {
	if !cfg.InviteBonus.Enabled || userID <= 0 {
		return 0, nil
	}
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT COALESCE(SUM(bonus_multiplier), 0)
FROM lucky_wheel_invite_bonus_events
WHERE inviter_user_id = ? AND consumed_session_id IS NULL`)
	rows, err := execCtx.QueryContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("query pending invite bonus: %w", err)
	}
	defer rows.Close()
	var total sql.NullFloat64
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, fmt.Errorf("scan pending invite bonus: %w", err)
		}
	}
	return math.Min(total.Float64, cfg.InviteBonus.MaxBonus), nil
}

func luckyWheelConsumeInviteBonuses(ctx context.Context, client *dbent.Client, userID, sessionID int64, consumedAt time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE lucky_wheel_invite_bonus_events
SET consumed_session_id = ?, consumed_at = ?
WHERE inviter_user_id = ? AND consumed_session_id IS NULL`)
	if _, err := execCtx.ExecContext(ctx, query, sessionID, consumedAt, userID); err != nil {
		return fmt.Errorf("consume invite bonuses: %w", err)
	}
	return nil
}

func luckyWheelTryClaimGoldenWindowBonus(ctx context.Context, client *dbent.Client, cfg *LuckyWheelConfig, order *dbent.PaymentOrder, now time.Time) (int, error) {
	if order == nil || !cfg.GoldenWindow.Enabled {
		return 0, nil
	}
	if order.PayAmount < cfg.GoldenWindow.MinAmount {
		return 0, nil
	}
	inWindow, claimDay, err := luckyWheelGoldenWindowState(cfg.GoldenWindow, now)
	if err != nil || !inWindow {
		return 0, err
	}
	for slot := 1; slot <= cfg.GoldenWindow.DailyQuota; slot++ {
		ok, err := luckyWheelInsertGoldenWindowClaim(ctx, client, claimDay, slot, order, now)
		if err != nil {
			return 0, err
		}
		if ok {
			return cfg.GoldenWindow.ExtraDraws, nil
		}
	}
	return 0, nil
}

func luckyWheelCountGoldenWindowClaims(ctx context.Context, client *dbent.Client, cfg *LuckyWheelConfig, now time.Time) (int, error) {
	if !cfg.GoldenWindow.Enabled {
		return 0, nil
	}
	_, claimDay, err := luckyWheelGoldenWindowState(cfg.GoldenWindow, now)
	if err != nil {
		return 0, err
	}
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT COUNT(*)
FROM lucky_wheel_golden_window_claims
WHERE claim_day = ?`)
	rows, err := execCtx.QueryContext(ctx, query, claimDay)
	if err != nil {
		return 0, fmt.Errorf("count golden window claims: %w", err)
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, fmt.Errorf("scan golden window claims: %w", err)
		}
	}
	return count, nil
}

func luckyWheelInsertInviteBonusEvent(ctx context.Context, client *dbent.Client, inviterID, inviteeID, sourceOrderID int64, bonus float64, createdAt time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	if client.Driver().Dialect() == dialect.Postgres {
		_, err = execCtx.ExecContext(ctx, `
INSERT INTO lucky_wheel_invite_bonus_events (
  inviter_user_id, invitee_user_id, source_order_id, bonus_multiplier, created_at
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (invitee_user_id) DO NOTHING`, inviterID, inviteeID, sourceOrderID, bonus, createdAt)
	} else {
		_, err = execCtx.ExecContext(ctx, `
INSERT OR IGNORE INTO lucky_wheel_invite_bonus_events (
  inviter_user_id, invitee_user_id, source_order_id, bonus_multiplier, created_at
) VALUES (?, ?, ?, ?, ?)`, inviterID, inviteeID, sourceOrderID, bonus, createdAt)
	}
	if err != nil {
		return fmt.Errorf("insert invite bonus event: %w", err)
	}
	return nil
}

func luckyWheelInsertGoldenWindowClaim(ctx context.Context, client *dbent.Client, claimDay string, slot int, order *dbent.PaymentOrder, createdAt time.Time) (bool, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return false, err
	}
	var result sql.Result
	if client.Driver().Dialect() == dialect.Postgres {
		result, err = execCtx.ExecContext(ctx, `
INSERT INTO lucky_wheel_golden_window_claims (
  claim_day, slot_number, source_order_id, user_id, created_at
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (claim_day, slot_number) DO NOTHING`, claimDay, slot, int64(order.ID), order.UserID, createdAt)
	} else {
		result, err = execCtx.ExecContext(ctx, `
INSERT OR IGNORE INTO lucky_wheel_golden_window_claims (
  claim_day, slot_number, source_order_id, user_id, created_at
) VALUES (?, ?, ?, ?, ?)`, claimDay, slot, int64(order.ID), order.UserID, createdAt)
	}
	if err != nil {
		return false, fmt.Errorf("insert golden window claim: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read golden window claim rows: %w", err)
	}
	return affected > 0, nil
}

func luckyWheelInsertSession(ctx context.Context, client *dbent.Client, session *LuckyWheelSession) (int64, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	d := client.Driver().Dialect()
	if d == dialect.Postgres {
		rows, err := execCtx.QueryContext(ctx, `
INSERT INTO lucky_wheel_sessions (
  user_id, source_order_id, source_order_type, source_pay_amount,
  matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
  total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
  golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
  created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, FALSE, NULL, NULL, $14, $14)
RETURNING id`,
			session.UserID, session.SourceOrderID, session.SourceOrderType, session.SourcePayAmount,
			session.MatchedTierID, session.MatchedTierName, session.MinMultiplier, session.MaxMultiplier,
			session.TotalDraws, session.CompletedDraws, session.BestMultiplier, session.InviteBonusMultiplier,
			session.GoldenWindowExtraDraws, session.CreatedAt,
		)
		if err != nil {
			return 0, fmt.Errorf("insert lucky wheel session: %w", err)
		}
		defer rows.Close()
		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return 0, fmt.Errorf("scan lucky wheel session id: %w", err)
			}
			return id, nil
		}
		return 0, fmt.Errorf("insert lucky wheel session: no id returned")
	}
	result, err := execCtx.ExecContext(ctx, `
INSERT INTO lucky_wheel_sessions (
  user_id, source_order_id, source_order_type, source_pay_amount,
  matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
  total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
  golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NULL, NULL, ?, ?)`,
		session.UserID, session.SourceOrderID, session.SourceOrderType, session.SourcePayAmount,
		session.MatchedTierID, session.MatchedTierName, session.MinMultiplier, session.MaxMultiplier,
		session.TotalDraws, session.CompletedDraws, session.BestMultiplier, session.InviteBonusMultiplier,
		session.GoldenWindowExtraDraws, session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert lucky wheel session: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read lucky wheel session id: %w", err)
	}
	return id, nil
}

func luckyWheelInsertDrawRecord(ctx context.Context, client *dbent.Client, record *LuckyWheelDrawRecord) (int64, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	d := client.Driver().Dialect()
	if d == dialect.Postgres {
		rows, err := execCtx.QueryContext(ctx, `
INSERT INTO lucky_wheel_draw_records (
  session_id, user_id, draw_index, base_multiplier,
  invite_bonus_multiplier, final_multiplier, is_best, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`,
			record.SessionID, record.UserID, record.DrawIndex, record.BaseMultiplier,
			record.InviteBonusMultiplier, record.FinalMultiplier, record.IsBest, record.CreatedAt,
		)
		if err != nil {
			return 0, fmt.Errorf("insert lucky wheel draw record: %w", err)
		}
		defer rows.Close()
		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return 0, fmt.Errorf("scan lucky wheel draw record id: %w", err)
			}
			return id, nil
		}
		return 0, fmt.Errorf("insert lucky wheel draw record: no id returned")
	}
	result, err := execCtx.ExecContext(ctx, `
INSERT INTO lucky_wheel_draw_records (
  session_id, user_id, draw_index, base_multiplier,
  invite_bonus_multiplier, final_multiplier, is_best, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		record.SessionID, record.UserID, record.DrawIndex, record.BaseMultiplier,
		record.InviteBonusMultiplier, record.FinalMultiplier, record.IsBest, record.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert lucky wheel draw record: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read lucky wheel draw record id: %w", err)
	}
	return id, nil
}

func luckyWheelUpdateSessionAfterDraw(ctx context.Context, client *dbent.Client, sessionID int64, completedDraws int, bestMultiplier float64, settled bool, settledBonus *float64, updatedAt time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE lucky_wheel_sessions
SET completed_draws = ?, best_multiplier = ?, settled = ?, settled_bonus_amount = ?, settled_at = ?, updated_at = ?
WHERE id = ?`)
	var settledAt any
	var settledBonusValue any
	if settled {
		settledAt = updatedAt
		if settledBonus != nil {
			settledBonusValue = *settledBonus
		}
	}
	if _, err := execCtx.ExecContext(ctx, query, completedDraws, bestMultiplier, settled, settledBonusValue, settledAt, updatedAt, sessionID); err != nil {
		return fmt.Errorf("update lucky wheel session after draw: %w", err)
	}
	return nil
}

func luckyWheelGetSessionByOrderID(ctx context.Context, client *dbent.Client, sourceOrderID int64) (*LuckyWheelSession, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount,
       matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
       total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
       golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
       created_at, updated_at
FROM lucky_wheel_sessions
WHERE source_order_id = ?
LIMIT 1`)
	rows, err := execCtx.QueryContext(ctx, query, sourceOrderID)
	if err != nil {
		return nil, fmt.Errorf("query lucky wheel session by order: %w", err)
	}
	defer rows.Close()
	sessions, err := scanLuckyWheelSessions(rows)
	if err != nil || len(sessions) == 0 {
		return nil, err
	}
	return &sessions[0], nil
}

func luckyWheelLockSession(ctx context.Context, client *dbent.Client, userID, sessionID int64) (*LuckyWheelSession, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount,
       matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
       total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
       golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
       created_at, updated_at
FROM lucky_wheel_sessions
WHERE id = ? AND user_id = ?
LIMIT 1`
	if client.Driver().Dialect() == dialect.Postgres {
		query += " FOR UPDATE"
	}
	query = luckyWheelBindVars(client.Driver().Dialect(), query)
	rows, err := execCtx.QueryContext(ctx, query, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("lock lucky wheel session: %w", err)
	}
	defer rows.Close()
	sessions, err := scanLuckyWheelSessions(rows)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, infraerrors.NotFound("LUCKY_WHEEL_SESSION_NOT_FOUND", "lucky wheel session not found")
	}
	return &sessions[0], nil
}

func luckyWheelLoadSession(ctx context.Context, client *dbent.Client, sessionID, userID int64) (*LuckyWheelSession, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount,
       matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
       total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
       golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
       created_at, updated_at
FROM lucky_wheel_sessions
WHERE id = ? AND user_id = ?
LIMIT 1`)
	rows, err := execCtx.QueryContext(ctx, query, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("load lucky wheel session: %w", err)
	}
	defer rows.Close()
	sessions, err := scanLuckyWheelSessions(rows)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, infraerrors.NotFound("LUCKY_WHEEL_SESSION_NOT_FOUND", "lucky wheel session not found")
	}
	return &sessions[0], nil
}

func luckyWheelListSessionsBySettlement(ctx context.Context, client *dbent.Client, userID int64, settled bool, limit int) ([]LuckyWheelSession, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount,
       matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
       total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
       golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
       created_at, updated_at
FROM lucky_wheel_sessions
WHERE user_id = ? AND settled = ?
ORDER BY created_at DESC, id DESC
LIMIT ?`)
	rows, err := execCtx.QueryContext(ctx, query, userID, settled, limit)
	if err != nil {
		return nil, fmt.Errorf("list lucky wheel sessions: %w", err)
	}
	defer rows.Close()
	return scanLuckyWheelSessions(rows)
}

func luckyWheelListRecentSessions(ctx context.Context, client *dbent.Client, limit int) ([]LuckyWheelSession, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount,
       matched_tier_id, matched_tier_name, min_multiplier, max_multiplier,
       total_draws, completed_draws, best_multiplier, invite_bonus_multiplier,
       golden_window_extra_draws, settled, settled_bonus_amount, settled_at,
       created_at, updated_at
FROM lucky_wheel_sessions
ORDER BY created_at DESC, id DESC
LIMIT ?`)
	rows, err := execCtx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list lucky wheel recent sessions: %w", err)
	}
	defer rows.Close()
	return scanLuckyWheelSessions(rows)
}

func luckyWheelListMultiplierStats(ctx context.Context, client *dbent.Client) ([]LuckyWheelMultiplierStat, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	rows, err := execCtx.QueryContext(ctx, `
SELECT final_multiplier, COUNT(*) AS draw_count
FROM lucky_wheel_draw_records
GROUP BY final_multiplier
ORDER BY final_multiplier ASC`)
	if err != nil {
		return nil, fmt.Errorf("list lucky wheel multiplier stats: %w", err)
	}
	defer rows.Close()
	out := make([]LuckyWheelMultiplierStat, 0)
	for rows.Next() {
		var item LuckyWheelMultiplierStat
		if err := rows.Scan(&item.Multiplier, &item.DrawCount); err != nil {
			return nil, fmt.Errorf("scan lucky wheel multiplier stats: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func luckyWheelAttachDrawRecords(ctx context.Context, client *dbent.Client, sessions []LuckyWheelSession) error {
	if len(sessions) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(sessions))
	indexByID := make(map[int64]int, len(sessions))
	for i := range sessions {
		ids = append(ids, sessions[i].ID)
		indexByID[sessions[i].ID] = i
		sessions[i].RemainingDraws = luckyWheelMaxInt(sessions[i].TotalDraws-sessions[i].CompletedDraws, 0)
	}
	records, err := luckyWheelListDrawRecordsBySessionIDs(ctx, client, ids)
	if err != nil {
		return err
	}
	for _, record := range records {
		index, ok := indexByID[record.SessionID]
		if !ok {
			continue
		}
		sessions[index].DrawRecords = append(sessions[index].DrawRecords, record)
	}
	return nil
}

func luckyWheelListDrawRecordsBySessionIDs(ctx context.Context, client *dbent.Client, sessionIDs []int64) ([]LuckyWheelDrawRecord, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	d := client.Driver().Dialect()
	placeholders := make([]string, 0, len(sessionIDs))
	args := make([]any, 0, len(sessionIDs))
	for i, sessionID := range sessionIDs {
		if d == dialect.Postgres {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		} else {
			placeholders = append(placeholders, "?")
		}
		args = append(args, sessionID)
	}
	query := fmt.Sprintf(`
SELECT id, session_id, user_id, draw_index, base_multiplier,
       invite_bonus_multiplier, final_multiplier, is_best, created_at
FROM lucky_wheel_draw_records
WHERE session_id IN (%s)
ORDER BY draw_index ASC, id ASC`, strings.Join(placeholders, ","))
	rows, err := execCtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list lucky wheel draw records: %w", err)
	}
	defer rows.Close()
	out := make([]LuckyWheelDrawRecord, 0)
	for rows.Next() {
		var record LuckyWheelDrawRecord
		if err := rows.Scan(
			&record.ID,
			&record.SessionID,
			&record.UserID,
			&record.DrawIndex,
			&record.BaseMultiplier,
			&record.InviteBonusMultiplier,
			&record.FinalMultiplier,
			&record.IsBest,
			&record.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lucky wheel draw record: %w", err)
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func scanLuckyWheelSessions(rows *sql.Rows) ([]LuckyWheelSession, error) {
	out := make([]LuckyWheelSession, 0)
	for rows.Next() {
		var session LuckyWheelSession
		var settledBonus sql.NullFloat64
		var settledAt sql.NullTime
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SourceOrderID,
			&session.SourceOrderType,
			&session.SourcePayAmount,
			&session.MatchedTierID,
			&session.MatchedTierName,
			&session.MinMultiplier,
			&session.MaxMultiplier,
			&session.TotalDraws,
			&session.CompletedDraws,
			&session.BestMultiplier,
			&session.InviteBonusMultiplier,
			&session.GoldenWindowExtraDraws,
			&session.Settled,
			&settledBonus,
			&settledAt,
			&session.CreatedAt,
			&session.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lucky wheel session: %w", err)
		}
		if settledBonus.Valid {
			value := settledBonus.Float64
			session.SettledBonusAmount = &value
		}
		if settledAt.Valid {
			value := settledAt.Time
			session.SettledAt = &value
		}
		session.RemainingDraws = luckyWheelMaxInt(session.TotalDraws-session.CompletedDraws, 0)
		out = append(out, session)
	}
	return out, rows.Err()
}

func luckyWheelDDLForDialect(d string) []string {
	timestampType := "TIMESTAMP"
	idType := "INTEGER PRIMARY KEY AUTOINCREMENT"
	boolType := "BOOLEAN"
	if d == dialect.Postgres {
		timestampType = "TIMESTAMPTZ"
		idType = "BIGSERIAL PRIMARY KEY"
		boolType = "BOOLEAN"
	}
	return []string{
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS lucky_wheel_sessions (
  id %s,
  user_id BIGINT NOT NULL,
  source_order_id BIGINT NOT NULL UNIQUE,
  source_order_type VARCHAR(20) NOT NULL,
  source_pay_amount DECIMAL(20,2) NOT NULL,
  matched_tier_id VARCHAR(64) NOT NULL,
  matched_tier_name VARCHAR(128) NOT NULL,
  min_multiplier DECIMAL(10,4) NOT NULL,
  max_multiplier DECIMAL(10,4) NOT NULL,
  total_draws INTEGER NOT NULL,
  completed_draws INTEGER NOT NULL DEFAULT 0,
  best_multiplier DECIMAL(10,4) NOT NULL DEFAULT 0,
  invite_bonus_multiplier DECIMAL(10,4) NOT NULL DEFAULT 0,
  golden_window_extra_draws INTEGER NOT NULL DEFAULT 0,
  settled %s NOT NULL DEFAULT FALSE,
  settled_bonus_amount DECIMAL(20,2) NULL,
  settled_at %s NULL,
  created_at %s NOT NULL,
  updated_at %s NOT NULL
)`, idType, boolType, timestampType, timestampType, timestampType),
		`CREATE INDEX IF NOT EXISTS idx_lucky_wheel_sessions_user_settled_created ON lucky_wheel_sessions(user_id, settled, created_at DESC)`,
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS lucky_wheel_draw_records (
  id %s,
  session_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  draw_index INTEGER NOT NULL,
  base_multiplier DECIMAL(10,4) NOT NULL,
  invite_bonus_multiplier DECIMAL(10,4) NOT NULL DEFAULT 0,
  final_multiplier DECIMAL(10,4) NOT NULL,
  is_best %s NOT NULL DEFAULT FALSE,
  created_at %s NOT NULL
)`, idType, boolType, timestampType),
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_lucky_wheel_draw_records_session_draw_index ON lucky_wheel_draw_records(session_id, draw_index)`,
		`CREATE INDEX IF NOT EXISTS idx_lucky_wheel_draw_records_session_created ON lucky_wheel_draw_records(session_id, created_at ASC)`,
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS lucky_wheel_invite_bonus_events (
  id %s,
  inviter_user_id BIGINT NOT NULL,
  invitee_user_id BIGINT NOT NULL UNIQUE,
  source_order_id BIGINT NOT NULL UNIQUE,
  bonus_multiplier DECIMAL(10,4) NOT NULL,
  consumed_session_id BIGINT NULL,
  created_at %s NOT NULL,
  consumed_at %s NULL
)`, idType, timestampType, timestampType),
		`CREATE INDEX IF NOT EXISTS idx_lucky_wheel_invite_bonus_pending ON lucky_wheel_invite_bonus_events(inviter_user_id, consumed_session_id)`,
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS lucky_wheel_golden_window_claims (
  id %s,
  claim_day VARCHAR(20) NOT NULL,
  slot_number INTEGER NOT NULL,
  source_order_id BIGINT NOT NULL UNIQUE,
  user_id BIGINT NOT NULL,
  created_at %s NOT NULL
)`, idType, timestampType),
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_lucky_wheel_golden_window_day_slot ON lucky_wheel_golden_window_claims(claim_day, slot_number)`,
		`CREATE INDEX IF NOT EXISTS idx_lucky_wheel_golden_window_day_created ON lucky_wheel_golden_window_claims(claim_day, created_at ASC)`,
		`CREATE TABLE IF NOT EXISTS user_affiliates (
  user_id BIGINT PRIMARY KEY,
  inviter_id BIGINT NULL,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL
)`,
	}
}

func luckyWheelExecContext(ctx context.Context, client *dbent.Client) (interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, error) {
	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return existingTx, nil
	}
	if client == nil {
		return nil, fmt.Errorf("ent client is nil")
	}
	return client, nil
}

func luckyWheelBindVars(d, query string) string {
	if d != dialect.Postgres {
		return query
	}
	var out strings.Builder
	index := 1
	for _, r := range query {
		if r == '?' {
			out.WriteString(fmt.Sprintf("$%d", index))
			index++
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func isFinitePositive(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0
}

func strconvFormatBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func luckyWheelMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
