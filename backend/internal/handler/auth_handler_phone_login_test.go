//go:build unit

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authHandlerPhoneSMSCacheStub struct {
	verifyErr error
}

type authHandlerSettingRepoStub struct {
	values map[string]string
}

func (s *authHandlerSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *authHandlerSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", service.ErrSettingNotFound
}

func (s *authHandlerSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *authHandlerSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if v, ok := s.values[key]; ok {
			result[key] = v
		}
	}
	return result, nil
}

func (s *authHandlerSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *authHandlerSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *authHandlerSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func (s *authHandlerPhoneSMSCacheStub) GetPhoneVerificationCode(context.Context, string) (*service.VerificationCodeData, error) {
	return &service.VerificationCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func (s *authHandlerPhoneSMSCacheStub) SetPhoneVerificationCode(context.Context, string, *service.VerificationCodeData, time.Duration) error {
	return nil
}

func (s *authHandlerPhoneSMSCacheStub) DeletePhoneVerificationCode(context.Context, string) error {
	return nil
}

type authHandlerPhoneSMSProviderStub struct{}

func (s *authHandlerPhoneSMSProviderStub) SendVerificationCode(context.Context, string, string) error {
	return nil
}

func (s *authHandlerPhoneSMSProviderStub) SendTextMessage(context.Context, string, string) error {
	return nil
}

func (s *authHandlerPhoneSMSProviderStub) SendTemplateMessage(context.Context, string, string, string) error {
	return nil
}

func (s *authHandlerPhoneSMSProviderStub) SendActivityTemplateMessage(context.Context, string, string, []string) error {
	return nil
}

func newAuthHandlerForPhoneTests(t *testing.T, repo *userHandlerRepoStub, settings map[string]string) *AuthHandler {
	t.Helper()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:                 "test-secret",
			ExpireHour:             1,
			RefreshTokenExpireDays: 7,
		},
	}

	settingSvc := service.NewSettingService(&authHandlerSettingRepoStub{values: settings}, cfg)
	authSvc := service.NewAuthService(nil, repo, nil, &userHandlerRefreshTokenCacheStub{}, cfg, settingSvc, nil, nil, nil, nil, nil, nil)
	smsSvc := service.NewSMSService(&authHandlerPhoneSMSCacheStub{}, &authHandlerPhoneSMSProviderStub{})
	authSvc.SetSMSService(smsSvc)

	return &AuthHandler{
		cfg:         cfg,
		authService: authSvc,
		settingSvc:  settingSvc,
		smsService:  smsSvc,
	}
}

func TestAuthHandlerSendVerifyCodeRegisterRejectsExistingPhone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &userHandlerRepoStub{
		user: &service.User{
			ID:          41,
			Email:       "exists@test.com",
			PhoneNumber: "+8613800138000",
			Role:        service.RoleUser,
			Status:      service.StatusActive,
		},
	}
	handler := newAuthHandlerForPhoneTests(t, repo, map[string]string{
		service.SettingKeyPhoneVerifyEnabled: "true",
	})

	body := []byte(`{"phone_number":"13800138000","purpose":"register"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-verify-code", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SendVerifyCode(c)

	require.Equal(t, http.StatusConflict, recorder.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, http.StatusConflict, resp.Code)
	require.Equal(t, "PHONE_EXISTS", resp.Reason)
}

func TestAuthHandlerPhoneLoginAllowsSMSCodeWithoutPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &userHandlerRepoStub{
		user: &service.User{
			ID:           42,
			Email:        "phone-login" + service.LinuxDoConnectSyntheticEmailDomain,
			PhoneNumber:  "+8613800138000",
			Role:         service.RoleUser,
			Status:       service.StatusActive,
			TokenVersion: 1,
		},
	}
	handler := newAuthHandlerForPhoneTests(t, repo, map[string]string{
		service.SettingKeyPhoneVerifyEnabled: "true",
	})

	body := []byte(`{"identifier":"13800138000","phone_number":"13800138000","sms_code":"123456"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			User         struct {
				ID    int64  `json:"id"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.NotEmpty(t, resp.Data.AccessToken)
	require.NotEmpty(t, resp.Data.RefreshToken)
	require.Equal(t, int64(42), resp.Data.User.ID)
	require.Equal(t, "phone-login"+service.LinuxDoConnectSyntheticEmailDomain, resp.Data.User.Email)
}

func TestAuthHandlerPhoneLoginRejectsWhenPhoneVerifyDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &userHandlerRepoStub{
		user: &service.User{
			ID:           43,
			Email:        "phone-login-disabled" + service.LinuxDoConnectSyntheticEmailDomain,
			PhoneNumber:  "+8613800138000",
			Role:         service.RoleUser,
			Status:       service.StatusActive,
			TokenVersion: 1,
		},
	}
	handler := newAuthHandlerForPhoneTests(t, repo, map[string]string{
		service.SettingKeyPhoneVerifyEnabled: "false",
	})

	body := []byte(`{"identifier":"13800138000","phone_number":"13800138000","sms_code":"123456"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, "Phone verification is disabled", resp.Message)
}
