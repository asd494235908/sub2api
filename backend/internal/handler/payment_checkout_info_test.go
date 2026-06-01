package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestPaymentHandlerGetCheckoutInfoIncludesDailySaleMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	client := newCheckoutInfoHandlerTestClient(t)
	repo := &checkoutInfoSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:                    "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	handler := NewPaymentHandler(nil, configService, nil)

	_, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("EasyPay").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Group.Create().
		SetName("Subscription Group").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now()
	start := now.Add(time.Hour).Local().Format("15:04")
	end := now.Add(2 * time.Hour).Local().Format("15:04")
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(1).
		SetName("Daily Window Plan").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetDailyPurchaseLimit(2).
		SetDailySaleStartsAt(start).
		SetDailySaleEndsAt(end).
		Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("checkout@example.com").
		SetPasswordHash("hash").
		SetUsername("checkout-user").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(10).
		SetPayAmount(10).
		SetFeeRate(0).
		SetRechargeCode("PAY-checkout").
		SetOutTradeNo("sub2_checkout_daily_1").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(service.OrderStatusPending).
		SetPlanID(plan.ID).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("localhost").
		Save(ctx)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(rec)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/checkout-info", nil)
	gctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: user.ID, Concurrency: 1})

	handler.GetCheckoutInfo(gctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Plans []struct {
				Name                      string `json:"name"`
				DailySaleStartsAt         string `json:"daily_sale_starts_at"`
				DailySaleEndsAt           string `json:"daily_sale_ends_at"`
				DailyPurchaseLimit        int    `json:"daily_purchase_limit"`
				DailyPurchaseRemaining    *int   `json:"daily_purchase_remaining"`
				DailySaleStatus           string `json:"daily_sale_status"`
				DailySaleCountdownSeconds int    `json:"daily_sale_countdown_seconds"`
			} `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Plans, 1)
	got := body.Data.Plans[0]
	require.Equal(t, "Daily Window Plan", got.Name)
	require.Equal(t, start, got.DailySaleStartsAt)
	require.Equal(t, end, got.DailySaleEndsAt)
	require.Equal(t, 2, got.DailyPurchaseLimit)
	require.NotNil(t, got.DailyPurchaseRemaining)
	require.Equal(t, 1, *got.DailyPurchaseRemaining)
	require.Equal(t, "pending", got.DailySaleStatus)
	require.Greater(t, got.DailySaleCountdownSeconds, 0)
}

func TestPaymentHandlerGetCheckoutInfoIncludesPurchaseOnceAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	client := newCheckoutInfoHandlerTestClient(t)
	repo := &checkoutInfoSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:                    "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	handler := NewPaymentHandler(nil, configService, nil)

	_, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("EasyPay").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("Subscription Group").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().
		SetEmail("checkout-once@example.com").
		SetPasswordHash("hash").
		SetUsername("checkout-once-user").
		Save(ctx)
	require.NoError(t, err)
	expiresAt := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	_, err = client.UserSubscription.Create().
		SetUserID(user.ID).
		SetGroupID(int64(group.ID)).
		SetStartsAt(time.Now().Add(-time.Hour)).
		SetExpiresAt(expiresAt).
		SetStatus(service.SubscriptionStatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(int64(group.ID)).
		SetName("Once Plan").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetPurchaseOncePerActiveSubscription(true).
		Save(ctx)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(rec)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/checkout-info", nil)
	gctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: user.ID, Concurrency: 1})

	handler.GetCheckoutInfo(gctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Plans []struct {
				Name                              string  `json:"name"`
				PurchaseOncePerActiveSubscription bool    `json:"purchase_once_per_active_subscription"`
				PurchaseOnceAvailableForPayment   bool    `json:"purchase_once_available_for_payment"`
				PurchaseOnceUnavailableUntil      *string `json:"purchase_once_unavailable_until"`
			} `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Plans, 1)
	got := body.Data.Plans[0]
	require.Equal(t, "Once Plan", got.Name)
	require.True(t, got.PurchaseOncePerActiveSubscription)
	require.False(t, got.PurchaseOnceAvailableForPayment)
	require.NotNil(t, got.PurchaseOnceUnavailableUntil)
	require.Equal(t, expiresAt.Format(time.RFC3339), *got.PurchaseOnceUnavailableUntil)
}

