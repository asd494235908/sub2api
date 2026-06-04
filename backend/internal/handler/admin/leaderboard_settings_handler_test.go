package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newLeaderboardSettingsHandler(repo *settingHandlerRepoStub) *SettingHandler {
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	return NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)
}

func newLeaderboardSettingsHandlerWithAdmin(repo *settingHandlerRepoStub, adminService service.AdminService) *SettingHandler {
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	return NewSettingHandler(svc, nil, nil, nil, nil, nil, adminService)
}

func TestSettingHandlerLeaderboardSettingsRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{values: map[string]string{}}
	handler := newLeaderboardSettingsHandler(repo)

	rawBody := []byte(`{"ignored_user_ids":[7,8,7,0,-1]}`)
	putRec := httptest.NewRecorder()
	putCtx, _ := gin.CreateTestContext(putRec)
	putCtx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/leaderboard/settings", bytes.NewReader(rawBody))
	putCtx.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateLeaderboardSettings(putCtx)

	require.Equal(t, http.StatusOK, putRec.Code)
	require.JSONEq(t, `{"ignored_user_ids":[7,8]}`, repo.values[service.SettingKeyUsageLeaderboardSettings])

	getRec := httptest.NewRecorder()
	getCtx, _ := gin.CreateTestContext(getRec)
	getCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/leaderboard/settings", nil)

	handler.GetLeaderboardSettings(getCtx)

	require.Equal(t, http.StatusOK, getRec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{float64(7), float64(8)}, data["ignored_user_ids"])
}

func TestSettingHandlerLeaderboardSettingsRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{values: map[string]string{}}
	handler := newLeaderboardSettingsHandler(repo)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/leaderboard/settings", bytes.NewReader([]byte(`{"ignored_user_ids":"bad"}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateLeaderboardSettings(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSettingHandlerLeaderboardSettingsKeepsMissingIDsAndReturnsExistingUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyUsageLeaderboardSettings: `{"ignored_user_ids":[1,404,2]}`,
		},
	}
	adminSvc := newStubAdminService()
	adminSvc.users = []service.User{
		{ID: 1, Email: "one@example.com", Username: "one", Status: service.StatusActive},
		{ID: 2, Email: "two@example.com", Username: "two", Status: service.StatusActive},
	}
	adminSvc.missingUserIDs = map[int64]bool{404: true}
	handler := newLeaderboardSettingsHandlerWithAdmin(repo, adminSvc)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/leaderboard/settings", nil)

	handler.GetLeaderboardSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{float64(1), float64(2), float64(404)}, data["ignored_user_ids"])
	require.Equal(t, []any{
		map[string]any{"id": float64(1), "email": "one@example.com", "username": "one"},
		map[string]any{"id": float64(2), "email": "two@example.com", "username": "two"},
	}, data["ignored_users"])
}
