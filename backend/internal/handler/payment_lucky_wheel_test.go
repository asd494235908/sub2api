//go:build unit

package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

type luckyWheelHandlerUserRepoStub struct {
	balance float64
}

type luckyWheelHandlerSettingRepoStub struct {
	values map[string]string
}

func (s *luckyWheelHandlerSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get")
}
func (s *luckyWheelHandlerSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}
func (s *luckyWheelHandlerSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set")
}
func (s *luckyWheelHandlerSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}
func (s *luckyWheelHandlerSettingRepoStub) SetMultiple(_ context.Context, settings map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range settings {
		s.values[key] = value
	}
	return nil
}
func (s *luckyWheelHandlerSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll")
}
func (s *luckyWheelHandlerSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete")
}

func (s *luckyWheelHandlerUserRepoStub) Create(context.Context, *service.User) error {
	panic("unexpected Create")
}
func (s *luckyWheelHandlerUserRepoStub) GetByID(context.Context, int64) (*service.User, error) {
	return &service.User{ID: 1, Balance: s.balance}, nil
}
func (s *luckyWheelHandlerUserRepoStub) GetByEmail(context.Context, string) (*service.User, error) {
	panic("unexpected GetByEmail")
}
func (s *luckyWheelHandlerUserRepoStub) GetByPhone(context.Context, string) (*service.User, error) {
	panic("unexpected GetByPhone")
}
func (s *luckyWheelHandlerUserRepoStub) GetFirstAdmin(context.Context) (*service.User, error) {
	panic("unexpected GetFirstAdmin")
}
func (s *luckyWheelHandlerUserRepoStub) Update(context.Context, *service.User) error {
	panic("unexpected Update")
}
func (s *luckyWheelHandlerUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete")
}
func (s *luckyWheelHandlerUserRepoStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	panic("unexpected GetUserAvatar")
}
func (s *luckyWheelHandlerUserRepoStub) UpsertUserAvatar(context.Context, int64, service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	panic("unexpected UpsertUserAvatar")
}
func (s *luckyWheelHandlerUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar")
}
func (s *luckyWheelHandlerUserRepoStub) List(context.Context, pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected List")
}
func (s *luckyWheelHandlerUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters")
}
func (s *luckyWheelHandlerUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs")
}
func (s *luckyWheelHandlerUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID")
}
func (s *luckyWheelHandlerUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt")
}
func (s *luckyWheelHandlerUserRepoStub) UpdateBalance(_ context.Context, _ int64, amount float64) error {
	s.balance += amount
	return nil
}
func (s *luckyWheelHandlerUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance")
}
func (s *luckyWheelHandlerUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency")
}
func (s *luckyWheelHandlerUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail")
}
func (s *luckyWheelHandlerUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups")
}
func (s *luckyWheelHandlerUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups")
}
func (s *luckyWheelHandlerUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups")
}
func (s *luckyWheelHandlerUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]service.UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities")
}
func (s *luckyWheelHandlerUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider")
}
func (s *luckyWheelHandlerUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret")
}
func (s *luckyWheelHandlerUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp")
}
func (s *luckyWheelHandlerUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp")
}

func newLuckyWheelHandlerTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", "file:payment_handler_lucky_wheel?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func newLuckyWheelPaymentHandler(t *testing.T) (*PaymentHandler, *service.PaymentService, *service.PaymentConfigService, *luckyWheelHandlerSettingRepoStub) {
	t.Helper()
	client := newLuckyWheelHandlerTestClient(t)
	repo := &luckyWheelHandlerSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:       "true",
			service.SettingKeyLuckyWheelEnabled: "true",
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	userRepo := &luckyWheelHandlerUserRepoStub{}
	paymentService := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, configService, userRepo, nil, nil)
	handler := NewPaymentHandler(paymentService, configService, nil)
	return handler, paymentService, configService, repo
}

