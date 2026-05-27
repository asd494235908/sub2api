//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type paymentOrderLifecycleQueryProvider struct {
	key               string
	lastQueryTradeNo  string
	lastCancelTradeNo string
	queryCalls        int
	cancelCalls       int
	responses         []*payment.QueryOrderResponse
	resp              *payment.QueryOrderResponse
}

type paymentOrderLifecycleRedeemRepo struct {
	codesByCode map[string]*RedeemCode
	nextID      int64
	useCalls    []struct {
		id     int64
		userID int64
	}
}

type paymentOrderLifecycleGroupRepo struct {
	groupRepoNoop
	group *Group
}

func (r *paymentOrderLifecycleGroupRepo) GetByID(_ context.Context, id int64) (*Group, error) {
	if r.group == nil || r.group.ID != id {
		return nil, ErrGroupNotFound
	}
	group := *r.group
	return &group, nil
}

type paymentOrderLifecycleUserSubRepo struct {
	userSubRepoNoop
	nextID int64
	subs   map[string]*UserSubscription
}

func newPaymentOrderLifecycleUserSubRepo() *paymentOrderLifecycleUserSubRepo {
	return &paymentOrderLifecycleUserSubRepo{
		nextID: 1,
		subs:   map[string]*UserSubscription{},
	}
}

func (r *paymentOrderLifecycleUserSubRepo) key(userID, groupID int64) string {
	return strconvFormatInt(userID) + ":" + strconvFormatInt(groupID)
}

func (r *paymentOrderLifecycleUserSubRepo) GetByUserIDAndGroupID(_ context.Context, userID, groupID int64) (*UserSubscription, error) {
	sub := r.subs[r.key(userID, groupID)]
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	copy := *sub
	return &copy, nil
}

func (r *paymentOrderLifecycleUserSubRepo) Create(_ context.Context, sub *UserSubscription) error {
	if sub == nil {
		return nil
	}
	copy := *sub
	if copy.ID == 0 {
		copy.ID = r.nextID
		r.nextID++
	}
	sub.ID = copy.ID
	r.subs[r.key(copy.UserID, copy.GroupID)] = &copy
	return nil
}

func (r *paymentOrderLifecycleUserSubRepo) GetByID(_ context.Context, id int64) (*UserSubscription, error) {
	for _, sub := range r.subs {
		if sub.ID != id {
			continue
		}
		copy := *sub
		return &copy, nil
	}
	return nil, ErrSubscriptionNotFound
}

func (p *paymentOrderLifecycleQueryProvider) Name() string {
	return "payment-order-lifecycle-query-provider"
}

func (p *paymentOrderLifecycleQueryProvider) ProviderKey() string {
	if p.key != "" {
		return p.key
	}
	return payment.TypeAlipay
}

func (p *paymentOrderLifecycleQueryProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{p.ProviderKey()}
}

func (p *paymentOrderLifecycleQueryProvider) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}

func (p *paymentOrderLifecycleQueryProvider) QueryOrder(_ context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	p.lastQueryTradeNo = tradeNo
	p.queryCalls++
	if len(p.responses) > 0 {
		resp := p.responses[0]
		if len(p.responses) > 1 {
			p.responses = p.responses[1:]
		}
		return resp, nil
	}
	return p.resp, nil
}

func (p *paymentOrderLifecycleQueryProvider) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	panic("unexpected call")
}

func (p *paymentOrderLifecycleQueryProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

func (p *paymentOrderLifecycleQueryProvider) CancelPayment(_ context.Context, tradeNo string) error {
	p.lastCancelTradeNo = tradeNo
	p.cancelCalls++
	return nil
}

func (r *paymentOrderLifecycleRedeemRepo) Create(_ context.Context, code *RedeemCode) error {
	if r.codesByCode == nil {
		r.codesByCode = map[string]*RedeemCode{}
	}
	if r.nextID == 0 {
		r.nextID = 1
	}
	cloned := *code
	cloned.ID = r.nextID
	r.nextID++
	r.codesByCode[cloned.Code] = &cloned
	code.ID = cloned.ID
	return nil
}

func (r *paymentOrderLifecycleRedeemRepo) CreateBatch(context.Context, []RedeemCode) error {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) GetByID(_ context.Context, id int64) (*RedeemCode, error) {
	for _, code := range r.codesByCode {
		if code.ID != id {
			continue
		}
		cloned := *code
		return &cloned, nil
	}
	return nil, ErrRedeemCodeNotFound
}

func (r *paymentOrderLifecycleRedeemRepo) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	redeemCode, ok := r.codesByCode[code]
	if !ok {
		return nil, ErrRedeemCodeNotFound
	}
	cloned := *redeemCode
	return &cloned, nil
}

