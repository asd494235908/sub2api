package service

import (
	"bytes"
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	SMSBroadcastModeFreeform SMSBroadcastMode = "freeform"
	SMSBroadcastModeTemplate SMSBroadcastMode = "template"

	SMSBroadcastStatusDraft     SMSBroadcastStatus = "draft"
	SMSBroadcastStatusQueued    SMSBroadcastStatus = "queued"
	SMSBroadcastStatusRunning   SMSBroadcastStatus = "running"
	SMSBroadcastStatusSucceeded SMSBroadcastStatus = "succeeded"
	SMSBroadcastStatusFailed    SMSBroadcastStatus = "failed"
	SMSBroadcastStatusCanceled  SMSBroadcastStatus = "canceled"

	maxSMSBroadcastMessageRunes  = 500
	defaultSMSBroadcastPageSize  = 200
	defaultSMSBroadcastBatchSize = 50
)

var (
	ErrSMSBroadcastNilInput                  = infraerrors.BadRequest("SMS_BROADCAST_INPUT_REQUIRED", "sms broadcast input is required")
	ErrSMSBroadcastTitleRequired             = infraerrors.BadRequest("SMS_BROADCAST_TITLE_REQUIRED", "sms broadcast title is required")
	ErrSMSBroadcastTemplateIDRequired        = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_ID_REQUIRED", "sms broadcast template id is required")
	ErrSMSBroadcastMessageRequired           = infraerrors.BadRequest("SMS_BROADCAST_MESSAGE_REQUIRED", "sms broadcast message is required")
	ErrSMSBroadcastMessageTooLong            = infraerrors.BadRequest("SMS_BROADCAST_MESSAGE_TOO_LONG", "sms broadcast message is too long")
	ErrSMSBroadcastInvalidMode               = infraerrors.BadRequest("SMS_BROADCAST_MODE_INVALID", "sms broadcast mode is invalid")
	ErrSMSBroadcastTemplateVarKeyRequired    = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_VAR_KEY_REQUIRED", "sms broadcast template variable key is required")
	ErrSMSBroadcastTemplateVarValueRequired  = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_VAR_VALUE_REQUIRED", "sms broadcast template variable value is required")
	ErrSMSBroadcastTemplateVarKeyDuplicate   = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_VAR_KEY_DUPLICATE", "sms broadcast template variable key is duplicated")
	ErrSMSBroadcastTemplateVarSourceInvalid  = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_VAR_SOURCE_INVALID", "sms broadcast template variable source is invalid")
	ErrSMSBroadcastTemplateVarUserFieldEmpty = infraerrors.BadRequest("SMS_BROADCAST_TEMPLATE_VAR_USER_FIELD_EMPTY", "sms broadcast template variable user field is empty")
	ErrSMSBroadcastAudienceRequired          = infraerrors.BadRequest("SMS_BROADCAST_AUDIENCE_REQUIRED", "sms broadcast selected users are required")
	ErrSMSBroadcastAudienceUserInvalid       = infraerrors.BadRequest("SMS_BROADCAST_AUDIENCE_USER_INVALID", "sms broadcast selected user is invalid")
	ErrSMSBroadcastAudiencePhoneRequired     = infraerrors.BadRequest("SMS_BROADCAST_AUDIENCE_PHONE_REQUIRED", "sms broadcast selected user must have a valid phone number")
	ErrSMSBroadcastCampaignNotFound          = infraerrors.NotFound("SMS_BROADCAST_CAMPAIGN_NOT_FOUND", "sms broadcast campaign not found")
	ErrSMSBroadcastServiceUnavailable        = infraerrors.ServiceUnavailable("SMS_BROADCAST_SERVICE_UNAVAILABLE", "sms broadcast service unavailable")
)

type SMSBroadcastMode string
type SMSBroadcastStatus string
type SMSBroadcastTemplateVarSource string

const (
	SMSBroadcastTemplateVarSourcePhoneNumber SMSBroadcastTemplateVarSource = "phone_number"
	SMSBroadcastTemplateVarSourceEmail       SMSBroadcastTemplateVarSource = "email"
	SMSBroadcastTemplateVarSourceUsername    SMSBroadcastTemplateVarSource = "username"
)

