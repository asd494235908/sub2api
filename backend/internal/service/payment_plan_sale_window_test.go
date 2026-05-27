package service

import (
	"context"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestCreatePlanValidatesSaleWindowAndDailyPurchaseLimit(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)

	startsAt := time.Now().Add(time.Hour)
	endsAt := startsAt.Add(-time.Minute)
	_, err := svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:            1,
		Name:               "Future Plan",
		Price:              10,
		ValidityDays:       30,
		ValidityUnit:       "days",
		ForSale:            true,
		SaleStartsAt:       &startsAt,
		SaleEndsAt:         &endsAt,
		DailyPurchaseLimit: 0,
	})
	if err == nil || !strings.Contains(err.Error(), "sale end time must be after sale start time") {
		t.Fatalf("CreatePlan invalid sale window error = %v", err)
	}

	_, err = svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:            1,
		Name:               "Limited Plan",
		Price:              10,
		ValidityDays:       30,
		ValidityUnit:       "days",
		ForSale:            true,
		DailyPurchaseLimit: -1,
	})
	if err == nil || !strings.Contains(err.Error(), "daily purchase limit must be >= 0") {
		t.Fatalf("CreatePlan invalid daily purchase limit error = %v", err)
	}
}

func TestListPlansForSaleFiltersBySaleWindow(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)
	now := time.Now()

	_, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("available").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetSaleStartsAt(now.Add(-time.Hour)).
		SetSaleEndsAt(now.Add(time.Hour)).
		Save(ctx)
	if err != nil {
		t.Fatalf("create available plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("future").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetSaleStartsAt(now.Add(time.Hour)).
		Save(ctx)
	if err != nil {
		t.Fatalf("create future plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("expired").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetSaleEndsAt(now.Add(-time.Hour)).
		Save(ctx)
	if err != nil {
		t.Fatalf("create expired plan: %v", err)
	}

	plans, err := svc.ListPlansForSale(ctx)
	if err != nil {
		t.Fatalf("ListPlansForSale returned error: %v", err)
	}
	if len(plans) != 1 || plans[0].Name != "available" {
		t.Fatalf("ListPlansForSale = %#v, want only available", planNames(plans))
	}
}

func TestListPlansForSaleFiltersByDailySaleWindow(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)
	insideStart, insideEnd := dailyWindowAround(time.Now(), -time.Hour, time.Hour)
	outsideStart, outsideEnd := dailyWindowAround(time.Now(), time.Hour, 2*time.Hour)

	_, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("daily-available").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(insideStart).
		SetDailySaleEndsAt(insideEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create daily available plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("daily-unavailable").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(outsideStart).
		SetDailySaleEndsAt(outsideEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create daily unavailable plan: %v", err)
	}

	plans, err := svc.ListPlansForSale(ctx)
	if err != nil {
		t.Fatalf("ListPlansForSale returned error: %v", err)
	}
	if len(plans) != 1 || plans[0].Name != "daily-available" {
		t.Fatalf("ListPlansForSale = %#v, want only daily-available", planNames(plans))
	}
}

func TestListPlansVisibleOnCheckoutIncludesPendingDailySaleWindow(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)
	insideStart, insideEnd := dailyWindowAround(time.Now(), -time.Hour, time.Hour)
	outsideStart, outsideEnd := dailyWindowAround(time.Now(), time.Hour, 2*time.Hour)
	now := time.Now()

	_, err := client.Group.Create().
		SetName("subscription").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(PlatformOpenAI).
		SetSubscriptionType(SubscriptionTypeSubscription).
		Save(ctx)
	if err != nil {
		t.Fatalf("create subscription group: %v", err)
	}
	_, err = client.Group.Create().
		SetName("standard").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(PlatformOpenAI).
		SetSubscriptionType(SubscriptionTypeStandard).
		Save(ctx)
	if err != nil {
		t.Fatalf("create standard group: %v", err)
	}

	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("daily-available").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(insideStart).
		SetDailySaleEndsAt(insideEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create daily available plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("daily-pending").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(outsideStart).
		SetDailySaleEndsAt(outsideEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create daily pending plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("future-global").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetSaleStartsAt(now.Add(time.Hour)).
		Save(ctx)
	if err != nil {
		t.Fatalf("create future global plan: %v", err)
	}
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(2).
		SetName("standard-group").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(outsideStart).
		SetDailySaleEndsAt(outsideEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create standard group plan: %v", err)
	}

	plans, err := svc.ListCheckoutSubscriptionPlans(ctx)
	if err != nil {
		t.Fatalf("ListCheckoutSubscriptionPlans returned error: %v", err)
	}
	names := planNames(plans)
	if len(names) != 2 || names[0] != "daily-available" || names[1] != "daily-pending" {
		t.Fatalf("ListCheckoutSubscriptionPlans = %v, want daily-available and daily-pending", names)
	}
}

func TestSubscriptionPlanDailyPurchaseRemainingCountsActiveOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)
	now := time.Now()

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("limited").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(3).
		Save(ctx)
	if err != nil {
		t.Fatalf("create limited plan: %v", err)
	}
	user, err := client.User.Create().
		SetEmail("remaining@example.com").
		SetPasswordHash("hash").
		SetUsername("remaining-user").
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	for i, status := range []string{OrderStatusPending, OrderStatusCompleted, OrderStatusCancelled} {
		_, err = client.PaymentOrder.Create().
			SetUserID(user.ID).
			SetUserEmail("remaining@example.com").
			SetUserName("remaining-user").
			SetAmount(10).
			SetPayAmount(10).
			SetFeeRate(0).
			SetRechargeCode("PAY-remaining").
			SetOutTradeNo("sub2_remaining_" + status).
			SetPaymentType(payment.TypeAlipay).
			SetPaymentTradeNo("").
			SetOrderType(payment.OrderTypeSubscription).
			SetStatus(status).
			SetPlanID(plan.ID).
			SetExpiresAt(now.Add(time.Hour)).
			SetCreatedAt(now.Add(time.Duration(i) * time.Minute)).
			SetClientIP("127.0.0.1").
			SetSrcHost("localhost").
			Save(ctx)
		if err != nil {
			t.Fatalf("create order %s: %v", status, err)
		}
	}

	remaining, err := svc.SubscriptionPlanDailyPurchaseRemaining(ctx, plan, now)
	if err != nil {
		t.Fatalf("SubscriptionPlanDailyPurchaseRemaining returned error: %v", err)
	}
	if remaining == nil || *remaining != 1 {
		t.Fatalf("remaining = %v, want 1", remaining)
	}

	plan.DailyPurchaseLimit = 0
	remaining, err = svc.SubscriptionPlanDailyPurchaseRemaining(ctx, plan, now)
	if err != nil {
		t.Fatalf("SubscriptionPlanDailyPurchaseRemaining unlimited returned error: %v", err)
	}
	if remaining != nil {
		t.Fatalf("remaining for unlimited plan = %v, want nil", *remaining)
	}
}

func TestValidateSubOrderRejectsUnavailableSaleWindowAndDailyLimit(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	now := time.Now()
	groupRepo := &paymentPlanSaleWindowGroupRepo{
		group: &Group{ID: 1, Status: payment.EntityStatusActive, SubscriptionType: SubscriptionTypeSubscription},
	}
	svc := &PaymentService{
		entClient:     client,
		configService: NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil),
		groupRepo:     groupRepo,
	}

	futurePlan, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("future").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetSaleStartsAt(now.Add(time.Hour)).
		Save(ctx)
	if err != nil {
		t.Fatalf("create future plan: %v", err)
	}
	if _, err := svc.validateSubOrder(ctx, CreateOrderRequest{PlanID: futurePlan.ID}); err == nil || !strings.Contains(err.Error(), "plan not found or not for sale") {
		t.Fatalf("validateSubOrder future plan error = %v", err)
	}

	dailyStart, dailyEnd := dailyWindowAround(time.Now(), time.Hour, 2*time.Hour)
	dailyUnavailablePlan, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("daily-unavailable").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(0).
		SetDailySaleStartsAt(dailyStart).
		SetDailySaleEndsAt(dailyEnd).
		Save(ctx)
	if err != nil {
		t.Fatalf("create daily unavailable plan: %v", err)
	}
	if _, err := svc.validateSubOrder(ctx, CreateOrderRequest{PlanID: dailyUnavailablePlan.ID}); err == nil || !strings.Contains(err.Error(), "plan not found or not for sale") {
		t.Fatalf("validateSubOrder daily unavailable plan error = %v", err)
	}

	limitedPlan, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("limited").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("create limited plan: %v", err)
	}
	user, err := client.User.Create().
		SetEmail("plan-limit@example.com").
		SetPasswordHash("hash").
		SetUsername("plan-limit-user").
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, err = client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail("user@example.com").
		SetUserName("user").
		SetAmount(10).
		SetPayAmount(10).
		SetFeeRate(0).
		SetRechargeCode("PAY-1").
		SetOutTradeNo("sub2_plan_limit_1").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusPending).
		SetPlanID(limitedPlan.ID).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("localhost").
		Save(ctx)
	if err != nil {
		t.Fatalf("create counted order: %v", err)
	}
	if _, err := svc.validateSubOrder(ctx, CreateOrderRequest{PlanID: limitedPlan.ID}); err == nil || !strings.Contains(err.Error(), "daily purchase limit reached") {
		t.Fatalf("validateSubOrder daily limit error = %v", err)
	}
}

