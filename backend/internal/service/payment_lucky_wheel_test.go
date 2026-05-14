//go:build unit

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func newLuckyWheelTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:lucky_wheel?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newLuckyWheelPaymentService(t *testing.T) (*PaymentService, *settingPublicRepoStub, *mockUserRepo) {
	t.Helper()

	client := newLuckyWheelTestClient(t)
	settingRepo := &settingPublicRepoStub{
		values: map[string]string{
			SettingPaymentEnabled:       "true",
			SettingKeyLuckyWheelEnabled: "true",
		},
	}
	configService := NewPaymentConfigService(client, settingRepo, nil)
	userRepo := &mockUserRepo{
		getByIDUser: &User{ID: 1, Balance: 0},
	}
	svc := &PaymentService{
		entClient:       client,
		configService:   configService,
		userRepo:        userRepo,
		providersLoaded: true,
	}
	return svc, settingRepo, userRepo
}

func setLuckyWheelTestNow(t *testing.T, value time.Time) {
	t.Helper()
	previous := luckyWheelNow
	luckyWheelNow = func() time.Time { return value }
	t.Cleanup(func() { luckyWheelNow = previous })
}

func setLuckyWheelTestDrawSequence(t *testing.T, values ...float64) {
	t.Helper()
	previous := luckyWheelRandFloat64
	index := 0
	luckyWheelRandFloat64 = func() float64 {
		if index >= len(values) {
			return values[len(values)-1]
		}
		out := values[index]
		index++
		return out
	}
	t.Cleanup(func() { luckyWheelRandFloat64 = previous })
}

func saveLuckyWheelConfig(t *testing.T, repo *settingPublicRepoStub, cfg *LuckyWheelConfig) {
	t.Helper()
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyLuckyWheelConfig] = string(raw)
}

func defaultLuckyWheelConfig() *LuckyWheelConfig {
	return &LuckyWheelConfig{
		EligibleOrderTypes:   []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		MultiplierStep:       0.1,
		GlobalMaxMultiplier:  3.0,
		AmountTiers: []LuckyWheelAmountTier{
			{
				ID:             "tier_20_50",
				Name:           "20-50",
				MinAmount:      20,
				MaxAmount:      luckyWheelPtrFloat64(50),
				MinMultiplier:  1.1,
				MaxMultiplier:  2.0,
				DrawCount:      2,
			},
			{
				ID:             "tier_51_plus",
				Name:           "51+",
				MinAmount:      51,
				MaxAmount:      nil,
				MinMultiplier:  1.2,
				MaxMultiplier:  3.0,
				DrawCount:      3,
			},
		},
		InviteBonus: LuckyWheelInviteBonusConfig{
			Enabled:          true,
			QualifyingAmount: 20,
			BonusPerInvitee:  0.2,
			MaxBonus:         1.0,
			ConsumePolicy:    LuckyWheelInviteBonusConsumeNextSessionOnce,
		},
		GoldenWindow: LuckyWheelGoldenWindowConfig{
			Enabled:       true,
			Timezone:      "Asia/Shanghai",
			StartTime:     "20:00",
			EndTime:       "22:00",
			MinAmount:     51,
			ExtraDraws:    1,
			DailyQuota:    5,
		},
	}
}

func luckyWheelPtrFloat64(v float64) *float64 {
	return &v
}

func TestNormalizeLuckyWheelConfig_RejectsInvalidTierCoverage(t *testing.T) {
	cfg := defaultLuckyWheelConfig()
	cfg.AmountTiers[0].MinAmount = 21

	_, err := normalizeLuckyWheelConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "start")
}

func TestNormalizeLuckyWheelConfig_RejectsInvalidMultiplierStep(t *testing.T) {
	cfg := defaultLuckyWheelConfig()
	cfg.MultiplierStep = 0

	_, err := normalizeLuckyWheelConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "step")
}

func TestGrantLuckyWheelSessionForOrder_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newLuckyWheelPaymentService(t)
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC))

	order := &dbent.PaymentOrder{
		ID:        101,
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 88,
	}

	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, order))
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, order))

	rows, err := svc.entClient.QueryContext(ctx, "SELECT COUNT(*) FROM lucky_wheel_sessions WHERE source_order_id = ?", int64(order.ID))
	require.NoError(t, err)
	defer rows.Close()

	var count int
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 1, count)
}

func TestGrantLuckyWheelSessionForOrder_AwardsGoldenWindowExtraDraw(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newLuckyWheelPaymentService(t)
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 12, 30, 0, 0, time.UTC))

	order := &dbent.PaymentOrder{
		ID:        201,
		UserID:    1,
		OrderType: payment.OrderTypeSubscription,
		PayAmount: 88,
	}

	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, order))

	rows, err := svc.entClient.QueryContext(ctx, `
SELECT total_draws, completed_draws, invite_bonus_multiplier, golden_window_extra_draws
FROM lucky_wheel_sessions
WHERE source_order_id = ?`, int64(order.ID))
	require.NoError(t, err)
	defer rows.Close()

	var totalDraws, completedDraws int
	var inviteBonus, goldenExtra float64
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&totalDraws, &completedDraws, &inviteBonus, &goldenExtra))
	require.Equal(t, 4, totalDraws)
	require.Equal(t, 0, completedDraws)
	require.InDelta(t, 0.0, inviteBonus, 1e-9)
	require.InDelta(t, 1.0, goldenExtra, 1e-9)
}

