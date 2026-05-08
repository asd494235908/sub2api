//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSettingService_AffiliateSignupRewardSettings_DefaultsAndParsing(t *testing.T) {
	svc := NewSettingService(&settingRepoStub{values: map[string]string{}}, &config.Config{})

	require.False(t, svc.IsAffiliateSignupRewardEnabled(context.Background()))
	require.InDelta(t, 0.0, svc.GetAffiliateSignupRewardAmount(context.Background()), 1e-9)

	got := svc.parseSettings(map[string]string{
		SettingKeyAffiliateSignupRewardEnabled: "true",
		SettingKeyAffiliateSignupRewardAmount:  "15.5",
	})
	require.True(t, got.AffiliateSignupRewardEnabled)
	require.InDelta(t, 15.5, got.AffiliateSignupRewardAmount, 1e-9)
}

func TestSettingService_UpdateSettings_PersistsAffiliateSignupRewardSettings(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		AffiliateSignupRewardEnabled: true,
		AffiliateSignupRewardAmount:  12.34,
	})
	require.NoError(t, err)
	require.Equal(t, "true", repo.updates[SettingKeyAffiliateSignupRewardEnabled])
	require.Equal(t, "12.34000000", repo.updates[SettingKeyAffiliateSignupRewardAmount])
}