func TestDailySaleWindowValidation(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{}, nil)

	_, err := svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:           1,
		Name:              "Daily Partial",
		Price:             10,
		ValidityDays:      30,
		ValidityUnit:      "days",
		ForSale:           true,
		DailySaleStartsAt: "09:00",
	})
	if err == nil || !strings.Contains(err.Error(), "daily sale start and end time must be set together") {
		t.Fatalf("CreatePlan partial daily window error = %v", err)
	}

	_, err = svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:           1,
		Name:              "Daily Invalid",
		Price:             10,
		ValidityDays:      30,
		ValidityUnit:      "days",
		ForSale:           true,
		DailySaleStartsAt: "9:00",
		DailySaleEndsAt:   "18:00",
	})
	if err == nil || !strings.Contains(err.Error(), "daily sale time must use HH:mm format") {
		t.Fatalf("CreatePlan invalid daily window error = %v", err)
	}

	_, err = svc.CreatePlan(ctx, CreatePlanRequest{
		GroupID:           1,
		Name:              "Daily Empty",
		Price:             10,
		ValidityDays:      30,
		ValidityUnit:      "days",
		ForSale:           true,
		DailySaleStartsAt: "",
		DailySaleEndsAt:   "",
	})
	if err != nil {
		t.Fatalf("CreatePlan empty daily window error = %v", err)
	}
}