type SMSBroadcastAudienceFilters struct {
	UserIDs              []int64          `json:"user_ids,omitempty"`
	Status               string           `json:"status,omitempty"`
	Role                 string           `json:"role,omitempty"`
	Search               string           `json:"search,omitempty"`
	GroupName            string           `json:"group_name,omitempty"`
	Attributes           map[int64]string `json:"attributes,omitempty"`
	IncludeSubscriptions *bool            `json:"include_subscriptions,omitempty"`
}

type SMSBroadcastRenderInput struct {
	Mode SMSBroadcastMode
	Body string
	User User
	Vars map[string]string
}

type SMSBroadcastTemplateVarRow struct {
	Key    string                        `json:"key"`
	Value  string                        `json:"value,omitempty"`
	Source SMSBroadcastTemplateVarSource `json:"source,omitempty"`
}

type SMSBroadcastRecipient struct {
	UserID       int64      `json:"user_id"`
	PhoneNumber  string     `json:"phone_number"`
	RawPhone     string     `json:"raw_phone"`
	User         User       `json:"-"`
	RenderedBody string     `json:"rendered_body"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SMSBroadcastCampaign struct {
	ID              int64                        `json:"id"`
	Title           string                       `json:"title"`
	Mode            SMSBroadcastMode             `json:"mode"`
	TemplateID      string                       `json:"template_id"`
	Body            string                       `json:"body"`
	RenderedBody    string                       `json:"rendered_body"`
	Status          SMSBroadcastStatus           `json:"status"`
	Audience        SMSBroadcastAudienceFilters  `json:"audience"`
	TemplateVars    map[string]string            `json:"template_vars,omitempty"`
	TemplateVarRows []SMSBroadcastTemplateVarRow `json:"template_var_rows,omitempty"`
	TotalRecipients int64                        `json:"total_recipients"`
	SentCount       int64                        `json:"sent_count"`
	FailedCount     int64                        `json:"failed_count"`
	SkippedCount    int64                        `json:"skipped_count"`
	CreatedBy       *int64                       `json:"created_by,omitempty"`
	UpdatedBy       *int64                       `json:"updated_by,omitempty"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	StartedAt       *time.Time                   `json:"started_at,omitempty"`
	FinishedAt      *time.Time                   `json:"finished_at,omitempty"`
	ErrorMessage    *string                      `json:"error_message,omitempty"`
}

type SMSBroadcastRecipientPage struct {
	Items []SMSBroadcastRecipient
	Page  *pagination.PaginationResult
}

type SMSBroadcastCampaignInput struct {
	Title      string
	Mode       SMSBroadcastMode
	TemplateID string
	Body       string
	Audience   SMSBroadcastAudienceFilters
	Vars       map[string]string
	VarRows    []SMSBroadcastTemplateVarRow
	ActorID    *int64
}

type SMSBroadcastRepository interface {
	CreateCampaign(ctx context.Context, campaign *SMSBroadcastCampaign) error
	UpdateCampaign(ctx context.Context, campaign *SMSBroadcastCampaign) error
	GetCampaignByID(ctx context.Context, id int64) (*SMSBroadcastCampaign, error)
	ListCampaigns(ctx context.Context, params pagination.PaginationParams) ([]SMSBroadcastCampaign, *pagination.PaginationResult, error)
	AppendRecipients(ctx context.Context, campaignID int64, recipients []SMSBroadcastRecipient) error
	ListRecipients(ctx context.Context, campaignID int64) ([]SMSBroadcastRecipient, error)
	ListRecipientsPaginated(ctx context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]SMSBroadcastRecipient, *pagination.PaginationResult, error)
	UpdateRecipient(ctx context.Context, campaignID int64, recipient *SMSBroadcastRecipient) error
}

type SMSBroadcastService struct {
	userRepo   UserRepository
	smsService *SMSService
	repo       SMSBroadcastRepository
	pageSize   int
	batchSize  int
	mu         sync.Mutex
	running    map[int64]struct{}
}

