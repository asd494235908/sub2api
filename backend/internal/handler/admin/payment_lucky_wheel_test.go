//go:build unit

package admin

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
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type adminLuckyWheelSettingRepoStub struct {
	values map[string]string
}

func (s *adminLuckyWheelSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get")
}
func (s *adminLuckyWheelSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}
func (s *adminLuckyWheelSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set")
}
func (s *adminLuckyWheelSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}
func (s *adminLuckyWheelSettingRepoStub) SetMultiple(_ context.Context, settings map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range settings {
		s.values[key] = value
	}
	return nil
}
func (s *adminLuckyWheelSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll")
}
func (s *adminLuckyWheelSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete")
}

type adminLuckyWheelUserRepoStub struct{}

func (s *adminLuckyWheelUserRepoStub) BatchUpdateLimits(context.Context, []int64, *int, *int) (int, error) {
	panic("unexpected BatchUpdateLimits")
}

func (s *adminLuckyWheelUserRepoStub) Create(context.Context, *service.User) error {
	panic("unexpected Create")
}
func (s *adminLuckyWheelUserRepoStub) GetByID(context.Context, int64) (*service.User, error) {
	return &service.User{ID: 1, Balance: 0}, nil
}
func (s *adminLuckyWheelUserRepoStub) GetByIDIncludeDeleted(ctx context.Context, id int64) (*service.User, error) {
	return s.GetByID(ctx, id)
}
func (s *adminLuckyWheelUserRepoStub) GetByEmail(context.Context, string) (*service.User, error) {
	panic("unexpected GetByEmail")
}
func (s *adminLuckyWheelUserRepoStub) GetByPhone(context.Context, string) (*service.User, error) {
	panic("unexpected GetByPhone")
}
func (s *adminLuckyWheelUserRepoStub) GetFirstAdmin(context.Context) (*service.User, error) {
	panic("unexpected GetFirstAdmin")
}
func (s *adminLuckyWheelUserRepoStub) Update(context.Context, *service.User) error {
	panic("unexpected Update")
}
func (s *adminLuckyWheelUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete")
}
func (s *adminLuckyWheelUserRepoStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	panic("unexpected GetUserAvatar")
}
func (s *adminLuckyWheelUserRepoStub) UpsertUserAvatar(context.Context, int64, service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	panic("unexpected UpsertUserAvatar")
}
func (s *adminLuckyWheelUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar")
}
func (s *adminLuckyWheelUserRepoStub) List(context.Context, pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected List")
}
func (s *adminLuckyWheelUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters")
}
func (s *adminLuckyWheelUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs")
}
func (s *adminLuckyWheelUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID")
}
func (s *adminLuckyWheelUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt")
}
func (s *adminLuckyWheelUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	return nil
}
func (s *adminLuckyWheelUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance")
}
func (s *adminLuckyWheelUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency")
}
func (s *adminLuckyWheelUserRepoStub) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchSetConcurrency")
}
func (s *adminLuckyWheelUserRepoStub) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchAddConcurrency")
}
func (s *adminLuckyWheelUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail")
}
func (s *adminLuckyWheelUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups")
}
func (s *adminLuckyWheelUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups")
}
func (s *adminLuckyWheelUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups")
}
func (s *adminLuckyWheelUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]service.UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities")
}
func (s *adminLuckyWheelUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider")
}
func (s *adminLuckyWheelUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret")
}
func (s *adminLuckyWheelUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp")
}
func (s *adminLuckyWheelUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp")
}

func newAdminLuckyWheelTestHandler(t *testing.T) (*PaymentHandler, *adminLuckyWheelSettingRepoStub) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:admin_lucky_wheel?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	repo := &adminLuckyWheelSettingRepoStub{
		values: map[string]string{
			service.SettingPaymentEnabled:       "true",
			service.SettingKeyLuckyWheelEnabled: "false",
		},
	}
	configService := service.NewPaymentConfigService(client, repo, nil)
	paymentService := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, configService, &adminLuckyWheelUserRepoStub{}, nil, nil)
	return NewPaymentHandler(paymentService, configService), repo
}

func TestUpdateLuckyWheelConfigRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &PaymentHandler{}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/lucky-wheel/config", bytes.NewBufferString(`{bad-json`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateLuckyWheelConfig(ctx)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetLuckyWheelStatsHandlesMissingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &PaymentHandler{}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/lucky-wheel/stats", nil)

	handler.GetLuckyWheelStats(ctx)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetLuckyWheelConfigHandlesMissingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &PaymentHandler{}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/lucky-wheel/config", nil)

	handler.GetLuckyWheelConfig(ctx)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateLuckyWheelConfigValidationErrorReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := map[string]any{
		"enabled": true,
		"config": map[string]any{
			"eligible_order_types":  []string{payment.OrderTypeBalance},
			"multiplier_step":       0,
			"global_max_multiplier": 3.0,
			"amount_tiers": []map[string]any{
				{"id": "tier_20_50", "name": "20-50", "min_amount": 20, "max_amount": 50, "min_multiplier": 1.1, "max_multiplier": 2.0, "draw_count": 2},
			},
		},
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/payment/lucky-wheel/config", bytes.NewReader(raw))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler, _ := newAdminLuckyWheelTestHandler(t)
	handler.UpdateLuckyWheelConfig(ctx)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLuckyWheelConfigBusinessErrorShape(t *testing.T) {
	err := infraerrors.BadRequest("LUCKY_WHEEL_STEP_INVALID", "multiplier step must be greater than 0")
	statusCode, body := infraerrors.ToHTTP(err)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Equal(t, int32(http.StatusBadRequest), body.Code)
}

func TestUpdateLuckyWheelConfigRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := newAdminLuckyWheelTestHandler(t)

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
	handler.UpdateLuckyWheelConfig(ctx)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/payment/lucky-wheel/config", nil)
	handler.GetLuckyWheelConfig(ctx)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Enabled bool `json:"enabled"`
			Config  struct {
				EligibleOrderTypes []string `json:"eligible_order_types"`
				AmountTiers        []struct {
					ID string `json:"id"`
				} `json:"amount_tiers"`
			} `json:"config"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.True(t, resp.Data.Enabled)
	require.ElementsMatch(t, []string{payment.OrderTypeBalance, payment.OrderTypeSubscription}, resp.Data.Config.EligibleOrderTypes)
	require.Len(t, resp.Data.Config.AmountTiers, 2)
	require.Equal(t, "tier_20_50", resp.Data.Config.AmountTiers[0].ID)
	require.Equal(t, "tier_51_plus", resp.Data.Config.AmountTiers[1].ID)
}