func TestSubscriptionPlanDailySaleWindowAvailability(t *testing.T) {
	dayWindow := &dbent.SubscriptionPlan{
		ForSale:           true,
		DailySaleStartsAt: stringPtr("09:00"),
		DailySaleEndsAt:   stringPtr("18:00"),
	}
	if !subscriptionPlanIsAvailableForSale(dayWindow, time.Date(2026, 5, 27, 10, 0, 0, 0, time.Local)) {
		t.Fatalf("day window should be available inside daily sale window")
	}
	if subscriptionPlanIsAvailableForSale(dayWindow, time.Date(2026, 5, 27, 8, 59, 0, 0, time.Local)) {
		t.Fatalf("day window should be unavailable before daily sale window")
	}

	overnightWindow := &dbent.SubscriptionPlan{
		ForSale:           true,
		DailySaleStartsAt: stringPtr("22:00"),
		DailySaleEndsAt:   stringPtr("02:00"),
	}
	if !subscriptionPlanIsAvailableForSale(overnightWindow, time.Date(2026, 5, 27, 23, 0, 0, 0, time.Local)) {
		t.Fatalf("overnight window should be available before midnight")
	}
	if !subscriptionPlanIsAvailableForSale(overnightWindow, time.Date(2026, 5, 28, 1, 0, 0, 0, time.Local)) {
		t.Fatalf("overnight window should be available after midnight")
	}
	if subscriptionPlanIsAvailableForSale(overnightWindow, time.Date(2026, 5, 27, 12, 0, 0, 0, time.Local)) {
		t.Fatalf("overnight window should be unavailable outside daily sale window")
	}

	manualOff := &dbent.SubscriptionPlan{
		ForSale:           false,
		DailySaleStartsAt: stringPtr("09:00"),
		DailySaleEndsAt:   stringPtr("18:00"),
	}
	if subscriptionPlanIsAvailableForSale(manualOff, time.Date(2026, 5, 27, 10, 0, 0, 0, time.Local)) {
		t.Fatalf("manual off plan should be unavailable inside daily sale window")
	}
}

func stringPtr(value string) *string {
	return &value
}

func dailyWindowAround(now time.Time, startOffset, endOffset time.Duration) (string, string) {
	return formatDailySaleTime(now.Add(startOffset)), formatDailySaleTime(now.Add(endOffset))
}

func formatDailySaleTime(value time.Time) string {
	return value.Local().Format("15:04")
}

func planNames(plans []*dbent.SubscriptionPlan) []string {
	names := make([]string, 0, len(plans))
	for _, plan := range plans {
		names = append(names, plan.Name)
	}
	return names
}

type paymentPlanSaleWindowGroupRepo struct {
	groupRepoNoop
	group *Group
}

func (r *paymentPlanSaleWindowGroupRepo) GetByID(_ context.Context, id int64) (*Group, error) {
	if r.group == nil || r.group.ID != id {
		return nil, ErrGroupNotFound
	}
	group := *r.group
	return &group, nil
}