func luckyWheelAuthContext(rec *httptest.ResponseRecorder, method, path string, body []byte) *gin.Context {
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})
	return ctx
}

func handlerDefaultLuckyWheelConfig() *service.LuckyWheelConfig {
	return &service.LuckyWheelConfig{
		EligibleOrderTypes:  []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		MultiplierStep:      0.1,
		GlobalMaxMultiplier: 3.0,
		AmountTiers: []service.LuckyWheelAmountTier{
			{ID: "tier_20_50", Name: "20-50", MinAmount: 20, MaxAmount: luckyWheelHandlerPtrFloat64(50), MinMultiplier: 1.1, MaxMultiplier: 1.1, DrawCount: 2},
			{ID: "tier_51_plus", Name: "51+", MinAmount: 51, MinMultiplier: 1.2, MaxMultiplier: 1.2, DrawCount: 3},
		},
		InviteBonus: service.LuckyWheelInviteBonusConfig{
			Enabled:          false,
			QualifyingAmount: 20,
			BonusPerInvitee:  0.2,
			MaxBonus:         1.0,
			ConsumePolicy:    service.LuckyWheelInviteBonusConsumeNextSessionOnce,
		},
		GoldenWindow: service.LuckyWheelGoldenWindowConfig{
			Enabled:    false,
			Timezone:   "Asia/Shanghai",
			StartTime:  "20:00",
			EndTime:    "22:00",
			MinAmount:  51,
			ExtraDraws: 1,
			DailyQuota: 5,
		},
	}
}

func luckyWheelHandlerPtrFloat64(v float64) *float64 {
	return &v
}

func saveLuckyWheelHandlerConfig(t *testing.T, repo *luckyWheelHandlerSettingRepoStub, cfg *service.LuckyWheelConfig) {
	t.Helper()
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[service.SettingKeyLuckyWheelConfig] = string(raw)
}

func TestPaymentHandler_GetLuckyWheelSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, paymentService, _, repo := newLuckyWheelPaymentHandler(t)
	saveLuckyWheelHandlerConfig(t, repo, handlerDefaultLuckyWheelConfig())

	order := &dbent.PaymentOrder{ID: 10, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 88}
	require.NoError(t, paymentService.GrantLuckyWheelChanceForOrder(context.Background(), order))

	rec := httptest.NewRecorder()
	ctx := luckyWheelAuthContext(rec, http.MethodGet, "/api/v1/payment/lucky-wheel", nil)
	handler.GetLuckyWheelSummary(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Enabled       bool `json:"enabled"`
			ActiveSession struct {
				ID         int64 `json:"id"`
				TotalDraws int   `json:"total_draws"`
			} `json:"active_session"`
			PendingSessions []map[string]any `json:"pending_sessions"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.True(t, resp.Data.Enabled)
	require.Equal(t, 1, len(resp.Data.PendingSessions))
	require.Equal(t, 3, resp.Data.ActiveSession.TotalDraws)
}

func TestPaymentHandler_DrawLuckyWheel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, paymentService, _, repo := newLuckyWheelPaymentHandler(t)
	saveLuckyWheelHandlerConfig(t, repo, handlerDefaultLuckyWheelConfig())

	order := &dbent.PaymentOrder{ID: 11, UserID: 1, OrderType: payment.OrderTypeSubscription, PayAmount: 50}
	require.NoError(t, paymentService.GrantLuckyWheelChanceForOrder(context.Background(), order))

	summary, err := paymentService.GetLuckyWheelSummary(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)

	body, err := json.Marshal(map[string]any{"session_id": summary.ActiveSession.ID})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx := luckyWheelAuthContext(rec, http.MethodPost, "/api/v1/payment/lucky-wheel/draw", body)
	handler.DrawLuckyWheel(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			SessionID      int64   `json:"session_id"`
			BestMultiplier float64 `json:"best_multiplier"`
			RemainingDraws int     `json:"remaining_draws"`
			DrawRecord     struct {
				DrawIndex       int     `json:"draw_index"`
				FinalMultiplier float64 `json:"final_multiplier"`
			} `json:"draw_record"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, summary.ActiveSession.ID, resp.Data.SessionID)
	require.Equal(t, 1, resp.Data.DrawRecord.DrawIndex)
	require.InDelta(t, 1.1, resp.Data.DrawRecord.FinalMultiplier, 1e-9)
	require.Equal(t, 1, resp.Data.RemainingDraws)
}