func (r *paymentOrderLifecycleRedeemRepo) Update(context.Context, *RedeemCode) error {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) BatchUpdate(context.Context, []int64, RedeemCodeBatchUpdateFields) (int64, error) {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) Delete(context.Context, int64) error {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) Use(_ context.Context, id, userID int64) error {
	for code, redeemCode := range r.codesByCode {
		if redeemCode.ID != id {
			continue
		}
		now := time.Now().UTC()
		redeemCode.Status = StatusUsed
		redeemCode.UsedBy = &userID
		redeemCode.UsedAt = &now
		r.codesByCode[code] = redeemCode
		r.useCalls = append(r.useCalls, struct {
			id     int64
			userID int64
		}{id: id, userID: userID})
		return nil
	}
	return ErrRedeemCodeNotFound
}

func (r *paymentOrderLifecycleRedeemRepo) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) ListByUser(context.Context, int64, int) ([]RedeemCode, error) {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected call")
}

func (r *paymentOrderLifecycleRedeemRepo) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	panic("unexpected call")
}

func TestVerifyOrderByOutTradeNoBackfillsTradeNoFromPaidQuery(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("checkpaid@example.com").
		SetPasswordHash("hash").
		SetUsername("checkpaid-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CHECKPAID-UPSTREAM-TRADE-NO").
		SetOutTradeNo("sub2_checkpaid_trade_no_missing").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		if userRepo.getByIDUser != nil {
			userRepo.getByIDUser.Balance += amount
		}
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		resp: &payment.QueryOrderResponse{
			TradeNo: "upstream-trade-123",
			Status:  payment.ProviderStatusPaid,
			Amount:  88,
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
	require.NoError(t, err)
	require.Equal(t, order.OutTradeNo, provider.lastQueryTradeNo)
	require.Equal(t, OrderStatusCompleted, got.Status)
	require.Equal(t, "upstream-trade-123", got.PaymentTradeNo)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.Equal(t, "upstream-trade-123", reloaded.PaymentTradeNo)

	require.Equal(t, 88.0, userRepo.getByIDUser.Balance)
	require.Len(t, redeemRepo.useCalls, 1)
	require.Equal(t, int64(1), redeemRepo.useCalls[0].id)
	require.Equal(t, user.ID, redeemRepo.useCalls[0].userID)
}

func TestVerifyOrderByOutTradeNoGrantsRechargeActivityChanceForCompletedOrders(t *testing.T) {
	tests := []struct {
		name      string
		orderType string
		build     func(t *testing.T, client *dbent.Client, user *dbent.User) *dbent.PaymentOrder
	}{
		{
			name:      "balance recharge",
			orderType: payment.OrderTypeBalance,
			build: func(t *testing.T, client *dbent.Client, user *dbent.User) *dbent.PaymentOrder {
				t.Helper()
				order, err := client.PaymentOrder.Create().
					SetUserID(user.ID).
					SetUserEmail(user.Email).
					SetUserName(user.Username).
					SetAmount(360).
					SetPayAmount(20).
					SetFeeRate(0).
					SetRechargeCode("RECHARGE-ACTIVITY-BALANCE").
					SetOutTradeNo("sub2_recharge_activity_balance").
					SetPaymentType(payment.TypeAlipay).
					SetPaymentTradeNo("").
					SetOrderType(payment.OrderTypeBalance).
					SetStatus(OrderStatusPending).
					SetExpiresAt(time.Now().Add(time.Hour)).
					SetClientIP("127.0.0.1").
					SetSrcHost("api.example.com").
					Save(context.Background())
				require.NoError(t, err)
				return order
			},
		},
		{
			name:      "subscription purchase",
			orderType: payment.OrderTypeSubscription,
			build: func(t *testing.T, client *dbent.Client, user *dbent.User) *dbent.PaymentOrder {
				t.Helper()
				order, err := client.PaymentOrder.Create().
					SetUserID(user.ID).
					SetUserEmail(user.Email).
					SetUserName(user.Username).
					SetAmount(30).
					SetPayAmount(30).
					SetFeeRate(0).
					SetRechargeCode("RECHARGE-ACTIVITY-SUBSCRIPTION").
					SetOutTradeNo("sub2_recharge_activity_subscription").
					SetPaymentType(payment.TypeAlipay).
					SetPaymentTradeNo("").
					SetOrderType(payment.OrderTypeSubscription).
					SetSubscriptionGroupID(10).
					SetSubscriptionDays(30).
					SetStatus(OrderStatusPending).
					SetExpiresAt(time.Now().Add(time.Hour)).
					SetClientIP("127.0.0.1").
					SetSrcHost("api.example.com").
					Save(context.Background())
				require.NoError(t, err)
				return order
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentOrderLifecycleTestClient(t)
			user, err := client.User.Create().
				SetEmail("recharge-activity-" + tt.orderType + "@example.com").
				SetPasswordHash("hash").
				SetUsername("recharge-activity-" + tt.orderType + "-user").
				Save(ctx)
			require.NoError(t, err)
			order := tt.build(t, client, user)

			userRepo := &mockUserRepo{
				getByIDUser: &User{ID: user.ID, Email: user.Email, Username: user.Username, Balance: 0},
			}
			userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
				require.Equal(t, user.ID, id)
				userRepo.getByIDUser.Balance += amount
				return nil
			}
			redeemRepo := &paymentOrderLifecycleRedeemRepo{}
			redeemService := NewRedeemService(redeemRepo, userRepo, nil, nil, nil, client, nil, nil, nil)
			groupRepo := &paymentOrderLifecycleGroupRepo{
				group: &Group{ID: 10, Status: payment.EntityStatusActive, SubscriptionType: SubscriptionTypeSubscription},
			}
			subRepo := newPaymentOrderLifecycleUserSubRepo()
			subscriptionService := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)
			settingRepo := &settingPublicRepoStub{
				values: map[string]string{
					SettingPaymentEnabled:             "true",
					SettingKeyRechargeActivityEnabled: "true",
				},
			}
			saveRechargeActivityConfig(t, settingRepo, defaultRechargeActivityConfig())
			configService := NewPaymentConfigService(client, settingRepo, nil)
			registry := payment.NewRegistry()
			provider := &paymentOrderLifecycleQueryProvider{
				resp: &payment.QueryOrderResponse{
					TradeNo: "upstream-" + tt.orderType,
					Status:  payment.ProviderStatusPaid,
					Amount:  order.PayAmount,
				},
			}
			registry.Register(provider)
			svc := &PaymentService{
				entClient:       client,
				registry:        registry,
				redeemService:   redeemService,
				subscriptionSvc: subscriptionService,
				configService:   configService,
				userRepo:        userRepo,
				groupRepo:       groupRepo,
				providersLoaded: true,
			}

			got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
			require.NoError(t, err)
			require.Equal(t, OrderStatusCompleted, got.Status)

			summary, err := svc.GetRechargeActivitySummary(ctx, user.ID)
			require.NoError(t, err)
			require.Len(t, summary.PendingChances, 1)
			require.Equal(t, int64(order.ID), summary.PendingChances[0].SourceOrderID)
			require.Equal(t, tt.orderType, summary.PendingChances[0].SourceOrderType)
			require.InDelta(t, order.PayAmount, summary.PendingChances[0].SourcePayAmount, 1e-9)

			_, err = svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
			require.NoError(t, err)
			summary, err = svc.GetRechargeActivitySummary(ctx, user.ID)
			require.NoError(t, err)
			require.Len(t, summary.PendingChances, 1)
		})
	}
}

