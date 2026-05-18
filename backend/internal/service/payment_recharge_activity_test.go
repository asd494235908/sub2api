//go:build unit

package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func newRechargeActivityPaymentService(t *testing.T) (*PaymentService, *settingPublicRepoStub, *mockUserRepo) {
	t.Helper()

	client := newLuckyWheelTestClient(t)
	settingRepo := &settingPublicRepoStub{
		values: map[string]string{
			SettingPaymentEnabled:             "true",
			SettingKeyRechargeActivityEnabled: "true",
		},
	}
	configService := NewPaymentConfigService(client, settingRepo, nil)
	userRepo := &mockUserRepo{
		getByIDUser: &User{ID: 1, Balance: 0},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		userRepo.getByIDUser.Balance += amount
		return nil
	}
	svc := &PaymentService{
		entClient:       client,
		configService:   configService,
		userRepo:        userRepo,
		providersLoaded: true,
	}
	return svc, settingRepo, userRepo
}

func defaultRechargeActivityConfig() *RechargeActivityConfig {
	return &RechargeActivityConfig{
		EligibleOrderTypes: []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		Prizes: []RechargeActivityPrize{
			{ID: "third", Name: "三等奖", RewardDescription: "联系客服领取实体礼品", Probability: 70, MinPayAmount: 20, Enabled: true, SortOrder: 3},
			{ID: "second", Name: "二等奖", RewardDescription: "赠送站外会员 30 天", Probability: 20, MinPayAmount: 50, Enabled: true, SortOrder: 2},
			{ID: "first", Name: "一等奖", RewardDescription: "人工发放定制奖励", Probability: 10, MinPayAmount: 100, Enabled: true, SortOrder: 1},
		},
	}
}

func saveRechargeActivityConfig(t *testing.T, repo *settingPublicRepoStub, cfg *RechargeActivityConfig) {
	t.Helper()
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyRechargeActivityConfig] = string(raw)
}

func TestNormalizeRechargeActivityConfig_AllowsExpandablePrizeList(t *testing.T) {
	cfg := defaultRechargeActivityConfig()
	cfg.Prizes = append(cfg.Prizes, RechargeActivityPrize{
		ID: "special", Name: "特别奖", RewardDescription: "线下人工发放", Probability: 5, MinPayAmount: 10, Enabled: false, SortOrder: 4,
	})

	normalized, err := normalizeRechargeActivityConfig(cfg)

	require.NoError(t, err)
	require.Len(t, normalized.Prizes, 4)
	require.Equal(t, "first", normalized.Prizes[0].ID)
	require.Equal(t, "special", normalized.Prizes[3].ID)
}

func TestRechargeActivityEligiblePrizesFilterByPayAmount(t *testing.T) {
	cfg := defaultRechargeActivityConfig()

	eligible := rechargeActivityEligiblePrizes(cfg.Prizes, 20)

	require.Len(t, eligible, 1)
	require.Equal(t, "third", eligible[0].ID)
}

func TestNormalizeRechargeActivityConfig_RejectsEnabledProbabilityTotalNotOneHundred(t *testing.T) {
	cfg := defaultRechargeActivityConfig()
	cfg.Prizes[0].Probability = 60

	_, err := normalizeRechargeActivityConfig(cfg)

	require.Error(t, err)
	require.Contains(t, err.Error(), "probability")
}

func TestRechargeActivityDrawNormalizesEligiblePrizeProbabilities(t *testing.T) {
	cfg := defaultRechargeActivityConfig()
	setRechargeActivityTestDrawSequence(t, 0.99)

	prize, err := drawRechargeActivityPrize(cfg.Prizes, 20)

	require.NoError(t, err)
	require.Equal(t, "third", prize.ID)
}

func TestRechargeActivityGrantChanceForBalanceAndSubscriptionOrders(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())

	user := createLuckyWheelTestUser(t, ctx, svc, "recharge-activity@example.com")
	balanceOrder := createLuckyWheelBalanceOrder(t, ctx, svc, user, 1, 360, 20)
	subscriptionOrder := createLuckyWheelSubscriptionOrder(t, ctx, svc, user, 2, 200, 30, luckyWheelPtrFloat64(1), 30, 30)

	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, balanceOrder))
	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, subscriptionOrder))
	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, balanceOrder))

	summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, summary.PendingChances, 2)
}

