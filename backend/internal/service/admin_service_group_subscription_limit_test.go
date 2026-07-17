//go:build unit

package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdminService_UpdateGroup_PreservesSubscriptionTotalLimitWhenOmitted(t *testing.T) {
	totalLimit := 1560.0
	existingGroup := &Group{
		ID:                        1,
		Name:                      "early-bird",
		Platform:                  PlatformOpenAI,
		Status:                    StatusActive,
		SubscriptionType:          SubscriptionTypeSubscription,
		SubscriptionTotalLimitUSD: &totalLimit,
	}
	repo := &groupRepoStubForAdmin{getByID: existingGroup}
	svc := &adminServiceImpl{groupRepo: repo}

	group, err := svc.UpdateGroup(context.Background(), 1, &UpdateGroupInput{Name: "early-bird-updated"})
	require.NoError(t, err)
	require.NotNil(t, group)

	require.NotNil(t, repo.updated)
	require.NotNil(t, repo.updated.SubscriptionTotalLimitUSD)
	require.InDelta(t, totalLimit, *repo.updated.SubscriptionTotalLimitUSD, 0.0001)
}