func TestExecuteBalanceFulfillmentCreatesRedeemCodeFromCreditedAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("balance-credited-amount@example.com").
		SetPasswordHash("hash").
		SetUsername("balance-credited-amount-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(360).
		SetPayAmount(20).
		SetFeeRate(0).
		SetRechargeCode("BALANCE-CREDITED-AMOUNT").
		SetOutTradeNo("sub2_balance_credited_amount").
		SetPaymentType(payment.TypeWxpay).
		SetPaymentTradeNo("wxpay-trade-credited-amount").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPaid).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		require.InDelta(t, 360.0, amount, 1e-9)
		userRepo.getByIDUser.Balance += amount
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	svc := &PaymentService{
		entClient:       client,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	require.NoError(t, svc.ExecuteBalanceFulfillment(ctx, order.ID))

	created := redeemRepo.codesByCode[order.RechargeCode]
	require.NotNil(t, created)
	require.Equal(t, RedeemTypeBalance, created.Type)
	require.InDelta(t, 360.0, created.Value, 1e-9)
	require.Len(t, redeemRepo.useCalls, 1)
	require.InDelta(t, 360.0, userRepo.getByIDUser.Balance, 1e-9)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	reloadedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 20.0, reloadedUser.TotalRecharged, 1e-9)
}