func TestRechargeActivityDrawRecordsManualRewardWithoutBalanceCredit(t *testing.T) {
	ctx := context.Background()
	svc, repo, userRepo := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	setRechargeActivityTestDrawSequence(t, 0.99)
	updateBalanceCalls := 0
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		updateBalanceCalls++
		userRepo.getByIDUser.Balance += amount
		return nil
	}

	user := createLuckyWheelTestUser(t, ctx, svc, "recharge-activity-draw@example.com")
	order := createLuckyWheelBalanceOrder(t, ctx, svc, user, 3, 360, 20)
	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, order))
	summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, summary.PendingChances, 1)

	result, err := svc.DrawRechargeActivity(ctx, user.ID, summary.PendingChances[0].ID)

	require.NoError(t, err)
	require.Equal(t, "third", result.Record.PrizeID)
	require.Equal(t, 0.0, result.Record.RewardAmount)
	require.Equal(t, "联系客服领取实体礼品", result.Record.RewardDescription)
	require.Equal(t, RechargeActivityFulfillmentPending, result.Record.FulfillmentStatus)
	require.Equal(t, 0, updateBalanceCalls)
	require.Equal(t, 0.0, userRepo.getByIDUser.Balance)
	reloaded, err := rechargeActivityLoadChance(ctx, svc.entClient, result.ChanceID, user.ID)
	require.NoError(t, err)
	require.True(t, reloaded.Drawn)
	require.NotNil(t, reloaded.DrawnAt)
}

func TestRechargeActivityRecordFulfillmentCanBeUpdatedByAdmin(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	setRechargeActivityTestDrawSequence(t, 0.99)

	user := createLuckyWheelTestUser(t, ctx, svc, "recharge-activity-fulfillment@example.com")
	order := createLuckyWheelBalanceOrder(t, ctx, svc, user, 5, 360, 20)
	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, order))
	summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
	require.NoError(t, err)
	result, err := svc.DrawRechargeActivity(ctx, user.ID, summary.PendingChances[0].ID)
	require.NoError(t, err)

	record, err := svc.UpdateRechargeActivityRecordFulfillment(ctx, result.Record.ID, 99, RechargeActivityFulfillmentFulfilled, "已联系用户发放")

	require.NoError(t, err)
	require.Equal(t, RechargeActivityFulfillmentFulfilled, record.FulfillmentStatus)
	require.Equal(t, "已联系用户发放", record.FulfillmentNote)
	require.NotNil(t, record.FulfilledAt)
	require.NotNil(t, record.FulfilledBy)
	require.Equal(t, int64(99), *record.FulfilledBy)
}

func TestRechargeActivityRecordFulfillmentRequiresNoteWhenFulfilled(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	setRechargeActivityTestDrawSequence(t, 0.99)

	user := createLuckyWheelTestUser(t, ctx, svc, "recharge-activity-fulfillment-note@example.com")
	order := createLuckyWheelBalanceOrder(t, ctx, svc, user, 6, 360, 20)
	require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, order))
	summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
	require.NoError(t, err)
	result, err := svc.DrawRechargeActivity(ctx, user.ID, summary.PendingChances[0].ID)
	require.NoError(t, err)

	_, err = svc.UpdateRechargeActivityRecordFulfillment(ctx, result.Record.ID, 99, RechargeActivityFulfillmentFulfilled, "   ")

	require.Error(t, err)
	require.Contains(t, err.Error(), "RECHARGE_ACTIVITY_FULFILLMENT_NOTE_REQUIRED")

	record, err := svc.UpdateRechargeActivityRecordFulfillment(ctx, result.Record.ID, 99, RechargeActivityFulfillmentPending, "   ")
	require.NoError(t, err)
	require.Equal(t, RechargeActivityFulfillmentPending, record.FulfillmentStatus)
	require.Empty(t, record.FulfillmentNote)
	require.Nil(t, record.FulfilledAt)
	require.Nil(t, record.FulfilledBy)
}

func TestRechargeActivityStatsPaginatesRecentRecordsWithUserInfo(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	setRechargeActivityTestDrawSequence(t, 0.99)

	firstUser := createLuckyWheelTestUser(t, ctx, svc, "first-winner@example.com")
	secondUser := createLuckyWheelTestUser(t, ctx, svc, "second-winner@example.com")
	thirdUser := createLuckyWheelTestUser(t, ctx, svc, "third-winner@example.com")
	_, err := svc.entClient.User.UpdateOneID(firstUser.ID).SetUsername("first-winner").Save(ctx)
	require.NoError(t, err)
	_, err = svc.entClient.User.UpdateOneID(secondUser.ID).SetUsername("second-winner").Save(ctx)
	require.NoError(t, err)
	_, err = svc.entClient.User.UpdateOneID(thirdUser.ID).SetUsername("third-winner").Save(ctx)
	require.NoError(t, err)

	for idx, user := range []*dbent.User{firstUser, secondUser, thirdUser} {
		order := createLuckyWheelBalanceOrder(t, ctx, svc, user, int64(30+idx), 360, 20)
		require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, order))
		summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, summary.PendingChances, 1)
		_, err = svc.DrawRechargeActivity(ctx, user.ID, summary.PendingChances[0].ID)
		require.NoError(t, err)
	}

	stats, err := svc.GetRechargeActivityStats(ctx, RechargeActivityStatsQuery{Page: 2, PageSize: 1})

	require.NoError(t, err)
	require.Equal(t, int64(3), stats.TotalChances)
	require.Equal(t, int64(3), stats.DrawnChances)
	require.Equal(t, int64(3), stats.PendingFulfillments)
	require.Equal(t, int64(3), stats.RecentRecordsTotal)
	require.Equal(t, 2, stats.RecentRecordsPage)
	require.Equal(t, 1, stats.RecentRecordsPageSize)
	require.Len(t, stats.RecentRecords, 1)
	require.Equal(t, secondUser.ID, stats.RecentRecords[0].UserID)
	require.Equal(t, "second-winner@example.com", stats.RecentRecords[0].UserEmail)
	require.Equal(t, "second-winner", stats.RecentRecords[0].UserName)
}

