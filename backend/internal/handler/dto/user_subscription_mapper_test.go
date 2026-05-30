package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionFromServiceIncludesInstanceTotalLimit(t *testing.T) {
	t.Parallel()

	groupLimit := 1560.0
	instanceLimit := 3120.0
	sub := &service.UserSubscription{
		ID:              47,
		UserID:          2,
		GroupID:         34,
		StartsAt:        time.Date(2026, 6, 26, 23, 57, 2, 0, time.UTC),
		ExpiresAt:       time.Date(2028, 6, 26, 23, 57, 2, 0, time.UTC),
		Status:          service.SubscriptionStatusActive,
		TotalLimitUSD:   &instanceLimit,
		TotalUsageUSD:   260,
		DailyUsageUSD:   1,
		WeeklyUsageUSD:  2,
		MonthlyUsageUSD: 3,
		Group: &service.Group{
			ID:                        34,
			Name:                      "Early Bird",
			SubscriptionTotalLimitUSD: &groupLimit,
		},
	}

	got := UserSubscriptionFromService(sub)
	require.NotNil(t, got)
	require.NotNil(t, got.TotalLimitUSD)
	require.Equal(t, instanceLimit, *got.TotalLimitUSD)
	require.Equal(t, 260.0, got.TotalUsageUSD)
	require.NotNil(t, got.Group)
	require.NotNil(t, got.Group.SubscriptionTotalLimitUSD)
	require.Equal(t, groupLimit, *got.Group.SubscriptionTotalLimitUSD)

	body, err := json.Marshal(got)
	require.NoError(t, err)
	require.Contains(t, string(body), `"total_limit_usd":3120`)
	require.Contains(t, string(body), `"total_usage_usd":260`)
	require.Contains(t, string(body), `"subscription_total_limit_usd":1560`)
}
