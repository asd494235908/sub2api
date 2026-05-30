package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionTotalLimitChecksAndUsage(t *testing.T) {
	totalLimit := 1560.0
	sub := &UserSubscription{
		ID:            1,
		UserID:        100,
		GroupID:       10,
		StartsAt:      time.Now().Add(-time.Hour),
		ExpiresAt:     time.Now().Add(365 * 24 * time.Hour),
		Status:        SubscriptionStatusActive,
		TotalLimitUSD: &totalLimit,
		TotalUsageUSD: 1559.5,
	}
	group := &Group{ID: 10, SubscriptionType: SubscriptionTypeSubscription}

	require.True(t, sub.CheckTotalLimit(group, 0.5))
	require.False(t, sub.CheckTotalLimit(group, 0.6))

	svc := &SubscriptionService{}
	require.NoError(t, svc.CheckUsageLimits(context.Background(), sub, group, 0.5))
	require.ErrorIs(t, svc.CheckUsageLimits(context.Background(), sub, group, 0.6), ErrSubscriptionTotalLimitExceeded)
}

func TestLegacySubscriptionWithoutTotalLimitIsUnlimited(t *testing.T) {
	sub := &UserSubscription{
		ID:            1,
		UserID:        100,
		GroupID:       10,
		StartsAt:      time.Now().Add(-time.Hour),
		ExpiresAt:     time.Now().Add(365 * 24 * time.Hour),
		Status:        SubscriptionStatusActive,
		TotalLimitUSD: nil,
		TotalUsageUSD: 999999,
	}
	group := &Group{ID: 10, SubscriptionType: SubscriptionTypeSubscription}

	require.True(t, sub.CheckTotalLimit(group, 1))
	require.NoError(t, (&SubscriptionService{}).CheckUsageLimits(context.Background(), sub, group, 1))
}

func TestBillingCacheSubscriptionTotalLimitExhausted(t *testing.T) {
	totalLimit := 1560.0
	cache := &billingCacheWorkerStub{
		subscriptionData: &SubscriptionCacheData{
			Status:        SubscriptionStatusActive,
			ExpiresAt:     time.Now().Add(time.Hour),
			TotalLimitUSD: &totalLimit,
			TotalUsage:    totalLimit,
		},
	}
	svc := &BillingCacheService{cache: cache}
	group := &Group{ID: 10, SubscriptionType: SubscriptionTypeSubscription}

	err := svc.checkSubscriptionEligibility(context.Background(), 100, group, nil)
	require.True(t, errors.Is(err, ErrSubscriptionTotalLimitExceeded), "got %v", err)
}