func TestPaymentHandlerGetCheckoutInfoIncludesWeeklySaleAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	client := newCheckoutInfoHandlerTestClient(t)
	repo := &checkoutInfoSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:                    "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	handler := NewPaymentHandler(nil, configService, nil)

	_, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("EasyPay").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("Subscription Group").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	today := handlerWeekdayNumber(time.Now())
	offDay := today%7 + 1
	_, err = client.SubscriptionPlan.Create().
		SetGroupID(int64(group.ID)).
		SetName("Weekly Off Day Plan").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("days").
		SetForSale(true).
		SetWeeklySaleDays([]int{offDay}).
		Save(ctx)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(rec)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/checkout-info", nil)
	gctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})

	handler.GetCheckoutInfo(gctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Plans []struct {
				Name                          string `json:"name"`
				WeeklySaleDays                []int  `json:"weekly_sale_days"`
				WeeklySaleStatus              string `json:"weekly_sale_status"`
				WeeklySaleAvailableForPayment bool   `json:"weekly_sale_available_for_payment"`
				DailySaleAvailableForPayment  bool   `json:"daily_sale_available_for_payment"`
			} `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Plans, 1)
	got := body.Data.Plans[0]
	require.Equal(t, "Weekly Off Day Plan", got.Name)
	require.Equal(t, []int{offDay}, got.WeeklySaleDays)
	require.Equal(t, "off_day", got.WeeklySaleStatus)
	require.False(t, got.WeeklySaleAvailableForPayment)
	require.False(t, got.DailySaleAvailableForPayment)
}

func TestPaymentHandlerGetCheckoutInfoIncludesMappedSupportedModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	client := newCheckoutInfoHandlerTestClient(t)
	repo := &checkoutInfoSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:                    "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	_, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("EasyPay").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("OpenAI Subscription Group").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	groupID := int64(group.ID)
	gatewayService := service.NewGatewayService(
		&checkoutInfoAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       10,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4":       "gpt-5.4",
								"gpt-5.4-mini":  "gpt-5.4-mini",
								"gpt-5.3-codex": "gpt-5.3-codex",
							},
						},
					},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := NewPaymentHandler(nil, configService, nil)
	handler.gatewayService = gatewayService

	_, err = client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetName("OpenAI Plan").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("day").
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(rec)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/checkout-info", nil)
	gctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})

	handler.GetCheckoutInfo(gctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Plans []struct {
				Name            string   `json:"name"`
				SupportedModels []string `json:"supported_models"`
			} `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Plans, 1)
	require.Equal(t, "OpenAI Plan", body.Data.Plans[0].Name)
	require.Equal(t, []string{"gpt-5.3-codex", "gpt-5.4", "gpt-5.4-mini"}, body.Data.Plans[0].SupportedModels)
}

func TestPaymentHandlerGetCheckoutInfoFallsBackToDefaultSupportedModels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	client := newCheckoutInfoHandlerTestClient(t)
	repo := &checkoutInfoSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:                    "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  service.VisibleMethodSourceEasyPayAlipay,
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	_, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName("EasyPay").
		SetConfig("{}").
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().
		SetName("OpenAI Default Group").
		SetStatus(payment.EntityStatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeSubscription).
		Save(ctx)
	require.NoError(t, err)
	groupID := int64(group.ID)
	gatewayService := service.NewGatewayService(
		&checkoutInfoAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{ID: 20, Platform: service.PlatformOpenAI},
				},
			},
		},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	handler := NewPaymentHandler(nil, configService, nil)
	handler.gatewayService = gatewayService

	_, err = client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetName("Default Models Plan").
		SetPrice(10).
		SetValidityDays(30).
		SetValidityUnit("day").
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(rec)
	gctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/checkout-info", nil)
	gctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})

	handler.GetCheckoutInfo(gctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Plans []struct {
				SupportedModels []string `json:"supported_models"`
			} `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Plans, 1)
	require.Contains(t, body.Data.Plans[0].SupportedModels, "gpt-5.4")
}

func newCheckoutInfoHandlerTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	dbName := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func handlerWeekdayNumber(value time.Time) int {
	weekday := int(value.Local().Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

type checkoutInfoSettingRepoStub struct {
	values map[string]string
}

func (s *checkoutInfoSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, nil
}

func (s *checkoutInfoSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *checkoutInfoSettingRepoStub) Set(context.Context, string, string) error { return nil }

func (s *checkoutInfoSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = s.values[key]
	}
	return out, nil
}

func (s *checkoutInfoSettingRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		if s.values == nil {
			s.values = map[string]string{}
		}
		s.values[key] = value
	}
	return nil
}

func (s *checkoutInfoSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

type checkoutInfoAccountRepoStub struct {
	service.AccountRepository

	byGroup map[int64][]service.Account
}

func (s *checkoutInfoAccountRepoStub) ListSchedulableByGroupID(_ context.Context, groupID int64) ([]service.Account, error) {
	accounts := s.byGroup[groupID]
	out := make([]service.Account, len(accounts))
	copy(out, accounts)
	return out, nil
}

func (s *checkoutInfoSettingRepoStub) Delete(context.Context, string) error { return nil }