func TestExecuteBalanceFulfillmentAppliesFirstRechargeBonusOncePerTier(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("first-recharge@example.com").
		SetPasswordHash("hash").
		SetUsername("first-recharge-user").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		userRepo.getByIDUser.Balance += amount
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	settingRepo := &settingPublicRepoStub{values: map[string]string{
		SettingPaymentEnabled: "true",
	}}
	configService := NewPaymentConfigService(client, settingRepo, nil)
	firstRechargeCfg := &FirstRechargeConfig{Tiers: []FirstRechargeTier{
		{ID: "tier-30", PayAmount: 30, BonusAmount: 30, Enabled: true, SortOrder: 1},
	}}
	svc := &PaymentService{
		entClient:       client,
		redeemService:   redeemService,
		configService:   configService,
		userRepo:        userRepo,
		providersLoaded: true,
	}
	_, err = svc.UpdateFirstRechargeConfig(ctx, true, firstRechargeCfg)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		order, err := client.PaymentOrder.Create().
			SetUserID(user.ID).
			SetUserEmail(user.Email).
			SetUserName(user.Username).
			SetAmount(30).
			SetPayAmount(30).
			SetFeeRate(0).
			SetRechargeCode("FIRST-RECHARGE-" + strconvFormatInt(int64(i))).
			SetOutTradeNo("sub2_first_recharge_" + strconvFormatInt(int64(i))).
			SetPaymentType(payment.TypeWxpay).
			SetPaymentTradeNo("trade-first-recharge-" + strconvFormatInt(int64(i))).
			SetOrderType(payment.OrderTypeBalance).
			SetStatus(OrderStatusPaid).
			SetExpiresAt(time.Now().Add(time.Hour)).
			SetClientIP("127.0.0.1").
			SetSrcHost("api.example.com").
			Save(ctx)
		require.NoError(t, err)
		require.NoError(t, svc.ExecuteBalanceFulfillment(ctx, order.ID))
	}

	var bonusCodes []RedeemCode
	for _, code := range redeemRepo.codesByCode {
		if code.Type == RedeemTypeFirstRechargeBonus {
			bonusCodes = append(bonusCodes, *code)
		}
	}
	require.Len(t, bonusCodes, 1)
	require.InDelta(t, 30.0, bonusCodes[0].Value, 1e-9)
	require.Contains(t, bonusCodes[0].Notes, "tier-30")
	require.InDelta(t, 90.0, userRepo.getByIDUser.Balance, 1e-9)
	reloadedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 60.0, reloadedUser.TotalRecharged, 1e-9)

	result, err := svc.GrantFirstRechargeChance(ctx, user.ID, "tier-30", 1, "admin", "manual grant")
	require.NoError(t, err)
	require.Equal(t, 1, result.Available)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(30).
		SetPayAmount(30).
		SetFeeRate(0).
		SetRechargeCode("FIRST-RECHARGE-GRANTED").
		SetOutTradeNo("sub2_first_recharge_granted").
		SetPaymentType(payment.TypeWxpay).
		SetPaymentTradeNo("trade-first-recharge-granted").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPaid).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, svc.ExecuteBalanceFulfillment(ctx, order.ID))

	bonusCodes = bonusCodes[:0]
	for _, code := range redeemRepo.codesByCode {
		if code.Type == RedeemTypeFirstRechargeBonus {
			bonusCodes = append(bonusCodes, *code)
		}
	}
	require.Len(t, bonusCodes, 2)
	require.InDelta(t, 150.0, userRepo.getByIDUser.Balance, 1e-9)
	reloadedUser, err = client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 90.0, reloadedUser.TotalRecharged, 1e-9)
}

