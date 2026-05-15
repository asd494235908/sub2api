package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrSMSNotConfigured     = infraerrors.ServiceUnavailable("SMS_NOT_CONFIGURED", "sms service not configured")
	ErrSMSCredentialInvalid = infraerrors.ServiceUnavailable("SMS_CREDENTIAL_INVALID", "sms credential is invalid")
	ErrSMSPhoneInvalid      = infraerrors.BadRequest("SMS_PHONE_INVALID", "phone number is invalid")
	ErrSMSQuotaExceeded     = infraerrors.TooManyRequests("SMS_QUOTA_EXCEEDED", "sms quota exceeded")
	ErrSMSSendFailed        = infraerrors.ServiceUnavailable("SMS_SEND_FAILED", "failed to send sms verification code")
	ErrSMSProviderDisabled  = infraerrors.BadRequest("SMS_PROVIDER_DISABLED", "sms provider is disabled")
)

const (
	defaultIHuyiSMSBaseURL = "https://api.ihuyi.com/sms/Submit.json"
	defaultIHuyiTemplateID = "309190"

	envSMSIHuyiEnabled    = "SMS_IHUYI_ENABLED"
	envSMSIHuyiAPIID      = "SMS_IHUYI_API_ID"
	envSMSIHuyiAPIKey     = "SMS_IHUYI_API_KEY"
	envSMSIHuyiTemplateID = "SMS_IHUYI_TEMPLATE_ID"
)

type SMSCache interface {
	GetPhoneVerificationCode(ctx context.Context, phoneNumber string) (*VerificationCodeData, error)
	SetPhoneVerificationCode(ctx context.Context, phoneNumber string, data *VerificationCodeData, ttl time.Duration) error
	DeletePhoneVerificationCode(ctx context.Context, phoneNumber string) error
}

type SMSProvider interface {
	SendVerificationCode(ctx context.Context, phoneNumber, code string) error
}

type SMSProviderFactory func(settings SMSProviderSettings) SMSProvider

type SMSService struct {
	cache           SMSCache
	provider        SMSProvider
	settingRepo     SettingRepository
	providerFactory SMSProviderFactory
}

func NewSMSService(cache SMSCache, provider SMSProvider) *SMSService {
	return &SMSService{cache: cache, provider: provider}
}

func NewSMSServiceWithProviderFactory(cache SMSCache, settingRepo SettingRepository, providerFactory SMSProviderFactory) *SMSService {
	return &SMSService{
		cache:           cache,
		settingRepo:     settingRepo,
		providerFactory: providerFactory,
	}
}

func (s *SMSService) GenerateVerifyCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

func (s *SMSService) SendVerifyCode(ctx context.Context, phoneNumber string) error {
	if s == nil || s.cache == nil {
		return ErrSMSNotConfigured
	}
	provider := s.resolveProvider()
	if provider == nil {
		return ErrSMSNotConfigured
	}
	phoneNumber = NormalizePhoneNumber(phoneNumber, "86")
	if phoneNumber == "" {
		return ErrSMSPhoneInvalid
	}

	existing, err := s.cache.GetPhoneVerificationCode(ctx, phoneNumber)
	if err == nil && existing != nil {
		remaining := verifyCodeCooldownRemaining(existing, time.Now())
		if remaining > 0 {
			return verifyCodeTooFrequentError(remaining)
		}
	}

	code, err := s.GenerateVerifyCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	data := &VerificationCodeData{
		Code:      code,
		Attempts:  0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(verifyCodeTTL),
	}
	if err := s.cache.SetPhoneVerificationCode(ctx, phoneNumber, data, verifyCodeTTL); err != nil {
		return fmt.Errorf("save phone verify code: %w", err)
	}
	if err := provider.SendVerificationCode(ctx, phoneNumber, code); err != nil {
		return err
	}
	return nil
}

func (s *SMSService) resolveProvider() SMSProvider {
	if s == nil {
		return nil
	}
	if s.providerFactory == nil {
		return s.provider
	}
	return s.providerFactory(ResolveIHuyiSMSProviderSettings(s.settingRepo))
}

func (s *SMSService) VerifyCode(ctx context.Context, phoneNumber, code string) error {
	if s == nil || s.cache == nil {
		return ErrSMSNotConfigured
	}
	phoneNumber = NormalizePhoneNumber(phoneNumber, "86")
	if phoneNumber == "" {
		return ErrSMSPhoneInvalid
	}

	data, err := s.cache.GetPhoneVerificationCode(ctx, phoneNumber)
	if err != nil || data == nil {
		return ErrInvalidVerifyCode
	}
	if data.Attempts >= maxVerifyCodeAttempts {
		return ErrVerifyCodeMaxAttempts
	}
	if subtle.ConstantTimeCompare([]byte(data.Code), []byte(code)) != 1 {
		data.Attempts++
		remaining := time.Until(data.ExpiresAt)
		if remaining <= 0 {
			return ErrInvalidVerifyCode
		}
		if err := s.cache.SetPhoneVerificationCode(ctx, phoneNumber, data, remaining); err != nil {
			slog.Error("failed to update phone verification attempt count", "phone_number", phoneNumber, "error", err)
		}
		if data.Attempts >= maxVerifyCodeAttempts {
			return ErrVerifyCodeMaxAttempts
		}
		return ErrInvalidVerifyCode
	}
	if err := s.cache.DeletePhoneVerificationCode(ctx, phoneNumber); err != nil {
		slog.Error("failed to delete phone verification code after success", "phone_number", phoneNumber, "error", err)
	}
	return nil
}

