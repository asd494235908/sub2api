package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingKeyAffiliateWithdrawEnabled           = "affiliate_withdraw_enabled"
	SettingKeyAffiliateWithdrawMinAmount         = "affiliate_withdraw_min_amount"
	SettingKeyAffiliateWithdrawMaxAmount         = "affiliate_withdraw_max_amount"
	SettingKeyAffiliateWithdrawDailyRequestLimit = "affiliate_withdraw_daily_request_limit"
	SettingKeyAffiliateWithdrawHelpText          = "affiliate_withdraw_help_text"

	AffiliateWithdrawalStatusPendingReview = "pending_review"
	AffiliateWithdrawalStatusApproved      = "approved"
	AffiliateWithdrawalStatusPaid          = "paid"
	AffiliateWithdrawalStatusRejected      = "rejected"
	AffiliateWithdrawalStatusFailed        = "failed"
	AffiliateWithdrawalStatusCancelled     = "cancelled"

	AffiliateWithdrawActionRequest = "withdraw_request"
	AffiliateWithdrawActionReject  = "withdraw_reject"
	AffiliateWithdrawActionPaid    = "withdraw_paid"
	AffiliateWithdrawActionFail    = "withdraw_fail"
)

var (
	ErrAffiliateWithdrawalUnavailable       = infraerrors.ServiceUnavailable("AFFILIATE_WITHDRAW_UNAVAILABLE", "affiliate withdrawal service unavailable")
	ErrAffiliateWithdrawalDisabled          = infraerrors.Forbidden("AFFILIATE_WITHDRAW_DISABLED", "affiliate withdrawal is disabled")
	ErrAffiliateWithdrawalInvalidAmount     = infraerrors.BadRequest("AFFILIATE_WITHDRAW_INVALID_AMOUNT", "invalid withdrawal amount")
	ErrAffiliateWithdrawalInvalidPayoutInfo = infraerrors.BadRequest("AFFILIATE_WITHDRAW_INVALID_PAYOUT_INFO", "invalid payout information")
	ErrAffiliateWithdrawalInsufficientQuota = infraerrors.BadRequest("AFFILIATE_WITHDRAW_INSUFFICIENT_QUOTA", "affiliate quota is not enough")
	ErrAffiliateWithdrawalDailyLimit        = infraerrors.BadRequest("AFFILIATE_WITHDRAW_DAILY_LIMIT", "daily withdrawal request limit exceeded")
	ErrAffiliateWithdrawalNotFound          = infraerrors.NotFound("AFFILIATE_WITHDRAW_NOT_FOUND", "affiliate withdrawal not found")
	ErrAffiliateWithdrawalInvalidStatus     = infraerrors.Conflict("AFFILIATE_WITHDRAW_INVALID_STATUS", "affiliate withdrawal status does not allow this operation")
)

type AffiliateWithdrawSettings struct {
	Enabled           bool    `json:"enabled"`
	MinAmount         float64 `json:"min_amount"`
	MaxAmount         float64 `json:"max_amount"`
	DailyRequestLimit int     `json:"daily_request_limit"`
	HelpText          string  `json:"help_text"`
}