func TestBulkUpdateFirstRechargeChancesPreservesHistory(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user1, err := client.User.Create().
		SetEmail("first-recharge-bulk-1@example.com").
		SetPasswordHash("hash").
		SetUsername("bulk-user-1").
		Save(ctx)
	require.NoError(t, err)
	user2, err := client.User.Create().
		SetEmail("first-recharge-bulk-2@example.com").
		SetPasswordHash("hash").
		SetUsername("bulk-user-2").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	require.NoError(t, svc.ensureFirstRechargeTables(ctx))
	now := time.Now().UTC()
	require.NoError(t, firstRechargeAddChance(ctx, client, user1.ID, "tier-30", 1, "admin", "seed", now))
	claimed, err := firstRechargeClaimChance(ctx, client, user1.ID, "tier-30", 1001, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, claimed)

	result, err := svc.BulkUpdateFirstRechargeChances(ctx, "tier-30", 2, FirstRechargeBulkChanceModeGrant, "admin", "bulk grant")
	require.NoError(t, err)
	require.Equal(t, 2, result.AffectedUsers)

	state1, err := firstRechargeLoadChance(ctx, client, user1.ID, "tier-30")
	require.NoError(t, err)
	require.Equal(t, 2, state1.Available)
	require.Equal(t, 1, state1.Consumed)
	state2, err := firstRechargeLoadChance(ctx, client, user2.ID, "tier-30")
	require.NoError(t, err)
	require.Equal(t, 2, state2.Available)
	require.Equal(t, 0, state2.Consumed)

	result, err = svc.BulkUpdateFirstRechargeChances(ctx, "tier-30", 1, FirstRechargeBulkChanceModeReset, "admin", "bulk reset")
	require.NoError(t, err)
	require.Equal(t, 2, result.AffectedUsers)
	state1, err = firstRechargeLoadChance(ctx, client, user1.ID, "tier-30")
	require.NoError(t, err)
	require.Equal(t, 1, state1.Available)
	require.Equal(t, 0, state1.Consumed)

	execCtx, err := luckyWheelExecContext(ctx, client)
	require.NoError(t, err)
	rows, err := execCtx.QueryContext(ctx, "SELECT COUNT(*) FROM first_recharge_consumption_logs WHERE user_id = ? AND tier_id = ?", user1.ID, "tier-30")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var consumptionLogs int
	require.NoError(t, rows.Scan(&consumptionLogs))
	require.Equal(t, 1, consumptionLogs)

	rows, err = execCtx.QueryContext(ctx, "SELECT COUNT(*) FROM first_recharge_grant_logs WHERE tier_id = ?", "tier-30")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var grantLogs int
	require.NoError(t, rows.Scan(&grantLogs))
	require.Equal(t, 5, grantLogs)
}

func TestResolveMemberLevelUsesTotalRechargedAndBalanceKeyOnly(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)
	settingRepo := &settingPublicRepoStub{values: map[string]string{
		SettingPaymentEnabled: "true",
	}}
	svc := &PaymentService{
		entClient:     client,
		configService: NewPaymentConfigService(client, settingRepo, nil),
	}
	_, err := svc.UpdateMemberLevelConfig(ctx, true, &MemberLevelConfig{Levels: []MemberLevel{
		{ID: "lv1", Name: "Lv1", MinRechargeAmount: 0, RateMultiplier: 1, Enabled: true, SortOrder: 1},
		{ID: "lv2", Name: "Lv2", MinRechargeAmount: 500, RateMultiplier: 0.8, Enabled: true, SortOrder: 2},
		{ID: "lv3", Name: "Lv3", MinRechargeAmount: 1000, RateMultiplier: 0.6, Enabled: true, SortOrder: 3},
	}})
	require.NoError(t, err)

	user := &User{ID: 1, TotalRecharged: 750}
	state, err := svc.ResolveMemberLevel(ctx, user)
	require.NoError(t, err)
	require.Equal(t, "lv2", state.LevelID)
	require.InDelta(t, 0.8, state.RateMultiplier, 1e-9)
	require.NotNil(t, state.NextThreshold)
	require.InDelta(t, 1000.0, *state.NextThreshold, 1e-9)
	require.InDelta(t, 50.0, state.Progress, 1e-9)

	balanceKey := &APIKey{ID: 1, Group: &Group{ID: 10, SubscriptionType: SubscriptionTypeStandard}}
	require.True(t, svc.MemberMultiplierEnabledForKey(ctx, balanceKey, nil))
	require.InDelta(t, 0.8, svc.ResolveMemberMultiplier(ctx, user, balanceKey, nil), 1e-9)

	subscriptionKey := &APIKey{ID: 2, Group: &Group{ID: 11, SubscriptionType: SubscriptionTypeSubscription}}
	require.False(t, svc.MemberMultiplierEnabledForKey(ctx, subscriptionKey, &UserSubscription{ID: 99}))
	require.InDelta(t, 1.0, svc.ResolveMemberMultiplier(ctx, user, subscriptionKey, &UserSubscription{ID: 99}), 1e-9)

	_, err = svc.UpdateMemberLevelConfig(ctx, false, &MemberLevelConfig{Levels: []MemberLevel{
		{ID: "lv1", Name: "Lv1", MinRechargeAmount: 0, RateMultiplier: 1, Enabled: true, SortOrder: 1},
		{ID: "lv2", Name: "Lv2", MinRechargeAmount: 500, RateMultiplier: 0.8, Enabled: true, SortOrder: 2},
	}})
	require.NoError(t, err)
	require.False(t, svc.MemberMultiplierEnabledForKey(ctx, balanceKey, nil))
}