func NewSMSBroadcastService(userRepo UserRepository, cache SMSCache, provider SMSProvider, settingRepo SettingRepository, repo SMSBroadcastRepository) *SMSBroadcastService {
	var smsService *SMSService
	if cache != nil || provider != nil {
		smsService = NewSMSService(cache, provider)
	}
	return NewSMSBroadcastServiceWithSMSService(userRepo, smsService, repo)
}

func NewSMSBroadcastServiceWithSMSService(userRepo UserRepository, smsService *SMSService, repo SMSBroadcastRepository) *SMSBroadcastService {
	return &SMSBroadcastService{
		userRepo:   userRepo,
		smsService: smsService,
		repo:       repo,
		pageSize:   defaultSMSBroadcastPageSize,
		batchSize:  defaultSMSBroadcastBatchSize,
		running:    map[int64]struct{}{},
	}
}

func (s *SMSBroadcastService) ValidateMessageBody(body string) error {
	if len([]rune(strings.TrimSpace(body))) == 0 {
		return ErrSMSBroadcastMessageRequired
	}
	if len([]rune(body)) > maxSMSBroadcastMessageRunes {
		return ErrSMSBroadcastMessageTooLong
	}
	return nil
}

func (s *SMSBroadcastService) RenderMessage(input SMSBroadcastRenderInput) (string, error) {
	body := strings.TrimSpace(input.Body)
	if input.Mode == "" {
		input.Mode = SMSBroadcastModeFreeform
	}
	if input.Mode == SMSBroadcastModeFreeform {
		if err := s.ValidateMessageBody(body); err != nil {
			return "", err
		}
		return body, nil
	}
	if input.Mode != SMSBroadcastModeTemplate {
		return "", ErrSMSBroadcastInvalidMode
	}
	tpl, err := template.New("sms-broadcast").Option("missingkey=error").Parse(body)
	if err != nil {
		return "", err
	}
	data := map[string]any{
		"User": input.User,
		"Vars": input.Vars,
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	rendered := strings.TrimSpace(buf.String())
	if err := s.ValidateMessageBody(rendered); err != nil {
		return "", err
	}
	return rendered, nil
}

func NormalizeSMSBroadcastTemplateVarRows(rows []SMSBroadcastTemplateVarRow) ([]SMSBroadcastTemplateVarRow, map[string]string, error) {
	normalizedRows := make([]SMSBroadcastTemplateVarRow, 0, len(rows))
	vars := make(map[string]string, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.Key)
		value := strings.TrimSpace(row.Value)
		source := SMSBroadcastTemplateVarSource(strings.TrimSpace(string(row.Source)))
		if key == "" && value == "" && source == "" {
			continue
		}
		if key == "" {
			return nil, nil, ErrSMSBroadcastTemplateVarKeyRequired
		}
		if source != "" {
			if !isValidSMSBroadcastTemplateVarSource(source) {
				return nil, nil, ErrSMSBroadcastTemplateVarSourceInvalid
			}
		} else if value == "" {
			return nil, nil, ErrSMSBroadcastTemplateVarValueRequired
		}
		if _, exists := seen[key]; exists {
			return nil, nil, ErrSMSBroadcastTemplateVarKeyDuplicate
		}
		seen[key] = struct{}{}
		normalizedRows = append(normalizedRows, SMSBroadcastTemplateVarRow{Key: key, Value: value, Source: source})
		if source == "" {
			vars[key] = value
		}
	}
	if vars == nil {
		vars = map[string]string{}
	}
	return normalizedRows, vars, nil
}

func isValidSMSBroadcastTemplateVarSource(source SMSBroadcastTemplateVarSource) bool {
	switch source {
	case SMSBroadcastTemplateVarSourcePhoneNumber, SMSBroadcastTemplateVarSourceEmail, SMSBroadcastTemplateVarSourceUsername:
		return true
	default:
		return false
	}
}