func TestPaymentHandler_DrawLuckyWheelRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, _ := newLuckyWheelPaymentHandler(t)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/lucky-wheel/draw", nil)

	handler.DrawLuckyWheel(ctx)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPaymentHandler_DrawLuckyWheelReturnsBadRequestWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, _, repo := newLuckyWheelPaymentHandler(t)
	repo.values[service.SettingKeyLuckyWheelEnabled] = "false"
	saveLuckyWheelHandlerConfig(t, repo, handlerDefaultLuckyWheelConfig())

	rec := httptest.NewRecorder()
	ctx := luckyWheelAuthContext(rec, http.MethodPost, "/api/v1/payment/lucky-wheel/draw", []byte(`{"session_id":1}`))

	handler.DrawLuckyWheel(ctx)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAdminPaymentHandler_UpdateLuckyWheelConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newLuckyWheelPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)

	body := map[string]any{
		"enabled": true,
		"config": map[string]any{
			"eligible_order_types":  []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
			"multiplier_step":       0.1,
			"global_max_multiplier": 3.0,
			"amount_tiers": []map[string]any{
				{"id": "tier_20_50", "name": "20-50", "min_amount": 20, "max_amount": 50, "min_multiplier": 1.1, "max_multiplier": 2.0, "draw_count": 2},
				{"id": "tier_51_plus", "name": "51+", "min_amount": 51, "min_multiplier": 1.2, "max_multiplier": 3.0, "draw_count": 3},
			},
			"invite_bonus": map[string]any{
				"enabled": true, "qualifying_amount": 20, "bonus_per_invitee": 0.2, "max_bonus": 1.0, "consume_policy": service.LuckyWheelInviteBonusConsumeNextSessionOnce,
			},
			"golden_window": map[string]any{
				"enabled": true, "timezone": "Asia/Shanghai", "start_time": "20:00", "end_time": "22:00", "min_amount": 51, "extra_draws": 1, "daily_quota": 5,
			},
		},
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/lucky-wheel/config", bytes.NewReader(raw))
	ctx.Request.Header.Set("Content-Type", "application/json")

	adminHandler.UpdateLuckyWheelConfig(ctx)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, repo.values[service.SettingKeyLuckyWheelConfig])
}

func TestAdminPaymentHandler_GetLuckyWheelStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newLuckyWheelPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveLuckyWheelHandlerConfig(t, repo, handlerDefaultLuckyWheelConfig())

	order := &dbent.PaymentOrder{ID: 22, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 50}
	require.NoError(t, paymentService.GrantLuckyWheelChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetLuckyWheelSummary(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)
	_, err = paymentService.DrawLuckyWheel(context.Background(), 1, summary.ActiveSession.ID)
	require.NoError(t, err)
	_, err = paymentService.DrawLuckyWheel(context.Background(), 1, summary.ActiveSession.ID)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/lucky-wheel/stats", nil)

	adminHandler.GetLuckyWheelStats(ctx)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			TotalSessions   int   `json:"total_sessions"`
			SettledSessions int   `json:"settled_sessions"`
			RecentSessions  []any `json:"recent_sessions"`
			MultiplierStats []any `json:"multiplier_stats"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Data.TotalSessions)
	require.Equal(t, 1, resp.Data.SettledSessions)
	require.NotEmpty(t, resp.Data.RecentSessions)
	require.NotEmpty(t, resp.Data.MultiplierStats)
}