func TestDrawLuckyWheel_SettlesOnLastDrawAndCreditsOnlyBonus(t *testing.T) {
	ctx := context.Background()
	svc, repo, userRepo := newLuckyWheelPaymentService(t)
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC))
	setLuckyWheelTestDrawSequence(t, 0.0, 0.999)

	userRepo.getByIDUser = &User{ID: 1, Balance: 100}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, int64(1), id)
		userRepo.getByIDUser.Balance += amount
		return nil
	}

	order := &dbent.PaymentOrder{
		ID:        301,
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 50,
	}

	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, order))

	summary, err := svc.GetLuckyWheelSummary(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)
	require.Equal(t, 2, summary.ActiveSession.TotalDraws)

	first, err := svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	require.False(t, first.Settled)
	require.InDelta(t, 1.1, first.DrawRecord.FinalMultiplier, 1e-9)
	require.InDelta(t, 100.0, userRepo.getByIDUser.Balance, 1e-9)

	second, err := svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	require.True(t, second.Settled)
	require.NotNil(t, second.SettledBonusAmount)
	require.InDelta(t, 50.0, *second.SettledBonusAmount, 1e-9)
	require.InDelta(t, 150.0, userRepo.getByIDUser.Balance, 1e-9)
	require.InDelta(t, 2.0, second.BestMultiplier, 1e-9)
	require.Equal(t, 0, second.RemainingDraws)
}

func TestDrawLuckyWheel_CapsFinalMultiplierByAdminMax(t *testing.T) {
	ctx := context.Background()
	svc, repo, userRepo := newLuckyWheelPaymentService(t)
	cfg := defaultLuckyWheelConfig()
	cfg.AmountTiers[1].DrawCount = 2
	cfg.AmountTiers[1].MaxMultiplier = 2.0
	cfg.GlobalMaxMultiplier = 2.0
	saveLuckyWheelConfig(t, repo, cfg)
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC))
	setLuckyWheelTestDrawSequence(t, 0.999)

	userRepo.getByIDUser = &User{ID: 1, Balance: 100}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, int64(1), id)
		userRepo.getByIDUser.Balance += amount
		return nil
	}

	require.NoError(t, svc.ensureLuckyWheelTables(ctx))
	_, err := svc.entClient.ExecContext(ctx, `
INSERT INTO user_affiliates (user_id, inviter_id, created_at, updated_at)
VALUES (?, ?, ?, ?)`, int64(9), int64(1), time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	inviteeOrder := &dbent.PaymentOrder{
		ID:        601,
		UserID:    9,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 20,
	}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, inviteeOrder))

	order := &dbent.PaymentOrder{
		ID:        602,
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 88,
	}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, order))

	summary, err := svc.GetLuckyWheelSummary(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)

	// first draw hits 1.2 base + 0.2 invite = 1.4, second draw can go higher but must cap at 2.0
	_, err = svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	result, err := svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	require.True(t, result.Settled)
	require.NotNil(t, result.SettledBonusAmount)
	require.InDelta(t, 2.0, result.BestMultiplier, 1e-9)
	require.InDelta(t, 188.0, userRepo.getByIDUser.Balance, 1e-9)
}

func TestGrantLuckyWheelSessionForOrder_ConsumesInviteBonusOnNextSession(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newLuckyWheelPaymentService(t)
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC))

	require.NoError(t, svc.ensureLuckyWheelTables(ctx))
	_, err := svc.entClient.ExecContext(ctx, `
INSERT INTO user_affiliates (user_id, inviter_id, created_at, updated_at)
VALUES (?, ?, ?, ?)`, int64(9), int64(1), time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	inviteeOrder := &dbent.PaymentOrder{
		ID:        401,
		UserID:    9,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 20,
	}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, inviteeOrder))

	inviterOrder := &dbent.PaymentOrder{
		ID:        402,
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 88,
	}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, inviterOrder))

	summary, err := svc.GetLuckyWheelSummary(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)
	require.InDelta(t, 0.2, summary.ActiveSession.InviteBonusMultiplier, 1e-9)

	nextOrder := &dbent.PaymentOrder{
		ID:        403,
		UserID:    1,
		OrderType: payment.OrderTypeBalance,
		PayAmount: 88,
	}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, nextOrder))

	rows, err := svc.entClient.QueryContext(ctx, `
SELECT invite_bonus_multiplier
FROM lucky_wheel_sessions
WHERE source_order_id = ?
`, int64(nextOrder.ID))
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())

	var bonus float64
	require.NoError(t, rows.Scan(&bonus))
	require.InDelta(t, 0.0, bonus, 1e-9)
}

func TestGetLuckyWheelSummary_ReturnsPendingAndHistorySessions(t *testing.T) {
	ctx := context.Background()
	svc, repo, userRepo := newLuckyWheelPaymentService(t)
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())
	setLuckyWheelTestNow(t, time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC))
	setLuckyWheelTestDrawSequence(t, 0.0, 0.999)

	userRepo.updateBalanceFn = func(context.Context, int64, float64) error { return nil }

	firstOrder := &dbent.PaymentOrder{ID: 501, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 50}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, firstOrder))

	summary, err := svc.GetLuckyWheelSummary(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)
	_, err = svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	_, err = svc.DrawLuckyWheel(ctx, 1, summary.ActiveSession.ID)
	require.NoError(t, err)

	secondOrder := &dbent.PaymentOrder{ID: 502, UserID: 1, OrderType: payment.OrderTypeSubscription, PayAmount: 88}
	require.NoError(t, svc.GrantLuckyWheelChanceForOrder(ctx, secondOrder))

	updated, err := svc.GetLuckyWheelSummary(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, updated.ActiveSession)
	require.EqualValues(t, secondOrder.ID, updated.ActiveSession.SourceOrderID)
	require.Len(t, updated.PendingSessions, 1)
	require.Len(t, updated.HistorySessions, 1)
	require.True(t, updated.HistorySessions[0].Settled)
}