func TestRechargeActivityStatsSearchesRecordsByWinningUser(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	setRechargeActivityTestDrawSequence(t, 0.99)

	alphaUser := createLuckyWheelTestUser(t, ctx, svc, "alpha-winner@example.com")
	betaUser := createLuckyWheelTestUser(t, ctx, svc, "beta-winner@example.com")
	_, err := svc.entClient.User.UpdateOneID(alphaUser.ID).SetUsername("alpha-name").Save(ctx)
	require.NoError(t, err)
	_, err = svc.entClient.User.UpdateOneID(betaUser.ID).SetUsername("beta-name").Save(ctx)
	require.NoError(t, err)

	for idx, user := range []*dbent.User{alphaUser, betaUser} {
		order := createLuckyWheelBalanceOrder(t, ctx, svc, user, int64(40+idx), 360, 20)
		require.NoError(t, svc.GrantRechargeActivityChanceForOrder(ctx, order))
		summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, summary.PendingChances, 1)
		_, err = svc.DrawRechargeActivity(ctx, user.ID, summary.PendingChances[0].ID)
		require.NoError(t, err)
	}

	byEmail, err := svc.GetRechargeActivityStats(ctx, RechargeActivityStatsQuery{Page: 1, PageSize: 20, UserKeyword: "ALPHA-WINNER"})
	require.NoError(t, err)
	require.Equal(t, int64(2), byEmail.TotalChances)
	require.Equal(t, int64(2), byEmail.DrawnChances)
	require.Equal(t, int64(1), byEmail.RecentRecordsTotal)
	require.Len(t, byEmail.RecentRecords, 1)
	require.Equal(t, alphaUser.ID, byEmail.RecentRecords[0].UserID)
	require.Equal(t, "ALPHA-WINNER", byEmail.RecentRecordsKeyword)

	byName, err := svc.GetRechargeActivityStats(ctx, RechargeActivityStatsQuery{UserKeyword: "beta-name"})
	require.NoError(t, err)
	require.Equal(t, int64(2), byName.PendingFulfillments)
	require.Equal(t, int64(1), byName.RecentRecordsTotal)
	require.Len(t, byName.RecentRecords, 1)
	require.Equal(t, betaUser.ID, byName.RecentRecords[0].UserID)

	byID, err := svc.GetRechargeActivityStats(ctx, RechargeActivityStatsQuery{UserKeyword: strconv.FormatInt(alphaUser.ID, 10)})
	require.NoError(t, err)
	require.Equal(t, int64(1), byID.RecentRecordsTotal)
	require.Len(t, byID.RecentRecords, 1)
	require.Equal(t, alphaUser.ID, byID.RecentRecords[0].UserID)
}

func TestMarkCompletedGrantsRechargeActivityWithoutSuppressingLuckyWheel(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newRechargeActivityPaymentService(t)
	repo.values[SettingKeyLuckyWheelEnabled] = "true"
	saveRechargeActivityConfig(t, repo, defaultRechargeActivityConfig())
	saveLuckyWheelConfig(t, repo, defaultLuckyWheelConfig())

	user := createLuckyWheelTestUser(t, ctx, svc, "recharge-activity-coexist@example.com")
	order := createLuckyWheelBalanceOrder(t, ctx, svc, user, 4, 360, 20)

	require.NoError(t, svc.markCompleted(ctx, &dbent.PaymentOrder{
		ID:           order.ID,
		UserID:       order.UserID,
		OrderType:    order.OrderType,
		PayAmount:    order.PayAmount,
		Amount:       order.Amount,
		RechargeCode: order.RechargeCode,
	}, "RECHARGE_SUCCESS"))

	rechargeSummary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, rechargeSummary.PendingChances, 1)
	luckySummary, err := svc.GetLuckyWheelSummary(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, luckySummary.ActiveSession)
}