func TestVerifyOrderByOutTradeNoRetriesZeroAmountPaidQueryOnce(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("checkpaid-retry@example.com").
		SetPasswordHash("hash").
		SetUsername("checkpaid-retry-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CHECKPAID-UPSTREAM-RETRY").
		SetOutTradeNo("sub2_checkpaid_retry_zero_amount").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		if userRepo.getByIDUser != nil {
			userRepo.getByIDUser.Balance += amount
		}
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		responses: []*payment.QueryOrderResponse{
			{
				TradeNo: "upstream-trade-zero",
				Status:  payment.ProviderStatusPaid,
				Amount:  0,
			},
			{
				TradeNo: "upstream-trade-retry",
				Status:  payment.ProviderStatusPaid,
				Amount:  88,
			},
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
	require.NoError(t, err)
	require.Equal(t, 2, provider.queryCalls)
	require.Equal(t, OrderStatusCompleted, got.Status)
	require.Equal(t, "upstream-trade-retry", got.PaymentTradeNo)
}

func TestVerifyOrderByOutTradeNoRejectsPaidQueryWithZeroAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("checkpaid-zero-amount@example.com").
		SetPasswordHash("hash").
		SetUsername("checkpaid-zero-amount-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CHECKPAID-ZERO-AMOUNT").
		SetOutTradeNo("sub2_checkpaid_zero_amount").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		resp: &payment.QueryOrderResponse{
			TradeNo: "upstream-trade-zero",
			Status:  payment.ProviderStatusPaid,
			Amount:  0,
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
	require.NoError(t, err)
	require.Equal(t, order.OutTradeNo, provider.lastQueryTradeNo)
	require.Equal(t, OrderStatusPending, got.Status)
	require.Empty(t, got.PaymentTradeNo)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, reloaded.Status)
	require.Empty(t, reloaded.PaymentTradeNo)

	require.Equal(t, 0.0, userRepo.getByIDUser.Balance)
	require.Empty(t, redeemRepo.useCalls)
}

func TestVerifyOrderByOutTradeNoDoesNotCancelUnpaidUpstreamOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("checkpaid-pending@example.com").
		SetPasswordHash("hash").
		SetUsername("checkpaid-pending-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CHECKPAID-PENDING").
		SetOutTradeNo("sub2_checkpaid_pending").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		resp: &payment.QueryOrderResponse{
			TradeNo: order.OutTradeNo,
			Status:  payment.ProviderStatusPending,
			Amount:  0,
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, got.Status)
	require.Equal(t, order.OutTradeNo, provider.lastQueryTradeNo)
	require.Zero(t, provider.cancelCalls)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, reloaded.Status)
}

func TestCancelOrderStillClosesUnpaidUpstreamOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("cancel-pending@example.com").
		SetPasswordHash("hash").
		SetUsername("cancel-pending-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CANCEL-PENDING").
		SetOutTradeNo("sub2_cancel_pending").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		resp: &payment.QueryOrderResponse{
			TradeNo: order.OutTradeNo,
			Status:  payment.ProviderStatusPending,
			Amount:  0,
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		providersLoaded: true,
	}

	outcome, err := svc.CancelOrder(ctx, order.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, checkPaidResultCancelled, outcome)
	require.Equal(t, order.OutTradeNo, provider.lastCancelTradeNo)
	require.Equal(t, 1, provider.cancelCalls)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, reloaded.Status)
}

