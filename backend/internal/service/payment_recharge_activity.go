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
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
)

const SettingKeyRechargeActivityConfig = "recharge_activity_config"

const (
	RechargeActivityFulfillmentPending   = "pending"
	RechargeActivityFulfillmentFulfilled = "fulfilled"
)

var (
	rechargeActivityNow         = func() time.Time { return time.Now().UTC() }
	rechargeActivityRandFloat64 = rand.Float64
)

type RechargeActivityConfig struct {
	EligibleOrderTypes []string                `json:"eligible_order_types"`
	IntroText          string                  `json:"intro_text"`
	RulesTitle         string                  `json:"rules_title"`
	RulesItems         []string                `json:"rules_items"`
	Prizes             []RechargeActivityPrize `json:"prizes"`
}

type RechargeActivityPrize struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	RewardAmount      float64 `json:"reward_amount"`
	RewardDescription string  `json:"reward_description"`
	Probability       float64 `json:"probability"`
	MinPayAmount      float64 `json:"min_pay_amount"`
	Enabled           bool    `json:"enabled"`
	SortOrder         int     `json:"sort_order"`
}

type RechargeActivityChance struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	SourceOrderID   int64      `json:"source_order_id"`
	SourceOrderType string     `json:"source_order_type"`
	SourcePayAmount float64    `json:"source_pay_amount"`
	Drawn           bool       `json:"drawn"`
	DrawnAt         *time.Time `json:"drawn_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type RechargeActivityDrawRecord struct {
	ID                int64      `json:"id"`
	ChanceID          int64      `json:"chance_id"`
	UserID            int64      `json:"user_id"`
	UserEmail         string     `json:"user_email,omitempty"`
	UserName          string     `json:"user_name,omitempty"`
	SourceOrderID     int64      `json:"source_order_id"`
	PrizeID           string     `json:"prize_id"`
	PrizeName         string     `json:"prize_name"`
	RewardAmount      float64    `json:"reward_amount"`
	RewardDescription string     `json:"reward_description"`
	Probability       float64    `json:"probability"`
	MinPayAmount      float64    `json:"min_pay_amount"`
	PrizeSnapshot     string     `json:"prize_snapshot"`
	EligiblePrizeIDs  []string   `json:"eligible_prize_ids"`
	FulfillmentStatus string     `json:"fulfillment_status"`
	FulfillmentNote   string     `json:"fulfillment_note"`
	FulfilledAt       *time.Time `json:"fulfilled_at,omitempty"`
	FulfilledBy       *int64     `json:"fulfilled_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type RechargeActivitySummary struct {
	Enabled        bool                         `json:"enabled"`
	Config         RechargeActivityConfig       `json:"config"`
	PendingChances []RechargeActivityChance     `json:"pending_chances"`
	HistoryRecords []RechargeActivityDrawRecord `json:"history_records"`
}

type RechargeActivityDrawResult struct {
	ChanceID int64                      `json:"chance_id"`
	Record   RechargeActivityDrawRecord `json:"record"`
	Chance   *RechargeActivityChance    `json:"chance,omitempty"`
}

type RechargeActivityStats struct {
	Enabled               bool                         `json:"enabled"`
	TotalChances          int64                        `json:"total_chances"`
	PendingChances        int64                        `json:"pending_chances"`
	DrawnChances          int64                        `json:"drawn_chances"`
	PendingFulfillments   int64                        `json:"pending_fulfillments"`
	FulfilledRecords      int64                        `json:"fulfilled_records"`
	TotalRewardAmount     float64                      `json:"total_reward_amount"`
	RecentRecords         []RechargeActivityDrawRecord `json:"recent_records"`
	RecentRecordsTotal    int64                        `json:"recent_records_total"`
	RecentRecordsPage     int                          `json:"recent_records_page"`
	RecentRecordsPageSize int                          `json:"recent_records_page_size"`
	RecentRecordsKeyword  string                       `json:"recent_records_keyword"`
}

type RechargeActivityStatsQuery struct {
	Page        int
	PageSize    int
	UserKeyword string
}

func setRechargeActivityTestDrawSequence(t interface{ Cleanup(func()) }, values ...float64) {
	previous := rechargeActivityRandFloat64
	index := 0
	rechargeActivityRandFloat64 = func() float64 {
		if len(values) == 0 {
			return 0
		}
		if index >= len(values) {
			return values[len(values)-1]
		}
		out := values[index]
		index++
		return out
	}
	t.Cleanup(func() { rechargeActivityRandFloat64 = previous })
}

