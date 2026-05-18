//go:build unit

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newRechargeActivityPaymentHandler(t *testing.T) (*PaymentHandler, *service.PaymentService, *service.PaymentConfigService, *luckyWheelHandlerSettingRepoStub) {
	t.Helper()
	client := newLuckyWheelHandlerTestClient(t)
	repo := &luckyWheelHandlerSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:             "true",
			service.SettingKeyRechargeActivityEnabled: "true",
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	userRepo := &luckyWheelHandlerUserRepoStub{}
	paymentService := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, configService, userRepo, nil, nil)
	handler := NewPaymentHandler(paymentService, configService, nil)
	return handler, paymentService, configService, repo
}

func rechargeActivityAuthContext(rec *httptest.ResponseRecorder, method, path string, body []byte) *gin.Context {
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})
	return ctx
}

func handlerDefaultRechargeActivityConfig() *service.RechargeActivityConfig {
	return &service.RechargeActivityConfig{
		EligibleOrderTypes: []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		Prizes: []service.RechargeActivityPrize{
			{ID: "third", Name: "三等奖", RewardDescription: "联系客服领取实体礼品", Probability: 70, MinPayAmount: 20, Enabled: true, SortOrder: 3},
			{ID: "second", Name: "二等奖", RewardDescription: "赠送站外会员 30 天", Probability: 20, MinPayAmount: 50, Enabled: true, SortOrder: 2},
			{ID: "first", Name: "一等奖", RewardDescription: "人工发放定制奖励", Probability: 10, MinPayAmount: 100, Enabled: true, SortOrder: 1},
		},
	}
}

func saveRechargeActivityHandlerConfig(t *testing.T, repo *luckyWheelHandlerSettingRepoStub, cfg *service.RechargeActivityConfig) {
	t.Helper()
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[service.SettingKeyRechargeActivityConfig] = string(raw)
}

func TestPaymentHandler_GetRechargeActivitySummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, paymentService, _, repo := newRechargeActivityPaymentHandler(t)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 10, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))

	rec := httptest.NewRecorder()
	ctx := rechargeActivityAuthContext(rec, http.MethodGet, "/api/v1/payment/recharge-activity", nil)
	handler.GetRechargeActivitySummary(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Enabled        bool                             `json:"enabled"`
			PendingChances []service.RechargeActivityChance `json:"pending_chances"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Data.Enabled)
	require.Len(t, resp.Data.PendingChances, 1)
}

func TestPaymentHandler_DrawRechargeActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, paymentService, _, repo := newRechargeActivityPaymentHandler(t)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 11, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetRechargeActivitySummary(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, summary.PendingChances, 1)

	body, err := json.Marshal(map[string]any{"chance_id": summary.PendingChances[0].ID})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	ctx := rechargeActivityAuthContext(rec, http.MethodPost, "/api/v1/payment/recharge-activity/draw", body)
	handler.DrawRechargeActivity(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "三等奖")
	require.Contains(t, rec.Body.String(), "联系客服领取实体礼品")
	require.Contains(t, rec.Body.String(), service.RechargeActivityFulfillmentPending)
}

func TestAdminPaymentHandler_UpdateRechargeActivityConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)

	body := map[string]any{
		"enabled": true,
		"config": map[string]any{
			"eligible_order_types": []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
			"prizes": []map[string]any{
				{"id": "third", "name": "三等奖", "reward_description": "联系客服领取实体礼品", "probability": 70, "min_pay_amount": 20, "enabled": true, "sort_order": 3},
				{"id": "second", "name": "二等奖", "reward_description": "赠送站外会员 30 天", "probability": 20, "min_pay_amount": 50, "enabled": true, "sort_order": 2},
				{"id": "first", "name": "一等奖", "reward_description": "人工发放定制奖励", "probability": 10, "min_pay_amount": 100, "enabled": true, "sort_order": 1},
			},
		},
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/recharge-activity/config", bytes.NewReader(raw))
	ctx.Request.Header.Set("Content-Type", "application/json")

	adminHandler.UpdateRechargeActivityConfig(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, repo.values[service.SettingKeyRechargeActivityConfig])
}

func TestAdminPaymentHandler_GetRechargeActivityStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 12, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetRechargeActivitySummary(context.Background(), 1)
	require.NoError(t, err)
	_, err = paymentService.DrawRechargeActivity(context.Background(), 1, summary.PendingChances[0].ID)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/recharge-activity/stats", nil)

	adminHandler.GetRechargeActivityStats(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "pending_fulfillments")
	require.Contains(t, rec.Body.String(), `"recent_records_total":1`)
	require.Contains(t, rec.Body.String(), `"recent_records_page":1`)
	require.Contains(t, rec.Body.String(), `"recent_records_page_size":20`)
}

func TestAdminPaymentHandler_GetRechargeActivityStatsAcceptsPaginationAndSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 15, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetRechargeActivitySummary(context.Background(), 1)
	require.NoError(t, err)
	_, err = paymentService.DrawRechargeActivity(context.Background(), 1, summary.PendingChances[0].ID)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/recharge-activity/stats?page=2&page_size=1&user_keyword=admin", nil)

	adminHandler.GetRechargeActivityStats(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"recent_records_page":2`)
	require.Contains(t, rec.Body.String(), `"recent_records_page_size":1`)
	require.Contains(t, rec.Body.String(), `"recent_records_keyword":"admin"`)
}

func TestAdminPaymentHandler_GetRechargeActivityStatsNormalizesInvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/recharge-activity/stats?page=-1&page_size=1000", nil)

	adminHandler.GetRechargeActivityStats(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"recent_records_page":1`)
	require.Contains(t, rec.Body.String(), `"recent_records_page_size":100`)
}

func TestAdminPaymentHandler_UpdateRechargeActivityRecordFulfillment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 13, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetRechargeActivitySummary(context.Background(), 1)
	require.NoError(t, err)
	result, err := paymentService.DrawRechargeActivity(context.Background(), 1, summary.PendingChances[0].ID)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{"status": service.RechargeActivityFulfillmentFulfilled, "note": "已人工发放"})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/recharge-activity/records/1/fulfillment", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(result.Record.ID, 10)}}
	ctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 99})

	adminHandler.UpdateRechargeActivityRecordFulfillment(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), service.RechargeActivityFulfillmentFulfilled)
	require.Contains(t, rec.Body.String(), "已人工发放")
}

func TestAdminPaymentHandler_UpdateRechargeActivityRecordFulfillmentRequiresNoteWhenFulfilled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, paymentService, configService, repo := newRechargeActivityPaymentHandler(t)
	adminHandler := adminhandler.NewPaymentHandler(paymentService, configService)
	saveRechargeActivityHandlerConfig(t, repo, handlerDefaultRechargeActivityConfig())

	order := &dbent.PaymentOrder{ID: 14, UserID: 1, OrderType: payment.OrderTypeBalance, PayAmount: 20}
	require.NoError(t, paymentService.GrantRechargeActivityChanceForOrder(context.Background(), order))
	summary, err := paymentService.GetRechargeActivitySummary(context.Background(), 1)
	require.NoError(t, err)
	result, err := paymentService.DrawRechargeActivity(context.Background(), 1, summary.PendingChances[0].ID)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{"status": service.RechargeActivityFulfillmentFulfilled, "note": "   "})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/recharge-activity/records/1/fulfillment", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(result.Record.ID, 10)}}
	ctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 99})

	adminHandler.UpdateRechargeActivityRecordFulfillment(ctx)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "RECHARGE_ACTIVITY_FULFILLMENT_NOTE_REQUIRED")
}
