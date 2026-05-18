//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingPublicRepoStub struct {
	values map[string]string
}

func (s *settingPublicRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingPublicRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *settingPublicRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingPublicRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingPublicRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range settings {
		s.values[key] = value
	}
	return nil
}

func (s *settingPublicRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *settingPublicRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_InitializeDefaultSettings_BackfillsMissingPhoneAndSMSDefaults(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled: "false",
			SettingKeyEmailVerifyEnabled:  "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{
		Default: config.DefaultConfig{
			UserConcurrency: 5,
			UserBalance:     0,
		},
	})

	err := svc.InitializeDefaultSettings(context.Background())
	require.NoError(t, err)

	require.Equal(t, "false", repo.values[SettingKeyPhoneVerifyEnabled])
	require.Equal(t, "false", repo.values[SettingKeySMSIHuyiEnabled])
	require.Equal(t, "", repo.values[SettingKeySMSIHuyiAPIID])
	require.Equal(t, "", repo.values[SettingKeySMSIHuyiAPIKey])
	require.Equal(t, "309190", repo.values[SettingKeySMSIHuyiTemplateID])
	require.Equal(t, "false", repo.values[SettingKeyWeeklyQuotaEnabled])
	require.Equal(t, "0", repo.values[SettingKeyWeeklyQuotaAmount])

	require.Equal(t, "false", repo.values[SettingKeyRegistrationEnabled])
	require.Equal(t, "true", repo.values[SettingKeyEmailVerifyEnabled])
}

func TestSettingService_GetPublicSettings_ExposesRegistrationEmailSuffixWhitelist(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled:              "true",
			SettingKeyEmailVerifyEnabled:               "true",
			SettingKeyRegistrationEmailSuffixWhitelist: `["@EXAMPLE.com"," @foo.bar ","@invalid_domain",""]`,
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"@example.com", "@foo.bar"}, settings.RegistrationEmailSuffixWhitelist)
}

func TestSettingService_GetPublicSettings_ExposesTablePreferences(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyTableDefaultPageSize: "50",
			SettingKeyTablePageSizeOptions: "[20,50,100]",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 50, settings.TableDefaultPageSize)
	require.Equal(t, []int{20, 50, 100}, settings.TablePageSizeOptions)
}

func TestSettingService_GetPublicSettings_ExposesForceEmailOnThirdPartySignup(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyForceEmailOnThirdPartySignup: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.ForceEmailOnThirdPartySignup)
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

func TestSettingService_GetPublicSettings_ExposesWeChatOAuthModeCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectAppID:               "wx-mp-app",
			SettingKeyWeChatConnectAppSecret:           "wx-mp-secret",
			SettingKeyWeChatConnectMode:                "mp",
			SettingKeyWeChatConnectScopes:              "snsapi_base",
			SettingKeyWeChatConnectOpenEnabled:         "true",
			SettingKeyWeChatConnectMPEnabled:           "true",
			SettingKeyWeChatConnectRedirectURL:         "https://api.example.com/api/v1/auth/oauth/wechat/callback",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.True(t, settings.WeChatOAuthMPEnabled)
}

func TestSettingService_GetPublicSettings_DoesNotExposeMobileOnlyWeChatAsWebOAuthAvailable(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectMobileEnabled:       "true",
			SettingKeyWeChatConnectMode:                "mobile",
			SettingKeyWeChatConnectMobileAppID:         "wx-mobile-app",
			SettingKeyWeChatConnectMobileAppSecret:     "wx-mobile-secret",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.False(t, settings.WeChatOAuthEnabled)
	require.False(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.True(t, settings.WeChatOAuthMobileEnabled)
}

func TestSettingService_GetPublicSettings_FallsBackToConfigForWeChatOAuthCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{
		WeChat: config.WeChatConnectConfig{
			Enabled:             true,
			OpenEnabled:         true,
			OpenAppID:           "wx-open-config",
			OpenAppSecret:       "wx-open-secret",
			FrontendRedirectURL: "/auth/wechat/config-callback",
		},
	})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.False(t, settings.WeChatOAuthMobileEnabled)
}