func (s *PaymentService) rechargeActivityEnabled(ctx context.Context) bool {
	if s == nil || s.configService == nil {
		return false
	}
	if !s.configService.IsPaymentEnabled(ctx) {
		return false
	}
	value, err := s.configService.settingRepo.GetValue(ctx, SettingKeyRechargeActivityEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *PaymentService) ensureRechargeActivityTables(ctx context.Context) error {
	if s == nil || s.entClient == nil {
		return fmt.Errorf("payment service is not initialized")
	}
	execCtx, err := luckyWheelExecContext(ctx, s.entClient)
	if err != nil {
		return err
	}
	for _, statement := range rechargeActivityDDLForDialect(s.entClient.Driver().Dialect()) {
		if _, err := execCtx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure recharge activity tables: %w", err)
		}
	}
	return nil
}

func (s *PaymentService) GetRechargeActivityConfig(ctx context.Context) (*RechargeActivityConfig, bool, error) {
	enabled := s.rechargeActivityEnabled(ctx)
	raw, err := s.configService.settingRepo.GetValue(ctx, SettingKeyRechargeActivityConfig)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, false, fmt.Errorf("get recharge activity config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg, normErr := normalizeRechargeActivityConfig(nil)
		return cfg, enabled, normErr
	}
	var cfg RechargeActivityConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, enabled, infraerrors.InternalServer("RECHARGE_ACTIVITY_CONFIG_INVALID", "recharge activity config is invalid")
	}
	normalized, normErr := normalizeRechargeActivityConfig(&cfg)
	if normErr != nil {
		fallback, fallbackErr := normalizeRechargeActivityConfig(nil)
		return fallback, enabled, fallbackErr
	}
	return normalized, enabled, nil
}

func (s *PaymentService) UpdateRechargeActivityConfig(ctx context.Context, enabled bool, cfg *RechargeActivityConfig) (*RechargeActivityConfig, error) {
	normalized, err := normalizeRechargeActivityConfig(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal recharge activity config: %w", err)
	}
	updates := map[string]string{
		SettingKeyRechargeActivityEnabled: strconvFormatBool(enabled),
		SettingKeyRechargeActivityConfig:  string(raw),
	}
	if err := s.configService.settingRepo.SetMultiple(ctx, updates); err != nil {
		return nil, fmt.Errorf("save recharge activity config: %w", err)
	}
	return normalized, nil
}

