//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type smsSettingRepoStub struct {
	values map[string]string
}

func (s *smsSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *smsSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *smsSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *smsSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *smsSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *smsSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *smsSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type smsCacheStub struct {
	stored map[string]*VerificationCodeData
}

func (s *smsCacheStub) GetPhoneVerificationCode(context.Context, string) (*VerificationCodeData, error) {
	return nil, ErrInvalidVerifyCode
}

func (s *smsCacheStub) SetPhoneVerificationCode(_ context.Context, phoneNumber string, data *VerificationCodeData, _ time.Duration) error {
	if s.stored == nil {
		s.stored = map[string]*VerificationCodeData{}
	}
	s.stored[phoneNumber] = data
	return nil
}

func (s *smsCacheStub) DeletePhoneVerificationCode(context.Context, string) error {
	return nil
}

type smsProviderFactoryRecorder struct {
	settings []SMSProviderSettings
}

func (r *smsProviderFactoryRecorder) Provider(settings SMSProviderSettings) SMSProvider {
	r.settings = append(r.settings, settings)
	return smsProviderFunc(func(context.Context, string, string) error {
		return nil
	})
}

type smsProviderFunc func(context.Context, string, string) error

func (f smsProviderFunc) SendVerificationCode(ctx context.Context, phoneNumber, code string) error {
	return f(ctx, phoneNumber, code)
}

func (f smsProviderFunc) SendTextMessage(ctx context.Context, phoneNumber, message string) error {
	return f(ctx, phoneNumber, message)
}

func (f smsProviderFunc) SendTemplateMessage(ctx context.Context, phoneNumber, templateID, message string) error {
	return f(ctx, phoneNumber, message)
}

func (f smsProviderFunc) SendActivityTemplateMessage(ctx context.Context, phoneNumber, templateID string, vars []string) error {
	return f(ctx, phoneNumber, strings.Join(vars, "|"))
}

func TestResolveIHuyiSMSProviderSettingsUsesDBSettings(t *testing.T) {
	t.Setenv(envSMSIHuyiEnabled, "false")
	t.Setenv(envSMSIHuyiAPIID, "env-id")
	t.Setenv(envSMSIHuyiAPIKey, "env-key")
	t.Setenv(envSMSIHuyiTemplateID, "env-template")

	settings := ResolveIHuyiSMSProviderSettings(&smsSettingRepoStub{values: map[string]string{
		SettingKeySMSIHuyiEnabled:    "true",
		SettingKeySMSIHuyiAPIID:      "db-id",
		SettingKeySMSIHuyiAPIKey:     "db-key",
		SettingKeySMSIHuyiTemplateID: "db-template",
	}})

	require.True(t, settings.Enabled)
	require.Equal(t, "db-id", settings.Account)
	require.Equal(t, "db-key", settings.Password)
	require.Equal(t, "db-template", settings.TemplateID)
}

func TestResolveIHuyiSMSProviderSettingsFallsBackToEnvironment(t *testing.T) {
	t.Setenv(envSMSIHuyiEnabled, "true")
	t.Setenv(envSMSIHuyiAPIID, "env-id")
	t.Setenv(envSMSIHuyiAPIKey, "env-key")
	t.Setenv(envSMSIHuyiTemplateID, "env-template")

	settings := ResolveIHuyiSMSProviderSettings(&smsSettingRepoStub{values: map[string]string{}})

	require.True(t, settings.Enabled)
	require.Equal(t, "env-id", settings.Account)
	require.Equal(t, "env-key", settings.Password)
	require.Equal(t, "env-template", settings.TemplateID)
}

func TestResolveIHuyiSMSProviderSettingsDBDisableOverridesEnvironment(t *testing.T) {
	t.Setenv(envSMSIHuyiEnabled, "true")
	t.Setenv(envSMSIHuyiAPIID, "env-id")
	t.Setenv(envSMSIHuyiAPIKey, "env-key")

	settings := ResolveIHuyiSMSProviderSettings(&smsSettingRepoStub{values: map[string]string{
		SettingKeySMSIHuyiEnabled: "false",
	}})

	require.False(t, settings.Enabled)
	require.Equal(t, "env-id", settings.Account)
	require.Equal(t, "env-key", settings.Password)
}

func TestResolveIHuyiSMSProviderSettingsDefaultsTemplateID(t *testing.T) {
	settings := ResolveIHuyiSMSProviderSettings(&smsSettingRepoStub{values: map[string]string{
		SettingKeySMSIHuyiEnabled:    "true",
		SettingKeySMSIHuyiAPIID:      "db-id",
		SettingKeySMSIHuyiAPIKey:     "db-key",
		SettingKeySMSIHuyiTemplateID: "",
	}})

	require.Equal(t, defaultIHuyiTemplateID, settings.TemplateID)
}

func TestSMSServiceSendVerifyCodeUsesLatestDBSettings(t *testing.T) {
	_ = os.Unsetenv(envSMSIHuyiEnabled)
	_ = os.Unsetenv(envSMSIHuyiAPIID)
	_ = os.Unsetenv(envSMSIHuyiAPIKey)
	_ = os.Unsetenv(envSMSIHuyiTemplateID)

	repo := &smsSettingRepoStub{values: map[string]string{
		SettingKeySMSIHuyiEnabled:    "true",
		SettingKeySMSIHuyiAPIID:      "db-id-1",
		SettingKeySMSIHuyiAPIKey:     "db-key-1",
		SettingKeySMSIHuyiTemplateID: "template-1",
	}}
	recorder := &smsProviderFactoryRecorder{}
	svc := NewSMSServiceWithProviderFactory(&smsCacheStub{}, repo, recorder.Provider)

	require.NoError(t, svc.SendVerifyCode(context.Background(), "13800138000"))

	repo.values[SettingKeySMSIHuyiAPIID] = "db-id-2"
	repo.values[SettingKeySMSIHuyiAPIKey] = "db-key-2"
	repo.values[SettingKeySMSIHuyiTemplateID] = "template-2"

	require.NoError(t, svc.SendVerifyCode(context.Background(), "13800138001"))
	require.Len(t, recorder.settings, 2)
	require.Equal(t, "db-id-1", recorder.settings[0].Account)
	require.Equal(t, "db-key-1", recorder.settings[0].Password)
	require.Equal(t, "template-1", recorder.settings[0].TemplateID)
	require.Equal(t, "db-id-2", recorder.settings[1].Account)
	require.Equal(t, "db-key-2", recorder.settings[1].Password)
	require.Equal(t, "template-2", recorder.settings[1].TemplateID)
}

func TestIHuyiSMSProviderSendTemplateMessageUsesExplicitTemplateID(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		require.NoError(t, json.NewEncoder(w).Encode(iHuyiSMSResponse{Code: 2}))
	}))
	t.Cleanup(server.Close)

	provider := NewIHuyiSMSProvider(SMSProviderSettings{
		Enabled:    true,
		Account:    "account",
		Password:   "password",
		TemplateID: "global-template",
	}, server.Client())
	provider.baseURL = server.URL

	err := provider.SendTemplateMessage(context.Background(), "+8613800138000", "broadcast-template", "Hello Alice")

	require.NoError(t, err)
	require.Equal(t, "account", form.Get("account"))
	require.Equal(t, "password", form.Get("password"))
	require.Equal(t, "13800138000", form.Get("mobile"))
	require.Equal(t, "broadcast-template", form.Get("templateid"))
	require.Equal(t, "Hello Alice", form.Get("content"))
}

