//go:build unit

package service

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSettingService_GetPublicSettings_ExposesDefaultHomeLinks(t *testing.T) {
	repo := &settingPublicRepoStub{values: map[string]string{}}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"id":"gpshop","label":"格品购物","label_zh":"格品购物","label_en":"Gepin Shop","url":"https://card.gepinkeji.com","enabled":true,"sort_order":0},
		{"id":"gpci","label":"格品生图","label_zh":"格品生图","label_en":"Gepin Image","url":"https://chat.gepinkeji.com/","enabled":true,"sort_order":1}
	]`, settings.HomeLinks)
}

func TestSettingService_GetPublicSettings_ExposesPhoneVerifyEnabled(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyPhoneVerifyEnabled: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.PhoneVerifyEnabled)
}

func TestSettingService_GetPublicSettings_ExposesHomepageContactFields(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyQQGroup:       "123456789",
			SettingKeyWeChatContact: "sub2api_support",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "123456789", settings.QQGroup)
	require.Equal(t, "sub2api_support", settings.WeChatContact)
}

func TestSettingService_GetPublicSettings_ExposesConfiguredHomeLinks(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyHomeLinks: `[
				{"id":"custom-b","label":"B","label_zh":"乙","label_en":"B","url":"https://b.example.com","enabled":false,"sort_order":9},
				{"id":"custom-a","label":"A","label_zh":"甲","label_en":"A","url":"https://a.example.com","enabled":true,"sort_order":2}
			]`,
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"id":"custom-a","label":"A","label_zh":"甲","label_en":"A","url":"https://a.example.com","enabled":true,"sort_order":0},
		{"id":"custom-b","label":"B","label_zh":"乙","label_en":"B","url":"https://b.example.com","enabled":false,"sort_order":1}
	]`, settings.HomeLinks)
}

func TestSettingService_GetPublicSettings_ExposesWeeklyQuotaEnabled(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeeklyQuotaEnabled: "true",
			SettingKeyWeeklyQuotaAmount:  "18.88",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeeklyQuotaEnabled)

	all, err := svc.GetAllSettings(context.Background())
	require.NoError(t, err)
	require.True(t, all.WeeklyQuotaEnabled)
	require.Equal(t, 18.88, all.WeeklyQuotaAmount)
}

func TestSettingService_GetPublicSettings_ExposesLuckyWheelEnabled(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyLuckyWheelEnabled: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.LuckyWheelEnabled)
}

func TestSettingService_GetPublicSettings_ExposesLocalPaymentAndActivityFlags(t *testing.T) {
	repo := &settingPublicRepoStub{values: map[string]string{
		SettingPaymentEnabled:             "true",
		SettingKeyAffiliateEnabled:        "true",
		SettingKeyWeeklyQuotaEnabled:      "true",
		SettingKeyLuckyWheelEnabled:       "true",
		SettingKeyRechargeActivityEnabled: "true",
	}}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.PaymentEnabled)
	require.True(t, settings.AffiliateEnabled)
	require.True(t, settings.WeeklyQuotaEnabled)
	require.True(t, settings.LuckyWheelEnabled)
	require.True(t, settings.RechargeActivityEnabled)
}

func TestSettingService_PromptArchiveSettings_RoundTrip(t *testing.T) {
	repo := &settingPublicRepoStub{values: map[string]string{}, allowSet: true}
	svc := NewSettingService(repo, &config.Config{})
	want := &PromptArchiveSettingsView{
		Enabled:         true,
		AllGroups:       false,
		GroupIDs:        []int64{7, 11},
		Bucket:          "archive-test",
		UpdatedByUserID: 42,
	}

	require.NoError(t, svc.SetPromptArchiveSettings(context.Background(), want))
	got, err := svc.GetPromptArchiveSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSettingService_GetPublicSettings_PreservesMultilineHomepageContactFields(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyQQGroup:       "123456789\n987654321",
			SettingKeyWeChatContact: "sub2api_support\nwechat_helper",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "123456789\n987654321", settings.QQGroup)
	require.Equal(t, "sub2api_support\nwechat_helper", settings.WeChatContact)
}
