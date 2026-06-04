package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type leaderboardUsageRepoStub struct {
	service.UsageLogRepository
	called         bool
	startTime      time.Time
	endTime        time.Time
	limit          int
	ignoredUserIDs []int64
	result         *usagestats.UserTokenLeaderboardResponse
}

func (s *leaderboardUsageRepoStub) GetUserTokenLeaderboard(ctx context.Context, startTime, endTime time.Time, limit int, ignoredUserIDs []int64) (*usagestats.UserTokenLeaderboardResponse, error) {
	s.called = true
	s.startTime = startTime
	s.endTime = endTime
	s.limit = limit
	s.ignoredUserIDs = append([]int64(nil), ignoredUserIDs...)
	if s.result != nil {
		return s.result, nil
	}
	return &usagestats.UserTokenLeaderboardResponse{Ranking: []usagestats.UserTokenLeaderboardItem{}}, nil
}

type leaderboardSettingRepoStub struct {
	values map[string]string
}

func (s *leaderboardSettingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *leaderboardSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}

func (s *leaderboardSettingRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *leaderboardSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (s *leaderboardSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		_ = s.Set(ctx, key, value)
	}
	return nil
}

func (s *leaderboardSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *leaderboardSettingRepoStub) Delete(ctx context.Context, key string) error {
	delete(s.values, key)
	return nil
}

func newLeaderboardTestRouter(usageRepo *leaderboardUsageRepoStub, settingRepo *leaderboardSettingRepoStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	usageSvc := service.NewUsageService(usageRepo, nil, nil, nil)
	settingSvc := service.NewSettingService(settingRepo, nil)
	handler := NewUsageHandler(usageSvc, nil, settingSvc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})
		c.Next()
	})
	router.GET("/usage/leaderboard", handler.Leaderboard)
	return router
}

type leaderboardHandlerResponse struct {
	Code int `json:"code"`
	Data struct {
		Period        string `json:"period"`
		StartDate     string `json:"start_date"`
		EndDate       string `json:"end_date"`
		TotalTokens   int64  `json:"total_tokens"`
		TotalRequests int64  `json:"total_requests"`
		Ranking       []struct {
			Rank        int64   `json:"rank"`
			UserID      int64   `json:"user_id"`
			DisplayName string  `json:"display_name"`
			Tokens      int64   `json:"tokens"`
			Requests    int64   `json:"requests"`
			ActualCost  float64 `json:"actual_cost"`
			Email       string  `json:"email"`
		} `json:"ranking"`
	} `json:"data"`
}

func TestUsageLeaderboardTodayMasksIdentityAndLoadsIgnoredUsers(t *testing.T) {
	usageRepo := &leaderboardUsageRepoStub{
		result: &usagestats.UserTokenLeaderboardResponse{
			Ranking: []usagestats.UserTokenLeaderboardItem{
				{Rank: 1, UserID: 7, Email: "private@example.com", Username: "", Tokens: 1000, Requests: 5, ActualCost: 1.25},
				{Rank: 2, UserID: 8, Email: "named@example.com", Username: "Alice", Tokens: 500, Requests: 7, ActualCost: 0.75},
			},
			TotalTokens:   1500,
			TotalRequests: 12,
		},
	}
	settingRepo := &leaderboardSettingRepoStub{
		values: map[string]string{
			service.SettingKeyUsageLeaderboardSettings: `{"ignored_user_ids":[99,100]}`,
		},
	}
	router := newLeaderboardTestRouter(usageRepo, settingRepo)

	req := httptest.NewRequest(http.MethodGet, "/usage/leaderboard?period=today&limit=200&timezone=UTC", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, usageRepo.called)
	require.Equal(t, 100, usageRepo.limit)
	require.Equal(t, []int64{99, 100}, usageRepo.ignoredUserIDs)
	require.Equal(t, usageRepo.startTime.AddDate(0, 0, 1), usageRepo.endTime)

	var got leaderboardHandlerResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, "today", got.Data.Period)
	require.Equal(t, int64(1500), got.Data.TotalTokens)
	require.Len(t, got.Data.Ranking, 2)
	require.Equal(t, "p***@example.com", got.Data.Ranking[0].DisplayName)
	require.Equal(t, "A***", got.Data.Ranking[1].DisplayName)
	require.Empty(t, got.Data.Ranking[0].Email)
	require.NotContains(t, rec.Body.String(), "private@example.com")
}