func (s *SMSBroadcastService) ListRecipients(ctx context.Context, filters SMSBroadcastAudienceFilters) ([]SMSBroadcastRecipient, error) {
	if s == nil || s.userRepo == nil {
		return nil, ErrSMSBroadcastServiceUnavailable
	}
	pageSize := s.pageSize
	if pageSize <= 0 {
		pageSize = defaultSMSBroadcastPageSize
	}
	seen := map[string]struct{}{}
	recipients := make([]SMSBroadcastRecipient, 0)
	for page := 1; ; page++ {
		users, result, err := s.userRepo.ListWithFilters(ctx, pagination.PaginationParams{
			Page:      page,
			PageSize:  pageSize,
			SortBy:    "id",
			SortOrder: pagination.SortOrderAsc,
		}, convertSMSBroadcastFilters(filters))
		if err != nil {
			return nil, err
		}
		for i := range users {
			normalized := NormalizePhoneNumber(users[i].PhoneNumber, "86")
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			recipients = append(recipients, SMSBroadcastRecipient{
				UserID:      users[i].ID,
				PhoneNumber: normalized,
				RawPhone:    users[i].PhoneNumber,
				User:        users[i],
			})
		}
		if result == nil || int64(page*pageSize) >= result.Total || len(users) == 0 || len(users) < pageSize {
			break
		}
	}
	return recipients, nil
}

func (s *SMSBroadcastService) ListRecipientsByUserIDs(ctx context.Context, userIDs []int64) ([]SMSBroadcastRecipient, []int64, error) {
	if s == nil || s.userRepo == nil {
		return nil, nil, ErrSMSBroadcastServiceUnavailable
	}
	normalizedIDs := make([]int64, 0, len(userIDs))
	seenIDs := map[int64]struct{}{}
	seenPhones := map[string]struct{}{}
	recipients := make([]SMSBroadcastRecipient, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			return nil, nil, ErrSMSBroadcastAudienceUserInvalid
		}
		if _, exists := seenIDs[userID]; exists {
			continue
		}
		seenIDs[userID] = struct{}{}
		normalizedIDs = append(normalizedIDs, userID)
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return nil, nil, ErrSMSBroadcastAudienceUserInvalid
			}
			return nil, nil, err
		}
		if user == nil || user.ID <= 0 {
			return nil, nil, ErrSMSBroadcastAudienceUserInvalid
		}
		normalized := NormalizePhoneNumber(user.PhoneNumber, "86")
		if normalized == "" {
			return nil, nil, ErrSMSBroadcastAudiencePhoneRequired
		}
		if _, exists := seenPhones[normalized]; exists {
			continue
		}
		seenPhones[normalized] = struct{}{}
		recipients = append(recipients, SMSBroadcastRecipient{
			UserID:      user.ID,
			PhoneNumber: normalized,
			RawPhone:    user.PhoneNumber,
			User:        *user,
		})
	}
	if len(normalizedIDs) == 0 {
		return nil, nil, ErrSMSBroadcastAudienceRequired
	}
	if len(recipients) == 0 {
		return nil, nil, ErrSMSBroadcastAudiencePhoneRequired
	}
	return recipients, normalizedIDs, nil
}

func (s *SMSBroadcastService) CreateCampaign(ctx context.Context, input *SMSBroadcastCampaignInput) (*SMSBroadcastCampaign, error) {
	if input == nil {
		return nil, ErrSMSBroadcastNilInput
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrSMSBroadcastTitleRequired
	}
	templateID := strings.TrimSpace(input.TemplateID)
	if templateID == "" {
		return nil, ErrSMSBroadcastTemplateIDRequired
	}
	mode := input.Mode
	if mode == "" {
		mode = SMSBroadcastModeTemplate
	}
	varRows := input.VarRows
	if len(varRows) == 0 && len(input.Vars) > 0 {
		keys := make([]string, 0, len(input.Vars))
		for key := range input.Vars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			varRows = append(varRows, SMSBroadcastTemplateVarRow{Key: key, Value: input.Vars[key]})
		}
	}
	normalizedRows, vars, err := NormalizeSMSBroadcastTemplateVarRows(varRows)
	if err != nil {
		return nil, err
	}
	campaign := &SMSBroadcastCampaign{
		Title:           title,
		Mode:            mode,
		TemplateID:      templateID,
		Body:            "",
		RenderedBody:    "",
		Status:          SMSBroadcastStatusDraft,
		Audience:        input.Audience,
		TemplateVars:    vars,
		TemplateVarRows: normalizedRows,
		CreatedBy:       input.ActorID,
		UpdatedBy:       input.ActorID,
	}
	if s.repo != nil {
		if err := s.repo.CreateCampaign(ctx, campaign); err != nil {
			return nil, err
		}
	}
	return campaign, nil
}