type SMSProviderSettings struct {
	Enabled    bool
	Account    string
	Password   string
	TemplateID string
}

type IHuyiSMSProvider struct {
	enabled    bool
	baseURL    string
	account    string
	password   string
	templateID string
	client     *http.Client
}

type iHuyiSMSResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	SMSID string `json:"smsid"`
}

func NewIHuyiSMSProvider(settings SMSProviderSettings, client *http.Client) *IHuyiSMSProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	templateID := strings.TrimSpace(settings.TemplateID)
	if templateID == "" {
		templateID = defaultIHuyiTemplateID
	}
	return &IHuyiSMSProvider{
		enabled:    settings.Enabled,
		baseURL:    defaultIHuyiSMSBaseURL,
		account:    strings.TrimSpace(settings.Account),
		password:   strings.TrimSpace(settings.Password),
		templateID: templateID,
		client:     client,
	}
}

func (p *IHuyiSMSProvider) SendVerificationCode(ctx context.Context, phoneNumber, code string) error {
	if p == nil {
		return ErrSMSNotConfigured
	}
	if !p.enabled {
		return ErrSMSProviderDisabled
	}
	if strings.TrimSpace(p.account) == "" || strings.TrimSpace(p.password) == "" {
		return ErrSMSNotConfigured
	}

	form := url.Values{}
	form.Set("account", p.account)
	form.Set("password", p.password)
	form.Set("mobile", normalizeIHuyiMobile(phoneNumber))
	form.Set("templateid", p.templateID)
	form.Set("content", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ErrSMSSendFailed.WithCause(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return ErrSMSSendFailed.WithCause(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrSMSSendFailed.WithCause(err)
	}

	var payload iHuyiSMSResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return ErrSMSSendFailed.WithCause(err)
	}
	if payload.Code == 2 {
		return nil
	}
	return mapIHuyiSMSResponseError(payload)
}

func mapIHuyiSMSResponseError(payload iHuyiSMSResponse) error {
	md := map[string]string{
		"provider": "ihuyi",
		"code":     strconv.Itoa(payload.Code),
	}
	switch payload.Code {
	case 4010, 4011, 4012, 4013:
		return ErrSMSCredentialInvalid.WithMetadata(md)
	case 4085, 4086:
		return ErrSMSQuotaExceeded.WithMetadata(md)
	case 4050:
		return ErrSMSPhoneInvalid.WithMetadata(md)
	default:
		return ErrSMSSendFailed.WithMetadata(md)
	}
}

func normalizeIHuyiMobile(phoneNumber string) string {
	normalized := NormalizePhoneNumber(phoneNumber, "86")
	normalized = strings.TrimPrefix(normalized, "+")
	if strings.HasPrefix(normalized, "86") && len(normalized) > 11 {
		return normalized[2:]
	}
	return normalized
}

type PlaceholderSMSProvider struct{}

func (p *PlaceholderSMSProvider) SendVerificationCode(ctx context.Context, phoneNumber, code string) error {
	slog.Warn("placeholder sms provider used", "phone_number", phoneNumber, "code", code)
	return nil
}

func LoadIHuyiSMSSettingsFromEnv() SMSProviderSettings {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv(envSMSIHuyiEnabled)), "true")
	return SMSProviderSettings{
		Enabled:    enabled,
		Account:    strings.TrimSpace(os.Getenv(envSMSIHuyiAPIID)),
		Password:   strings.TrimSpace(os.Getenv(envSMSIHuyiAPIKey)),
		TemplateID: strings.TrimSpace(os.Getenv(envSMSIHuyiTemplateID)),
	}
}

func ResolveIHuyiSMSProviderSettings(settingRepo SettingRepository) SMSProviderSettings {
	settings := LoadIHuyiSMSSettingsFromEnv()
	if settingRepo == nil {
		return settings
	}
	stored, err := settingRepo.GetMultiple(context.Background(), []string{
		SettingKeySMSIHuyiEnabled,
		SettingKeySMSIHuyiAPIID,
		SettingKeySMSIHuyiAPIKey,
		SettingKeySMSIHuyiTemplateID,
	})
	if err != nil {
		return settings
	}
	if rawEnabled, ok := stored[SettingKeySMSIHuyiEnabled]; ok {
		settings.Enabled = strings.TrimSpace(rawEnabled) == "true"
	}
	if rawAPIID, ok := stored[SettingKeySMSIHuyiAPIID]; ok {
		if trimmed := strings.TrimSpace(rawAPIID); trimmed != "" {
			settings.Account = trimmed
		}
	}
	if rawAPIKey, ok := stored[SettingKeySMSIHuyiAPIKey]; ok {
		if trimmed := strings.TrimSpace(rawAPIKey); trimmed != "" {
			settings.Password = trimmed
		}
	}
	if rawTemplateID, ok := stored[SettingKeySMSIHuyiTemplateID]; ok {
		settings.TemplateID = firstNonEmpty(rawTemplateID, defaultIHuyiTemplateID)
	} else {
		settings.TemplateID = firstNonEmpty(settings.TemplateID, defaultIHuyiTemplateID)
	}
	return settings
}