func TestIHuyiSMSProviderSendVerificationCodeKeepsGlobalTemplateID(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		require.NoError(t, json.NewEncoder(w).Encode(iHuyiSMSResponse{Code: 2}))
	}))
	t.Cleanup(server.Close)

	provider := NewIHuyiSMSProvider(SMSProviderSettings{
		Enabled:    true,
		Account:    "account",
		Password:   "password",
		TemplateID: "global-template",
	}, server.Client())
	provider.baseURL = server.URL

	err := provider.SendVerificationCode(context.Background(), "+8613800138000", "123456")

	require.NoError(t, err)
	require.Equal(t, "global-template", form.Get("templateid"))
	require.Equal(t, "123456", form.Get("content"))
}

func TestIHuyiSMSProviderSendActivityTemplateMessageJoinsVarsInOrder(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		require.NoError(t, json.NewEncoder(w).Encode(iHuyiSMSResponse{Code: 2}))
	}))
	t.Cleanup(server.Close)

	provider := NewIHuyiSMSProvider(SMSProviderSettings{
		Enabled:    true,
		Account:    "account",
		Password:   "password",
		TemplateID: "global-template",
	}, server.Client())
	provider.baseURL = server.URL

	err := provider.SendActivityTemplateMessage(context.Background(), "+8613800138000", "broadcast-template", []string{"20180515006", "张三", "100元"})

	require.NoError(t, err)
	require.Equal(t, "broadcast-template", form.Get("templateid"))
	require.Equal(t, "20180515006|张三|100元", form.Get("content"))
}