func TestReconcilePendingWxpayOrdersBackfillsPaidOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("wxpay-reconcile@example.com").
		SetPasswordHash("hash").
		SetUsername("wxpay-reconcile-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(50).
		SetPayAmount(50).
		SetFeeRate(0).
		SetRechargeCode("WXPAY-RECONCILE").
		SetOutTradeNo("sub2_wxpay_reconcile").
		SetPaymentType(payment.TypeWxpay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		if userRepo.getByIDUser != nil {
			userRepo.getByIDUser.Balance += amount
		}
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		key: payment.TypeWxpay,
		resp: &payment.QueryOrderResponse{
			TradeNo: "wxpay-upstream-trade-123",
			Status:  payment.ProviderStatusPaid,
			Amount:  50,
			Metadata: map[string]string{
				"trade_state": "SUCCESS",
			},
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	recovered, err := svc.ReconcilePendingWxpayOrders(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, recovered)
	require.Equal(t, order.OutTradeNo, provider.lastQueryTradeNo)
	require.Zero(t, provider.cancelCalls)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.Equal(t, "wxpay-upstream-trade-123", reloaded.PaymentTradeNo)
	require.Equal(t, 50.0, userRepo.getByIDUser.Balance)
	require.Len(t, redeemRepo.useCalls, 1)
}

func TestVerifyOrderByOutTradeNoUsesOutTradeNoWhenPaymentTradeNoAlreadyExistsForAlipay(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("checkpaid-existing-trade@example.com").
		SetPasswordHash("hash").
		SetUsername("checkpaid-existing-trade-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("CHECKPAID-EXISTING-TRADE-NO").
		SetOutTradeNo("sub2_checkpaid_use_out_trade_no").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("upstream-trade-existing").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  0,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		if userRepo.getByIDUser != nil {
			userRepo.getByIDUser.Balance += amount
		}
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{
		codesByCode: map[string]*RedeemCode{
			order.RechargeCode: {
				ID:     1,
				Code:   order.RechargeCode,
				Type:   RedeemTypeBalance,
				Value:  order.Amount,
				Status: StatusUnused,
			},
		},
	}
	redeemService := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		nil,
		nil,
	)
	registry := payment.NewRegistry()
	provider := &paymentOrderLifecycleQueryProvider{
		resp: &payment.QueryOrderResponse{
			TradeNo: "upstream-trade-existing",
			Status:  payment.ProviderStatusPaid,
			Amount:  88,
		},
	}
	registry.Register(provider)

	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		redeemService:   redeemService,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	got, err := svc.VerifyOrderByOutTradeNo(ctx, order.OutTradeNo, user.ID)
	require.NoError(t, err)
	require.Equal(t, order.OutTradeNo, provider.lastQueryTradeNo)
	require.Equal(t, "upstream-trade-existing", got.PaymentTradeNo)
}

func TestPaymentOrderAllowsRegistryFallbackOnlyForLegacyOrdersWithoutPinnedProviderState(t *testing.T) {
	t.Parallel()

	require.True(t, paymentOrderAllowsRegistryFallback(&dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
	}))

	instanceID := "12"
	require.False(t, paymentOrderAllowsRegistryFallback(&dbent.PaymentOrder{
		PaymentType:        payment.TypeAlipay,
		ProviderInstanceID: &instanceID,
	}))

	require.False(t, paymentOrderAllowsRegistryFallback(&dbent.PaymentOrder{
		PaymentType: payment.TypeAlipay,
		ProviderSnapshot: map[string]any{
			"schema_version":       2,
			"provider_instance_id": "12",
		},
	}))
}

func TestPaymentOrderQueryReferenceUsesOutTradeNoForOfficialProviders(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		PaymentType:    payment.TypeWxpay,
		OutTradeNo:     "sub2_out_trade_no",
		PaymentTradeNo: "wx-transaction-id",
	}

	require.Equal(t, "sub2_out_trade_no", paymentOrderQueryReference(order, &paymentOrderLifecycleQueryProvider{}))
	require.Equal(t, "sub2_out_trade_no", paymentOrderQueryReference(order, paymentFulfillmentTestProvider{
		key: payment.TypeWxpay,
	}))
}

func newPaymentOrderLifecycleTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_order_lifecycle?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
