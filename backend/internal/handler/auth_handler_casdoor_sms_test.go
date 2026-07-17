//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authHandlerCasdoorSMSProviderSpy struct {
	err       error
	phone     string
	template  string
	message   string
	callCount int
}

func (s *authHandlerCasdoorSMSProviderSpy) SendVerificationCode(context.Context, string, string) error {
	panic("unexpected SendVerificationCode call")
}

func (s *authHandlerCasdoorSMSProviderSpy) SendTextMessage(context.Context, string, string) error {
	panic("unexpected SendTextMessage call")
}

func (s *authHandlerCasdoorSMSProviderSpy) SendTemplateMessage(_ context.Context, phoneNumber, templateID, message string) error {
	s.callCount++
	s.phone = phoneNumber
	s.template = templateID
	s.message = message
	return s.err
}

func (s *authHandlerCasdoorSMSProviderSpy) SendActivityTemplateMessage(context.Context, string, string, []string) error {
	panic("unexpected SendActivityTemplateMessage call")
}

func newAuthHandlerForCasdoorSMSTest(provider service.SMSProvider) *AuthHandler {
	return &AuthHandler{
		smsService: service.NewSMSService(&authHandlerPhoneSMSCacheStub{}, provider),
	}
}

func performCasdoorSMSRequest(handler *AuthHandler, secret string, form url.Values) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/internal/casdoor/ihuyi-sms", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if secret != "" {
		req.Header.Set("X-Casdoor-SMS-Secret", secret)
	}
	c.Request = req

	handler.SendCasdoorIHuyiSMS(c)
	return recorder
}

func TestAuthHandlerSendCasdoorIHuyiSMSRequiresSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CASDOOR_IHUYI_SMS_SECRET", "test-secret")

	provider := &authHandlerCasdoorSMSProviderSpy{}
	handler := newAuthHandlerForCasdoorSMSTest(provider)

	form := url.Values{"mobile": {"+8613800138000"}, "content": {"123456"}}
	recorder := performCasdoorSMSRequest(handler, "wrong-secret", form)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Zero(t, provider.callCount)
}

func TestAuthHandlerSendCasdoorIHuyiSMSRequiresConfiguredSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CASDOOR_IHUYI_SMS_SECRET", "")

	provider := &authHandlerCasdoorSMSProviderSpy{}
	handler := newAuthHandlerForCasdoorSMSTest(provider)

	form := url.Values{"mobile": {"+8613800138000"}, "content": {"123456"}}
	recorder := performCasdoorSMSRequest(handler, "test-secret", form)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Zero(t, provider.callCount)
}

func TestAuthHandlerSendCasdoorIHuyiSMSSendsTemplateMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CASDOOR_IHUYI_SMS_SECRET", "test-secret")

	provider := &authHandlerCasdoorSMSProviderSpy{}
	handler := newAuthHandlerForCasdoorSMSTest(provider)

	form := url.Values{"mobile": {"+8613800138000"}, "content": {"123456"}}
	recorder := performCasdoorSMSRequest(handler, "test-secret", form)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, provider.callCount)
	require.Equal(t, "+8613800138000", provider.phone)
	require.Empty(t, provider.template)
	require.Equal(t, "123456", provider.message)
}