func (s *PaymentService) GrantRechargeActivityChanceForOrder(ctx context.Context, order *dbent.PaymentOrder) error {
	if order == nil || !s.rechargeActivityEnabled(ctx) {
		return nil
	}
	if err := s.ensureRechargeActivityTables(ctx); err != nil {
		return err
	}
	cfg, _, err := s.GetRechargeActivityConfig(ctx)
	if err != nil {
		return err
	}
	if !rechargeActivityOrderTypeEligible(cfg, order.OrderType) {
		return nil
	}
	if len(rechargeActivityEligiblePrizes(cfg.Prizes, order.PayAmount)) == 0 {
		return nil
	}
	now := rechargeActivityNow()
	_, err = rechargeActivityInsertChance(ctx, s.entClient, &RechargeActivityChance{
		UserID:          order.UserID,
		SourceOrderID:   int64(order.ID),
		SourceOrderType: order.OrderType,
		SourcePayAmount: order.PayAmount,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	return err
}

func (s *PaymentService) GetRechargeActivitySummary(ctx context.Context, userID int64) (*RechargeActivitySummary, error) {
	if err := s.ensureRechargeActivityTables(ctx); err != nil {
		return nil, err
	}
	cfg, enabled, err := s.GetRechargeActivityConfig(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := rechargeActivityListPendingChances(ctx, s.entClient, userID, 20)
	if err != nil {
		return nil, err
	}
	history, err := rechargeActivityListUserRecords(ctx, s.entClient, userID, 20)
	if err != nil {
		return nil, err
	}
	return &RechargeActivitySummary{
		Enabled:        enabled,
		Config:         *cfg,
		PendingChances: pending,
		HistoryRecords: history,
	}, nil
}

func (s *PaymentService) DrawRechargeActivity(ctx context.Context, userID int64, chanceID int64) (*RechargeActivityDrawResult, error) {
	if !s.rechargeActivityEnabled(ctx) {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_DISABLED", "recharge activity is disabled")
	}
	if err := s.ensureRechargeActivityTables(ctx); err != nil {
		return nil, err
	}
	cfg, _, err := s.GetRechargeActivityConfig(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin recharge activity draw tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	chance, err := rechargeActivityLoadChance(txCtx, tx.Client(), chanceID, userID)
	if err != nil {
		return nil, err
	}
	if chance.Drawn {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_CHANCE_DRAWN", "recharge activity chance already drawn")
	}
	prize, err := drawRechargeActivityPrize(cfg.Prizes, chance.SourcePayAmount)
	if err != nil {
		return nil, err
	}
	eligibleIDs := rechargeActivityPrizeIDs(rechargeActivityEligiblePrizes(cfg.Prizes, chance.SourcePayAmount))
	snapshot, _ := json.Marshal(prize)
	now := rechargeActivityNow()
	record := &RechargeActivityDrawRecord{
		ChanceID:          chance.ID,
		UserID:            userID,
		SourceOrderID:     chance.SourceOrderID,
		PrizeID:           prize.ID,
		PrizeName:         prize.Name,
		RewardAmount:      0,
		RewardDescription: prize.RewardDescription,
		Probability:       prize.Probability,
		MinPayAmount:      prize.MinPayAmount,
		PrizeSnapshot:     string(snapshot),
		EligiblePrizeIDs:  eligibleIDs,
		FulfillmentStatus: RechargeActivityFulfillmentPending,
		FulfillmentNote:   "",
		CreatedAt:         now,
	}
	recordID, err := rechargeActivityInsertDrawRecord(txCtx, tx.Client(), record)
	if err != nil {
		return nil, err
	}
	record.ID = recordID
	if err := rechargeActivityMarkChanceDrawn(txCtx, tx.Client(), chance.ID, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit recharge activity draw tx: %w", err)
	}
	updated, err := rechargeActivityLoadChance(ctx, s.entClient, chanceID, userID)
	if err != nil {
		return nil, err
	}
	return &RechargeActivityDrawResult{ChanceID: chance.ID, Record: *record, Chance: updated}, nil
}

func (s *PaymentService) GetRechargeActivityStats(ctx context.Context, query RechargeActivityStatsQuery) (*RechargeActivityStats, error) {
	if err := s.ensureRechargeActivityTables(ctx); err != nil {
		return nil, err
	}
	enabled := s.rechargeActivityEnabled(ctx)
	stats, err := rechargeActivityLoadStats(ctx, s.entClient, query)
	if err != nil {
		return nil, err
	}
	stats.Enabled = enabled
	return stats, nil
}

func (s *PaymentService) UpdateRechargeActivityRecordFulfillment(ctx context.Context, recordID int64, adminID int64, status string, note string) (*RechargeActivityDrawRecord, error) {
	if err := s.ensureRechargeActivityTables(ctx); err != nil {
		return nil, err
	}
	status = strings.TrimSpace(status)
	if status != RechargeActivityFulfillmentPending && status != RechargeActivityFulfillmentFulfilled {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_FULFILLMENT_STATUS_INVALID", "fulfillment status is invalid")
	}
	note = strings.TrimSpace(note)
	if status == RechargeActivityFulfillmentFulfilled && note == "" {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_FULFILLMENT_NOTE_REQUIRED", "fulfillment note is required")
	}
	now := rechargeActivityNow()
	if err := rechargeActivityUpdateRecordFulfillment(ctx, s.entClient, recordID, adminID, status, note, now); err != nil {
		return nil, err
	}
	return rechargeActivityLoadRecord(ctx, s.entClient, recordID)
}

func normalizeRechargeActivityConfig(cfg *RechargeActivityConfig) (*RechargeActivityConfig, error) {
	if rechargeActivityConfigLooksEmpty(cfg) {
		return cloneDefaultRechargeActivityConfig(), nil
	}
	normalized := &RechargeActivityConfig{
		EligibleOrderTypes: normalizeLuckyWheelOrderTypes(cfg.EligibleOrderTypes),
		IntroText:          strings.TrimSpace(cfg.IntroText),
		RulesTitle:         strings.TrimSpace(cfg.RulesTitle),
	}
	for _, item := range cfg.RulesItems {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			normalized.RulesItems = append(normalized.RulesItems, trimmed)
		}
	}
	if normalized.IntroText == "" {
		normalized.IntroText = defaultRechargeActivityIntroText
	}
	if normalized.RulesTitle == "" {
		normalized.RulesTitle = defaultRechargeActivityRulesTitle
	}
	if len(normalized.RulesItems) == 0 {
		normalized.RulesItems = append(normalized.RulesItems, defaultRechargeActivityRulesItems...)
	}
	if len(normalized.EligibleOrderTypes) == 0 {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_ORDER_TYPES_REQUIRED", "at least one eligible order type is required")
	}
	seen := map[string]struct{}{}
	totalEnabledProbability := 0.0
	for _, prize := range cfg.Prizes {
		prize.ID = strings.TrimSpace(prize.ID)
		prize.Name = strings.TrimSpace(prize.Name)
		if prize.ID == "" {
			return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_PRIZE_ID_REQUIRED", "prize id is required")
		}
		if _, ok := seen[prize.ID]; ok {
			return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_PRIZE_DUPLICATE", "prize ids must be unique")
		}
		seen[prize.ID] = struct{}{}
		if prize.Name == "" {
			prize.Name = prize.ID
		}
		prize.RewardDescription = strings.TrimSpace(prize.RewardDescription)
		prize.RewardAmount = 0
		if !isFinitePositive(prize.Probability) {
			return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_PROBABILITY_INVALID", "prize probability must be greater than 0")
		}
		if prize.MinPayAmount < 0 || math.IsNaN(prize.MinPayAmount) || math.IsInf(prize.MinPayAmount, 0) {
			return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_MIN_AMOUNT_INVALID", "min pay amount must be >= 0")
		}
		if prize.Enabled {
			totalEnabledProbability += prize.Probability
		}
		normalized.Prizes = append(normalized.Prizes, prize)
	}
	if len(normalized.Prizes) == 0 {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_PRIZES_REQUIRED", "at least one prize is required")
	}
	if totalEnabledProbability <= 0 {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_ENABLED_PRIZES_REQUIRED", "at least one enabled prize is required")
	}
	if math.Abs(totalEnabledProbability-100) > 0.000001 {
		return nil, infraerrors.BadRequest("RECHARGE_ACTIVITY_PROBABILITY_TOTAL_INVALID", "enabled prize probability total must equal 100")
	}
	sort.SliceStable(normalized.Prizes, func(i, j int) bool {
		if normalized.Prizes[i].SortOrder == normalized.Prizes[j].SortOrder {
			return normalized.Prizes[i].ID < normalized.Prizes[j].ID
		}
		return normalized.Prizes[i].SortOrder < normalized.Prizes[j].SortOrder
	})
	return normalized, nil
}

func rechargeActivityConfigLooksEmpty(cfg *RechargeActivityConfig) bool {
	return cfg == nil || (len(cfg.Prizes) == 0 && len(cfg.EligibleOrderTypes) == 0)
}

func cloneDefaultRechargeActivityConfig() *RechargeActivityConfig {
	return &RechargeActivityConfig{
		EligibleOrderTypes: []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		IntroText:          defaultRechargeActivityIntroText,
		RulesTitle:         defaultRechargeActivityRulesTitle,
		RulesItems:         append([]string(nil), defaultRechargeActivityRulesItems...),
		Prizes: []RechargeActivityPrize{
			{ID: "third", Name: "三等奖", RewardDescription: "中奖后请联系客服领取三等奖。", Probability: 70, MinPayAmount: 20, Enabled: true, SortOrder: 3},
			{ID: "second", Name: "二等奖", RewardDescription: "中奖后由管理员人工发放二等奖权益。", Probability: 20, MinPayAmount: 50, Enabled: true, SortOrder: 2},
			{ID: "first", Name: "一等奖", RewardDescription: "中奖后由管理员联系确认并发放一等奖。", Probability: 10, MinPayAmount: 100, Enabled: true, SortOrder: 1},
		},
	}
}

const defaultRechargeActivityIntroText = "充值活动按订单实付人民币发放抽奖机会，达标奖品按管理员配置概率抽取，中奖后由管理员人工发放奖励。"
const defaultRechargeActivityRulesTitle = "活动规则"

var defaultRechargeActivityRulesItems = []string{
	"余额充值和订阅购买订单完成后，各获得 1 次抽奖机会。",
	"奖品按最低实付金额过滤；过滤后的可中奖品按相对概率自动重算。",
	"中奖结果会进入待发放记录，由管理员按奖励说明人工处理。",
}

func rechargeActivityOrderTypeEligible(cfg *RechargeActivityConfig, orderType string) bool {
	for _, candidate := range cfg.EligibleOrderTypes {
		if candidate == orderType {
			return true
		}
	}
	return false
}

func rechargeActivityEligiblePrizes(prizes []RechargeActivityPrize, payAmount float64) []RechargeActivityPrize {
	out := make([]RechargeActivityPrize, 0, len(prizes))
	for _, prize := range prizes {
		if !prize.Enabled || payAmount < prize.MinPayAmount {
			continue
		}
		out = append(out, prize)
	}
	return out
}

func drawRechargeActivityPrize(prizes []RechargeActivityPrize, payAmount float64) (RechargeActivityPrize, error) {
	eligible := rechargeActivityEligiblePrizes(prizes, payAmount)
	if len(eligible) == 0 {
		return RechargeActivityPrize{}, infraerrors.BadRequest("RECHARGE_ACTIVITY_NO_ELIGIBLE_PRIZE", "no eligible prize for this payment amount")
	}
	total := 0.0
	for _, prize := range eligible {
		total += prize.Probability
	}
	if total <= 0 {
		return RechargeActivityPrize{}, infraerrors.BadRequest("RECHARGE_ACTIVITY_NO_ELIGIBLE_PRIZE", "no eligible prize for this payment amount")
	}
	target := rechargeActivityRandFloat64() * total
	acc := 0.0
	for _, prize := range eligible {
		acc += prize.Probability
		if target < acc {
			return prize, nil
		}
	}
	return eligible[len(eligible)-1], nil
}

func rechargeActivityPrizeIDs(prizes []RechargeActivityPrize) []string {
	out := make([]string, 0, len(prizes))
	for _, prize := range prizes {
		out = append(out, prize.ID)
	}
	return out
}

func rechargeActivityInsertChance(ctx context.Context, client *dbent.Client, chance *RechargeActivityChance) (int64, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	dialectName := client.Driver().Dialect()
	if dialectName == dialect.Postgres {
		rows, err := execCtx.QueryContext(ctx, `
INSERT INTO recharge_activity_chances (
  user_id, source_order_id, source_order_type, source_pay_amount, drawn, drawn_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, FALSE, NULL, $5, $6)
ON CONFLICT (source_order_id) DO UPDATE SET source_order_id = EXCLUDED.source_order_id
RETURNING id`, chance.UserID, chance.SourceOrderID, chance.SourceOrderType, chance.SourcePayAmount, chance.CreatedAt, chance.UpdatedAt)
		if err != nil {
			return 0, fmt.Errorf("insert recharge activity chance: %w", err)
		}
		defer rows.Close()
		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return 0, fmt.Errorf("scan recharge activity chance id: %w", err)
			}
			return id, nil
		}
		return 0, fmt.Errorf("insert recharge activity chance: no id returned")
	}
	res, err := execCtx.ExecContext(ctx, `
INSERT OR IGNORE INTO recharge_activity_chances (
  user_id, source_order_id, source_order_type, source_pay_amount, drawn, drawn_at, created_at, updated_at
) VALUES (?, ?, ?, ?, FALSE, NULL, ?, ?)`, chance.UserID, chance.SourceOrderID, chance.SourceOrderType, chance.SourcePayAmount, chance.CreatedAt, chance.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert recharge activity chance: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func rechargeActivityLoadChance(ctx context.Context, client *dbent.Client, chanceID int64, userID int64) (*RechargeActivityChance, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount, drawn, drawn_at, created_at, updated_at
FROM recharge_activity_chances
WHERE id = ? AND user_id = ?`)
	rows, err := execCtx.QueryContext(ctx, query, chanceID, userID)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity chance: %w", err)
	}
	defer rows.Close()
	chances, err := scanRechargeActivityChances(rows)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity chance: %w", err)
	}
	if len(chances) == 0 {
		return nil, infraerrors.NotFound("RECHARGE_ACTIVITY_CHANCE_NOT_FOUND", "recharge activity chance not found")
	}
	return &chances[0], nil
}

func rechargeActivityListPendingChances(ctx context.Context, client *dbent.Client, userID int64, limit int) ([]RechargeActivityChance, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, user_id, source_order_id, source_order_type, source_pay_amount, drawn, drawn_at, created_at, updated_at
FROM recharge_activity_chances
WHERE user_id = ? AND drawn = FALSE
ORDER BY created_at ASC, id ASC
LIMIT ?`)
	rows, err := execCtx.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recharge activity pending chances: %w", err)
	}
	defer rows.Close()
	return scanRechargeActivityChances(rows)
}

func rechargeActivityListUserRecords(ctx context.Context, client *dbent.Client, userID int64, limit int) ([]RechargeActivityDrawRecord, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, chance_id, user_id, source_order_id, prize_id, prize_name, reward_amount, reward_description,
       probability, min_pay_amount, prize_snapshot, eligible_prize_ids,
       fulfillment_status, fulfillment_note, fulfilled_at, fulfilled_by, created_at
FROM recharge_activity_draw_records
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?`)
	rows, err := execCtx.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recharge activity records: %w", err)
	}
	defer rows.Close()
	return scanRechargeActivityDrawRecords(rows)
}

func rechargeActivityInsertDrawRecord(ctx context.Context, client *dbent.Client, record *RechargeActivityDrawRecord) (int64, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return 0, err
	}
	eligibleRaw, _ := json.Marshal(record.EligiblePrizeIDs)
	if client.Driver().Dialect() == dialect.Postgres {
		rows, err := execCtx.QueryContext(ctx, `
INSERT INTO recharge_activity_draw_records (
  chance_id, user_id, source_order_id, prize_id, prize_name, reward_amount, reward_description, probability, min_pay_amount,
  prize_snapshot, eligible_prize_ids, fulfillment_status, fulfillment_note, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id`, record.ChanceID, record.UserID, record.SourceOrderID, record.PrizeID, record.PrizeName, record.RewardAmount, record.RewardDescription, record.Probability, record.MinPayAmount, record.PrizeSnapshot, string(eligibleRaw), record.FulfillmentStatus, record.FulfillmentNote, record.CreatedAt)
		if err != nil {
			return 0, fmt.Errorf("insert recharge activity draw record: %w", err)
		}
		defer rows.Close()
		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return 0, fmt.Errorf("scan recharge activity draw record id: %w", err)
			}
			return id, nil
		}
		return 0, fmt.Errorf("insert recharge activity draw record: no id returned")
	}
	res, err := execCtx.ExecContext(ctx, `
INSERT INTO recharge_activity_draw_records (
  chance_id, user_id, source_order_id, prize_id, prize_name, reward_amount, reward_description, probability, min_pay_amount,
  prize_snapshot, eligible_prize_ids, fulfillment_status, fulfillment_note, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, record.ChanceID, record.UserID, record.SourceOrderID, record.PrizeID, record.PrizeName, record.RewardAmount, record.RewardDescription, record.Probability, record.MinPayAmount, record.PrizeSnapshot, string(eligibleRaw), record.FulfillmentStatus, record.FulfillmentNote, record.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert recharge activity draw record: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func rechargeActivityLoadRecord(ctx context.Context, client *dbent.Client, recordID int64) (*RechargeActivityDrawRecord, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT id, chance_id, user_id, source_order_id, prize_id, prize_name, reward_amount, reward_description,
       probability, min_pay_amount, prize_snapshot, eligible_prize_ids,
       fulfillment_status, fulfillment_note, fulfilled_at, fulfilled_by, created_at
FROM recharge_activity_draw_records
WHERE id = ?`)
	rows, err := execCtx.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity record: %w", err)
	}
	defer rows.Close()
	records, err := scanRechargeActivityDrawRecords(rows)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity record: %w", err)
	}
	if len(records) == 0 {
		return nil, infraerrors.NotFound("RECHARGE_ACTIVITY_RECORD_NOT_FOUND", "recharge activity record not found")
	}
	return &records[0], nil
}

func rechargeActivityUpdateRecordFulfillment(ctx context.Context, client *dbent.Client, recordID int64, adminID int64, status string, note string, now time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	var query string
	var args []any
	if status == RechargeActivityFulfillmentFulfilled {
		query = luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE recharge_activity_draw_records
SET fulfillment_status = ?, fulfillment_note = ?, fulfilled_at = ?, fulfilled_by = ?
WHERE id = ?`)
		args = []any{status, note, now, adminID, recordID}
	} else {
		query = luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE recharge_activity_draw_records
SET fulfillment_status = ?, fulfillment_note = ?, fulfilled_at = NULL, fulfilled_by = NULL
WHERE id = ?`)
		args = []any{status, note, recordID}
	}
	res, err := execCtx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update recharge activity fulfillment: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return infraerrors.NotFound("RECHARGE_ACTIVITY_RECORD_NOT_FOUND", "recharge activity record not found")
	}
	return nil
}

func rechargeActivityMarkChanceDrawn(ctx context.Context, client *dbent.Client, chanceID int64, now time.Time) error {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return err
	}
	query := luckyWheelBindVars(client.Driver().Dialect(), `
UPDATE recharge_activity_chances
SET drawn = TRUE, drawn_at = ?, updated_at = ?
WHERE id = ? AND drawn = FALSE`)
	res, err := execCtx.ExecContext(ctx, query, now, now, chanceID)
	if err != nil {
		return fmt.Errorf("mark recharge activity chance drawn: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return infraerrors.BadRequest("RECHARGE_ACTIVITY_CHANCE_DRAWN", "recharge activity chance already drawn")
	}
	return nil
}

func normalizeRechargeActivityStatsQuery(query RechargeActivityStatsQuery) RechargeActivityStatsQuery {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	query.UserKeyword = strings.TrimSpace(query.UserKeyword)
	return query
}

func rechargeActivityLoadStats(ctx context.Context, client *dbent.Client, query RechargeActivityStatsQuery) (*RechargeActivityStats, error) {
	execCtx, err := luckyWheelExecContext(ctx, client)
	if err != nil {
		return nil, err
	}
	query = normalizeRechargeActivityStatsQuery(query)
	stats := &RechargeActivityStats{
		RecentRecordsPage:     query.Page,
		RecentRecordsPageSize: query.PageSize,
		RecentRecordsKeyword:  query.UserKeyword,
	}
	rows, err := execCtx.QueryContext(ctx, `
SELECT COUNT(*), COALESCE(SUM(CASE WHEN drawn = FALSE THEN 1 ELSE 0 END), 0), COALESCE(SUM(CASE WHEN drawn = TRUE THEN 1 ELSE 0 END), 0)
FROM recharge_activity_chances`)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity chance stats: %w", err)
	}
	if rows.Next() {
		if err := rows.Scan(&stats.TotalChances, &stats.PendingChances, &stats.DrawnChances); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan recharge activity chance stats: %w", err)
		}
	}
	rows.Close()
	rows, err = execCtx.QueryContext(ctx, `
SELECT
  COALESCE(SUM(CASE WHEN fulfillment_status = 'pending' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN fulfillment_status = 'fulfilled' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(reward_amount), 0)
FROM recharge_activity_draw_records`)
	if err != nil {
		return nil, fmt.Errorf("load recharge activity fulfillment stats: %w", err)
	}
	if rows.Next() {
		if err := rows.Scan(&stats.PendingFulfillments, &stats.FulfilledRecords, &stats.TotalRewardAmount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan recharge activity fulfillment stats: %w", err)
		}
	}
	rows.Close()

	where, args := rechargeActivityRecentRecordsFilter(query.UserKeyword)
	countQuery := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT COUNT(*)
FROM recharge_activity_draw_records r
LEFT JOIN users u ON u.id = r.user_id
`+where)
	rows, err = execCtx.QueryContext(ctx, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("count recharge activity recent records: %w", err)
	}
	if rows.Next() {
		if err := rows.Scan(&stats.RecentRecordsTotal); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan recharge activity recent records total: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("count recharge activity recent records: %w", err)
	}
	rows.Close()

	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)
	recordsQuery := luckyWheelBindVars(client.Driver().Dialect(), `
SELECT r.id, r.chance_id, r.user_id, r.source_order_id, r.prize_id, r.prize_name, r.reward_amount, r.reward_description,
       r.probability, r.min_pay_amount, r.prize_snapshot, r.eligible_prize_ids,
       r.fulfillment_status, r.fulfillment_note, r.fulfilled_at, r.fulfilled_by, r.created_at,
       u.email, u.username
FROM recharge_activity_draw_records r
LEFT JOIN users u ON u.id = r.user_id
`+where+`
ORDER BY r.created_at DESC, r.id DESC
LIMIT ?
OFFSET ?`)
	rows, err = execCtx.QueryContext(ctx, recordsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list recharge activity recent records: %w", err)
	}
	defer rows.Close()
	stats.RecentRecords, err = scanRechargeActivityDrawRecordsWithUser(rows)
	return stats, err
}

func rechargeActivityRecentRecordsFilter(keyword string) (string, []any) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return "", nil
	}
	where := "WHERE (LOWER(COALESCE(u.email, '')) LIKE ? OR LOWER(COALESCE(u.username, '')) LIKE ?"
	args := []any{"%" + strings.ToLower(keyword) + "%", "%" + strings.ToLower(keyword) + "%"}
	if userID, err := strconv.ParseInt(keyword, 10, 64); err == nil {
		where += " OR r.user_id = ?"
		args = append(args, userID)
	}
	where += ")"
	return where, args
}