func (s *SMSBroadcastService) CreateAndQueueCampaign(ctx context.Context, input *SMSBroadcastCampaignInput) (*SMSBroadcastCampaign, error) {
	if s == nil || s.repo == nil {
		return nil, ErrSMSBroadcastServiceUnavailable
	}
	if input == nil {
		return nil, ErrSMSBroadcastNilInput
	}
	if len(input.Audience.UserIDs) == 0 {
		return nil, ErrSMSBroadcastAudienceRequired
	}
	recipients, normalizedUserIDs, err := s.ListRecipientsByUserIDs(ctx, input.Audience.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := validateSMSBroadcastRecipientTemplateVarSources(input.VarRows, recipients); err != nil {
		return nil, err
	}
	input.Audience.UserIDs = normalizedUserIDs
	campaign, err := s.CreateCampaign(ctx, input)
	if err != nil {
		return nil, err
	}
	campaign.Audience.UserIDs = normalizedUserIDs
	campaign.TotalRecipients = int64(len(recipients))
	campaign.Status = SMSBroadcastStatusQueued
	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return nil, err
	}
	if err := s.repo.AppendRecipients(ctx, campaign.ID, recipients); err != nil {
		return nil, err
	}
	_ = s.StartCampaign(ctx, campaign.ID)
	return campaign, nil
}

func (s *SMSBroadcastService) StartCampaign(ctx context.Context, campaignID int64) error {
	if s == nil {
		return ErrSMSBroadcastServiceUnavailable
	}
	s.mu.Lock()
	if _, ok := s.running[campaignID]; ok {
		s.mu.Unlock()
		return nil
	}
	s.running[campaignID] = struct{}{}
	s.mu.Unlock()
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, campaignID)
			s.mu.Unlock()
		}()
		_ = s.executeCampaign(context.Background(), campaignID)
	}()
	return nil
}

func (s *SMSBroadcastService) CancelCampaign(ctx context.Context, campaignID int64) error {
	if s.repo == nil {
		return ErrSMSBroadcastServiceUnavailable
	}
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		return err
	}
	now := time.Now()
	campaign.Status = SMSBroadcastStatusCanceled
	campaign.FinishedAt = &now
	return s.repo.UpdateCampaign(ctx, campaign)
}

func (s *SMSBroadcastService) ListCampaigns(ctx context.Context, params pagination.PaginationParams) ([]SMSBroadcastCampaign, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrSMSBroadcastServiceUnavailable
	}
	return s.repo.ListCampaigns(ctx, params)
}

func (s *SMSBroadcastService) GetCampaignByID(ctx context.Context, id int64) (*SMSBroadcastCampaign, error) {
	if s == nil || s.repo == nil {
		return nil, ErrSMSBroadcastServiceUnavailable
	}
	return s.repo.GetCampaignByID(ctx, id)
}

func (s *SMSBroadcastService) ListRecipientsPaginated(ctx context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]SMSBroadcastRecipient, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrSMSBroadcastServiceUnavailable
	}
	return s.repo.ListRecipientsPaginated(ctx, campaignID, params, strings.TrimSpace(status))
}

func (s *SMSBroadcastService) PreviewAudience(ctx context.Context, filters SMSBroadcastAudienceFilters) (int64, []SMSBroadcastRecipient, error) {
	recipients, err := s.ListRecipients(ctx, filters)
	if err != nil {
		return 0, nil, err
	}
	limit := int(math.Min(5, float64(len(recipients))))
	sample := make([]SMSBroadcastRecipient, 0, limit)
	for i := 0; i < limit; i++ {
		sample = append(sample, recipients[i])
	}
	return int64(len(recipients)), sample, nil
}