type AffiliateWithdrawal struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	UserEmail         string     `json:"user_email,omitempty"`
	Username          string     `json:"username,omitempty"`
	Amount            float64    `json:"amount"`
	Status            string     `json:"status"`
	PayoutMethod      string     `json:"payout_method"`
	PayoutAccountNote string     `json:"payout_account_note"`
	AdminNote         string     `json:"admin_note,omitempty"`
	PayoutChannel     string     `json:"payout_channel,omitempty"`
	PayoutTradeNo     string     `json:"payout_trade_no,omitempty"`
	RejectReason      string     `json:"reject_reason,omitempty"`
	FailureReason     string     `json:"failure_reason,omitempty"`
	ReviewedBy        *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	PaidBy            *int64     `json:"paid_by,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type AffiliateWithdrawalCreateInput struct {
	UserID            int64
	Amount            float64 `json:"amount"`
	PayoutMethod      string  `json:"payout_method"`
	PayoutAccountNote string  `json:"payout_account_note"`
}

type AffiliateWithdrawalPaidInput struct {
	PayoutChannel string `json:"payout_channel"`
	PayoutTradeNo string `json:"payout_trade_no"`
	AdminNote     string `json:"admin_note"`
}

type AffiliateWithdrawalListFilter struct {
	Search   string
	Status   string
	Page     int
	PageSize int
}

type AffiliateWithdrawalRepository interface {
	CreateWithdrawal(ctx context.Context, input AffiliateWithdrawalCreateInput) (*AffiliateWithdrawal, error)
	ListUserWithdrawals(ctx context.Context, userID int64, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error)
	ListAdminWithdrawals(ctx context.Context, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error)
	GetWithdrawal(ctx context.Context, id int64) (*AffiliateWithdrawal, error)
	ApproveWithdrawal(ctx context.Context, id, adminID int64, note string) (*AffiliateWithdrawal, error)
	RejectWithdrawal(ctx context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error)
	MarkWithdrawalPaid(ctx context.Context, id, adminID int64, input AffiliateWithdrawalPaidInput) (*AffiliateWithdrawal, error)
	MarkWithdrawalFailed(ctx context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error)
}

type AffiliateWithdrawalService struct {
	repo           AffiliateWithdrawalRepository
	settingService *SettingService
}

func NewAffiliateWithdrawalService(repo AffiliateWithdrawalRepository, settingService *SettingService) *AffiliateWithdrawalService {
	return &AffiliateWithdrawalService{repo: repo, settingService: settingService}
}

func (s *AffiliateWithdrawalService) GetSettings(ctx context.Context) AffiliateWithdrawSettings {
	out := defaultAffiliateWithdrawSettings()
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return out
	}
	if v, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateWithdrawEnabled); err == nil {
		out.Enabled = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateWithdrawMinAmount); err == nil {
		out.MinAmount = parseNonNegativeFloat(v, out.MinAmount)
	}
	if v, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateWithdrawMaxAmount); err == nil {
		out.MaxAmount = parseNonNegativeFloat(v, out.MaxAmount)
	}
	if v, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateWithdrawDailyRequestLimit); err == nil {
		out.DailyRequestLimit = parseNonNegativeInt(v, out.DailyRequestLimit)
	}
	if v, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateWithdrawHelpText); err == nil {
		out.HelpText = strings.TrimSpace(v)
	}
	return normalizeAffiliateWithdrawSettings(out)
}

func (s *AffiliateWithdrawalService) UpdateSettings(ctx context.Context, settings AffiliateWithdrawSettings) (AffiliateWithdrawSettings, error) {
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return AffiliateWithdrawSettings{}, ErrAffiliateWithdrawalUnavailable
	}
	normalized := normalizeAffiliateWithdrawSettings(settings)
	err := s.settingService.settingRepo.SetMultiple(ctx, map[string]string{
		SettingKeyAffiliateWithdrawEnabled:           strconv.FormatBool(normalized.Enabled),
		SettingKeyAffiliateWithdrawMinAmount:         strconv.FormatFloat(normalized.MinAmount, 'f', 8, 64),
		SettingKeyAffiliateWithdrawMaxAmount:         strconv.FormatFloat(normalized.MaxAmount, 'f', 8, 64),
		SettingKeyAffiliateWithdrawDailyRequestLimit: strconv.Itoa(normalized.DailyRequestLimit),
		SettingKeyAffiliateWithdrawHelpText:          normalized.HelpText,
	})
	if err != nil {
		return AffiliateWithdrawSettings{}, err
	}
	return normalized, nil
}

func (s *AffiliateWithdrawalService) CreateWithdrawal(ctx context.Context, userID int64, input AffiliateWithdrawalCreateInput) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAffiliateWithdrawalUnavailable
	}
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	settings := s.GetSettings(ctx)
	if !settings.Enabled {
		return nil, ErrAffiliateWithdrawalDisabled
	}
	input.UserID = userID
	input.Amount = roundTo(input.Amount, 8)
	input.PayoutMethod = strings.TrimSpace(input.PayoutMethod)
	input.PayoutAccountNote = strings.TrimSpace(input.PayoutAccountNote)
	if err := validateAffiliateWithdrawalCreateInput(input, settings); err != nil {
		return nil, err
	}
	if settings.DailyRequestLimit > 0 {
		items, _, err := s.repo.ListUserWithdrawals(ctx, userID, AffiliateWithdrawalListFilter{Page: 1, PageSize: 200})
		if err != nil {
			return nil, err
		}
		if countAffiliateWithdrawalCreatedToday(items) >= settings.DailyRequestLimit {
			return nil, ErrAffiliateWithdrawalDailyLimit
		}
	}
	return s.repo.CreateWithdrawal(ctx, input)
}

func (s *AffiliateWithdrawalService) ListUserWithdrawals(ctx context.Context, userID int64, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrAffiliateWithdrawalUnavailable
	}
	if userID <= 0 {
		return nil, 0, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	return s.repo.ListUserWithdrawals(ctx, userID, normalizeAffiliateWithdrawalListFilter(filter))
}

func (s *AffiliateWithdrawalService) ListAdminWithdrawals(ctx context.Context, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrAffiliateWithdrawalUnavailable
	}
	return s.repo.ListAdminWithdrawals(ctx, normalizeAffiliateWithdrawalListFilter(filter))
}

func (s *AffiliateWithdrawalService) ApproveWithdrawal(ctx context.Context, id, adminID int64, note string) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAffiliateWithdrawalUnavailable
	}
	return s.repo.ApproveWithdrawal(ctx, id, adminID, strings.TrimSpace(note))
}

func (s *AffiliateWithdrawalService) RejectWithdrawal(ctx context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAffiliateWithdrawalUnavailable
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, infraerrors.BadRequest("AFFILIATE_WITHDRAW_REASON_REQUIRED", "reject reason is required")
	}
	return s.repo.RejectWithdrawal(ctx, id, adminID, reason)
}

func (s *AffiliateWithdrawalService) MarkWithdrawalPaid(ctx context.Context, id, adminID int64, input AffiliateWithdrawalPaidInput) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAffiliateWithdrawalUnavailable
	}
	input.PayoutChannel = strings.TrimSpace(input.PayoutChannel)
	input.PayoutTradeNo = strings.TrimSpace(input.PayoutTradeNo)
	input.AdminNote = strings.TrimSpace(input.AdminNote)
	if input.PayoutChannel == "" {
		return nil, infraerrors.BadRequest("AFFILIATE_WITHDRAW_PAYOUT_CHANNEL_REQUIRED", "payout channel is required")
	}
	return s.repo.MarkWithdrawalPaid(ctx, id, adminID, input)
}

func (s *AffiliateWithdrawalService) MarkWithdrawalFailed(ctx context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAffiliateWithdrawalUnavailable
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, infraerrors.BadRequest("AFFILIATE_WITHDRAW_REASON_REQUIRED", "failure reason is required")
	}
	return s.repo.MarkWithdrawalFailed(ctx, id, adminID, reason)
}

func defaultAffiliateWithdrawSettings() AffiliateWithdrawSettings {
	return AffiliateWithdrawSettings{
		Enabled:           false,
		MinAmount:         1,
		MaxAmount:         0,
		DailyRequestLimit: 3,
		HelpText:          "",
	}
}

func normalizeAffiliateWithdrawSettings(settings AffiliateWithdrawSettings) AffiliateWithdrawSettings {
	if settings.MinAmount < 0 || math.IsNaN(settings.MinAmount) || math.IsInf(settings.MinAmount, 0) {
		settings.MinAmount = 0
	}
	if settings.MaxAmount < 0 || math.IsNaN(settings.MaxAmount) || math.IsInf(settings.MaxAmount, 0) {
		settings.MaxAmount = 0
	}
	if settings.DailyRequestLimit < 0 {
		settings.DailyRequestLimit = 0
	}
	settings.HelpText = strings.TrimSpace(settings.HelpText)
	return settings
}

func validateAffiliateWithdrawalCreateInput(input AffiliateWithdrawalCreateInput, settings AffiliateWithdrawSettings) error {
	if input.Amount <= 0 || math.IsNaN(input.Amount) || math.IsInf(input.Amount, 0) {
		return ErrAffiliateWithdrawalInvalidAmount
	}
	if settings.MinAmount > 0 && input.Amount+1e-9 < settings.MinAmount {
		return ErrAffiliateWithdrawalInvalidAmount
	}
	if settings.MaxAmount > 0 && input.Amount-settings.MaxAmount > 1e-9 {
		return ErrAffiliateWithdrawalInvalidAmount
	}
	if input.PayoutMethod == "" || len(input.PayoutMethod) > 32 || input.PayoutAccountNote == "" || len(input.PayoutAccountNote) > 500 {
		return ErrAffiliateWithdrawalInvalidPayoutInfo
	}
	return nil
}

func normalizeAffiliateWithdrawalListFilter(filter AffiliateWithdrawalListFilter) AffiliateWithdrawalListFilter {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = strings.TrimSpace(filter.Status)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return filter
}

func countAffiliateWithdrawalCreatedToday(items []AffiliateWithdrawal) int {
	now := time.Now()
	count := 0
	for _, item := range items {
		if item.CreatedAt.IsZero() {
			continue
		}
		t := item.CreatedAt.Local()
		if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
			count++
		}
	}
	return count
}

func parseNonNegativeFloat(raw string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return fallback
	}
	return value
}

func parseNonNegativeInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func IsAffiliateWithdrawalNotFound(err error) bool {
	return errors.Is(err, ErrAffiliateWithdrawalNotFound)
}