type rechargeActivityChanceScanner interface {
	Scan(dest ...any) error
}

func scanRechargeActivityChance(scanner rechargeActivityChanceScanner) (*RechargeActivityChance, error) {
	var chance RechargeActivityChance
	var drawnAt sql.NullTime
	if err := scanner.Scan(
		&chance.ID, &chance.UserID, &chance.SourceOrderID, &chance.SourceOrderType, &chance.SourcePayAmount,
		&chance.Drawn, &drawnAt, &chance.CreatedAt, &chance.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if drawnAt.Valid {
		chance.DrawnAt = &drawnAt.Time
	}
	return &chance, nil
}

func scanRechargeActivityChances(rows *sql.Rows) ([]RechargeActivityChance, error) {
	out := []RechargeActivityChance{}
	for rows.Next() {
		chance, err := scanRechargeActivityChance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *chance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanRechargeActivityDrawRecords(rows *sql.Rows) ([]RechargeActivityDrawRecord, error) {
	out := []RechargeActivityDrawRecord{}
	for rows.Next() {
		var record RechargeActivityDrawRecord
		var eligibleRaw string
		var fulfilledAt sql.NullTime
		var fulfilledBy sql.NullInt64
		if err := rows.Scan(
			&record.ID, &record.ChanceID, &record.UserID, &record.SourceOrderID, &record.PrizeID, &record.PrizeName,
			&record.RewardAmount, &record.RewardDescription, &record.Probability, &record.MinPayAmount, &record.PrizeSnapshot, &eligibleRaw,
			&record.FulfillmentStatus, &record.FulfillmentNote, &fulfilledAt, &fulfilledBy, &record.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eligibleRaw), &record.EligiblePrizeIDs)
		if record.EligiblePrizeIDs == nil {
			record.EligiblePrizeIDs = []string{}
		}
		if record.FulfillmentStatus == "" {
			record.FulfillmentStatus = RechargeActivityFulfillmentPending
		}
		if fulfilledAt.Valid {
			record.FulfilledAt = &fulfilledAt.Time
		}
		if fulfilledBy.Valid {
			record.FulfilledBy = &fulfilledBy.Int64
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanRechargeActivityDrawRecordsWithUser(rows *sql.Rows) ([]RechargeActivityDrawRecord, error) {
	out := []RechargeActivityDrawRecord{}
	for rows.Next() {
		var record RechargeActivityDrawRecord
		var eligibleRaw string
		var fulfilledAt sql.NullTime
		var fulfilledBy sql.NullInt64
		var userEmail sql.NullString
		var userName sql.NullString
		if err := rows.Scan(
			&record.ID, &record.ChanceID, &record.UserID, &record.SourceOrderID, &record.PrizeID, &record.PrizeName,
			&record.RewardAmount, &record.RewardDescription, &record.Probability, &record.MinPayAmount, &record.PrizeSnapshot, &eligibleRaw,
			&record.FulfillmentStatus, &record.FulfillmentNote, &fulfilledAt, &fulfilledBy, &record.CreatedAt,
			&userEmail, &userName,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eligibleRaw), &record.EligiblePrizeIDs)
		if record.EligiblePrizeIDs == nil {
			record.EligiblePrizeIDs = []string{}
		}
		if record.FulfillmentStatus == "" {
			record.FulfillmentStatus = RechargeActivityFulfillmentPending
		}
		if fulfilledAt.Valid {
			record.FulfilledAt = &fulfilledAt.Time
		}
		if fulfilledBy.Valid {
			record.FulfilledBy = &fulfilledBy.Int64
		}
		if userEmail.Valid {
			record.UserEmail = userEmail.String
		}
		if userName.Valid {
			record.UserName = userName.String
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func rechargeActivityDDLForDialect(dialectName string) []string {
	idType := "INTEGER PRIMARY KEY AUTOINCREMENT"
	timestampType := "TIMESTAMP"
	boolType := "BOOLEAN"
	if dialectName == dialect.Postgres {
		idType = "BIGSERIAL PRIMARY KEY"
	}
	statements := []string{
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS recharge_activity_chances (
  id %s,
  user_id BIGINT NOT NULL,
  source_order_id BIGINT NOT NULL UNIQUE,
  source_order_type VARCHAR(20) NOT NULL,
  source_pay_amount DECIMAL(20,2) NOT NULL,
  drawn %s NOT NULL DEFAULT FALSE,
  drawn_at %s NULL,
  created_at %s NOT NULL,
  updated_at %s NOT NULL
)`, idType, boolType, timestampType, timestampType, timestampType),
		`CREATE INDEX IF NOT EXISTS idx_recharge_activity_chances_user_drawn_created ON recharge_activity_chances(user_id, drawn, created_at DESC)`,
		fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS recharge_activity_draw_records (
  id %s,
  chance_id BIGINT NOT NULL UNIQUE,
  user_id BIGINT NOT NULL,
  source_order_id BIGINT NOT NULL,
  prize_id VARCHAR(64) NOT NULL,
  prize_name VARCHAR(128) NOT NULL,
  reward_amount DECIMAL(20,2) NOT NULL,
  reward_description TEXT NOT NULL DEFAULT '',
  probability DECIMAL(10,4) NOT NULL,
  min_pay_amount DECIMAL(20,2) NOT NULL,
  prize_snapshot TEXT NOT NULL,
  eligible_prize_ids TEXT NOT NULL,
  fulfillment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
  fulfillment_note TEXT NOT NULL DEFAULT '',
  fulfilled_at %s NULL,
  fulfilled_by BIGINT NULL,
  created_at %s NOT NULL
)`, idType, timestampType, timestampType),
		`CREATE INDEX IF NOT EXISTS idx_recharge_activity_draw_records_user_created ON recharge_activity_draw_records(user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_recharge_activity_draw_records_order ON recharge_activity_draw_records(source_order_id)`,
	}
	if dialectName == dialect.Postgres {
		statements = append(statements,
			`ALTER TABLE recharge_activity_draw_records ADD COLUMN IF NOT EXISTS reward_description TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE recharge_activity_draw_records ADD COLUMN IF NOT EXISTS fulfillment_status VARCHAR(20) NOT NULL DEFAULT 'pending'`,
			`ALTER TABLE recharge_activity_draw_records ADD COLUMN IF NOT EXISTS fulfillment_note TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE recharge_activity_draw_records ADD COLUMN IF NOT EXISTS fulfilled_at TIMESTAMP NULL`,
			`ALTER TABLE recharge_activity_draw_records ADD COLUMN IF NOT EXISTS fulfilled_by BIGINT NULL`,
		)
	}
	return statements
}