func (s *SMSBroadcastService) executeCampaign(ctx context.Context, campaignID int64) error {
	if s.repo == nil || s.smsService == nil {
		return ErrSMSBroadcastServiceUnavailable
	}
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		return err
	}
	if campaign == nil {
		return ErrSMSBroadcastCampaignNotFound
	}
	recipients, err := s.repo.ListRecipients(ctx, campaignID)
	if err != nil {
		return err
	}
	now := time.Now()
	campaign.Status = SMSBroadcastStatusRunning
	campaign.StartedAt = &now
	campaign.TotalRecipients = int64(len(recipients))
	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return err
	}

	var sent, failed, skipped int64
	for i := range recipients {
		if ctx.Err() != nil {
			break
		}
		if latest, err := s.repo.GetCampaignByID(ctx, campaignID); err == nil && latest != nil && latest.Status == SMSBroadcastStatusCanceled {
			return nil
		}
		if recipients[i].User.ID == 0 && s.userRepo != nil {
			if user, err := s.userRepo.GetByID(ctx, recipients[i].UserID); err == nil && user != nil {
				recipients[i].User = *user
			}
		}
		varValues, err := templateVarValues(campaign.TemplateVarRows, recipients[i].User)
		if err != nil {
			failed++
			msg := err.Error()
			recipients[i].Status = "failed"
			recipients[i].ErrorMessage = &msg
			_ = s.repo.UpdateRecipient(ctx, campaignID, &recipients[i])
			continue
		}
		if err := s.smsService.SendActivityTemplateMessage(ctx, recipients[i].PhoneNumber, campaign.TemplateID, varValues); err != nil {
			failed++
			msg := err.Error()
			recipients[i].Status = "failed"
			recipients[i].ErrorMessage = &msg
		} else {
			sent++
			recipients[i].Status = "succeeded"
			sentAt := time.Now()
			recipients[i].SentAt = &sentAt
		}
		_ = s.repo.UpdateRecipient(ctx, campaignID, &recipients[i])
	}
	finished := time.Now()
	campaign.SentCount = sent
	campaign.FailedCount = failed
	campaign.SkippedCount = skipped
	campaign.FinishedAt = &finished
	if failed > 0 && sent == 0 {
		campaign.Status = SMSBroadcastStatusFailed
	} else {
		campaign.Status = SMSBroadcastStatusSucceeded
	}
	return s.repo.UpdateCampaign(ctx, campaign)
}

func templateVarValues(rows []SMSBroadcastTemplateVarRow, user User) ([]string, error) {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		value, err := resolveSMSBroadcastTemplateVarValue(row, user)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func resolveSMSBroadcastTemplateVarValue(row SMSBroadcastTemplateVarRow, user User) (string, error) {
	source := SMSBroadcastTemplateVarSource(strings.TrimSpace(string(row.Source)))
	if source == "" {
		return row.Value, nil
	}
	var value string
	switch source {
	case SMSBroadcastTemplateVarSourcePhoneNumber:
		value = NormalizePhoneNumber(user.PhoneNumber, "86")
	case SMSBroadcastTemplateVarSourceEmail:
		value = strings.TrimSpace(user.Email)
	case SMSBroadcastTemplateVarSourceUsername:
		value = strings.TrimSpace(user.Username)
	default:
		return "", ErrSMSBroadcastTemplateVarSourceInvalid
	}
	if value == "" {
		return "", ErrSMSBroadcastTemplateVarUserFieldEmpty
	}
	return value, nil
}

func validateSMSBroadcastRecipientTemplateVarSources(rows []SMSBroadcastTemplateVarRow, recipients []SMSBroadcastRecipient) error {
	normalizedRows, _, err := NormalizeSMSBroadcastTemplateVarRows(rows)
	if err != nil {
		return err
	}
	for i := range recipients {
		for _, row := range normalizedRows {
			if row.Source == "" {
				continue
			}
			if _, err := resolveSMSBroadcastTemplateVarValue(row, recipients[i].User); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertSMSBroadcastFilters(filters SMSBroadcastAudienceFilters) UserListFilters {
	return UserListFilters{
		Status:               filters.Status,
		Role:                 filters.Role,
		Search:               filters.Search,
		GroupName:            filters.GroupName,
		Attributes:           filters.Attributes,
		IncludeSubscriptions: filters.IncludeSubscriptions,
	}
}

func sortRecipientsByUserID(recipients []SMSBroadcastRecipient) {
	sort.SliceStable(recipients, func(i, j int) bool {
		return recipients[i].UserID < recipients[j].UserID
	})
}