func TestUsageLeaderboardLimitDefaultsAndBounds(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCode  int
		wantLimit int
		wantCall  bool
	}{
		{name: "missing limit uses default", query: "/usage/leaderboard?period=today&timezone=UTC", wantCode: http.StatusOK, wantLimit: 20, wantCall: true},
		{name: "zero limit falls back to default", query: "/usage/leaderboard?period=today&limit=0&timezone=UTC", wantCode: http.StatusOK, wantLimit: 20, wantCall: true},
		{name: "negative limit falls back to default", query: "/usage/leaderboard?period=today&limit=-3&timezone=UTC", wantCode: http.StatusOK, wantLimit: 20, wantCall: true},
		{name: "too large limit is capped", query: "/usage/leaderboard?period=today&limit=200&timezone=UTC", wantCode: http.StatusOK, wantLimit: 100, wantCall: true},
		{name: "non numeric limit is rejected", query: "/usage/leaderboard?period=today&limit=bad&timezone=UTC", wantCode: http.StatusBadRequest, wantLimit: 0, wantCall: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usageRepo := &leaderboardUsageRepoStub{}
			settingRepo := &leaderboardSettingRepoStub{values: map[string]string{}}
			router := newLeaderboardTestRouter(usageRepo, settingRepo)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, tt.wantCode, rec.Code)
			require.Equal(t, tt.wantCall, usageRepo.called)
			if tt.wantCall {
				require.Equal(t, tt.wantLimit, usageRepo.limit)
			}
		})
	}
}

func TestUsageLeaderboardYesterdayUsesPreviousDayRange(t *testing.T) {
	usageRepo := &leaderboardUsageRepoStub{}
	settingRepo := &leaderboardSettingRepoStub{values: map[string]string{}}
	router := newLeaderboardTestRouter(usageRepo, settingRepo)

	req := httptest.NewRequest(http.MethodGet, "/usage/leaderboard?period=yesterday&timezone=UTC", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, usageRepo.called)
	require.Equal(t, usageRepo.startTime.AddDate(0, 0, 1), usageRepo.endTime)
	require.True(t, usageRepo.endTime.Before(time.Now().UTC()) || usageRepo.endTime.Equal(time.Now().UTC()))
}

func TestUsageLeaderboardUsesRequestedTimezoneBoundaries(t *testing.T) {
	usageRepo := &leaderboardUsageRepoStub{}
	settingRepo := &leaderboardSettingRepoStub{values: map[string]string{}}
	router := newLeaderboardTestRouter(usageRepo, settingRepo)

	req := httptest.NewRequest(http.MethodGet, "/usage/leaderboard?period=today&timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, usageRepo.called)
	require.Equal(t, "Asia/Shanghai", usageRepo.startTime.Location().String())
	require.Equal(t, "Asia/Shanghai", usageRepo.endTime.Location().String())
	require.Equal(t, 0, usageRepo.startTime.Hour())
	require.Equal(t, 0, usageRepo.startTime.Minute())
	require.Equal(t, usageRepo.startTime.AddDate(0, 0, 1), usageRepo.endTime)
}

func TestUsageLeaderboardInvalidTimezoneFallsBackToServerTimezone(t *testing.T) {
	usageRepo := &leaderboardUsageRepoStub{}
	settingRepo := &leaderboardSettingRepoStub{values: map[string]string{}}
	router := newLeaderboardTestRouter(usageRepo, settingRepo)

	req := httptest.NewRequest(http.MethodGet, "/usage/leaderboard?period=today&timezone=Invalid/Zone", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, usageRepo.called)
	require.Equal(t, time.Local.String(), usageRepo.startTime.Location().String())
	require.Equal(t, usageRepo.startTime.AddDate(0, 0, 1), usageRepo.endTime)
}

func TestUsageLeaderboardRejectsInvalidPeriod(t *testing.T) {
	usageRepo := &leaderboardUsageRepoStub{}
	settingRepo := &leaderboardSettingRepoStub{values: map[string]string{}}
	router := newLeaderboardTestRouter(usageRepo, settingRepo)

	req := httptest.NewRequest(http.MethodGet, "/usage/leaderboard?period=week", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.False(t, usageRepo.called)
}
