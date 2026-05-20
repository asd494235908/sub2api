//go:build unit

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type smsBroadcastUserRepoStub struct {
	users      []User
	usersByID  map[int64]*User
	filters    UserListFilters
	getByIDSeq []int64
}

func (s *smsBroadcastUserRepoStub) Create(context.Context, *User) error {
	panic("unexpected Create call")
}
func (s *smsBroadcastUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	s.getByIDSeq = append(s.getByIDSeq, id)
	if s.usersByID != nil {
		user, ok := s.usersByID[id]
		if !ok {
			return nil, nil
		}
		copied := *user
		return &copied, nil
	}
	for i := range s.users {
		if s.users[i].ID == id {
			copied := s.users[i]
			return &copied, nil
		}
	}
	return nil, nil
}
func (s *smsBroadcastUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}
func (s *smsBroadcastUserRepoStub) GetByPhone(context.Context, string) (*User, error) {
	panic("unexpected GetByPhone call")
}
func (s *smsBroadcastUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}
func (s *smsBroadcastUserRepoStub) Update(context.Context, *User) error {
	panic("unexpected Update call")
}
func (s *smsBroadcastUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}
func (s *smsBroadcastUserRepoStub) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}
func (s *smsBroadcastUserRepoStub) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}
func (s *smsBroadcastUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}
func (s *smsBroadcastUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *smsBroadcastUserRepoStub) ListWithFilters(_ context.Context, params pagination.PaginationParams, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	s.filters = filters
	out := make([]User, 0, len(s.users))
	for _, user := range s.users {
		if filters.Status != "" && user.Status != filters.Status {
			continue
		}
		if filters.Role != "" && user.Role != filters.Role {
			continue
		}
		out = append(out, user)
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Page: params.Page, PageSize: params.PageSize}, nil
}
func (s *smsBroadcastUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}
func (s *smsBroadcastUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}
func (s *smsBroadcastUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}
func (s *smsBroadcastUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}
func (s *smsBroadcastUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}
func (s *smsBroadcastUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}
func (s *smsBroadcastUserRepoStub) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchSetConcurrency call")
}
func (s *smsBroadcastUserRepoStub) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchAddConcurrency call")
}
func (s *smsBroadcastUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}
func (s *smsBroadcastUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}
func (s *smsBroadcastUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}
func (s *smsBroadcastUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}
func (s *smsBroadcastUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}
func (s *smsBroadcastUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}
func (s *smsBroadcastUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}
func (s *smsBroadcastUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}
func (s *smsBroadcastUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

type smsBroadcastSMSCacheStub struct{}

func (smsBroadcastSMSCacheStub) GetPhoneVerificationCode(context.Context, string) (*VerificationCodeData, error) {
	panic("unexpected GetPhoneVerificationCode call")
}
func (smsBroadcastSMSCacheStub) SetPhoneVerificationCode(context.Context, string, *VerificationCodeData, time.Duration) error {
	panic("unexpected SetPhoneVerificationCode call")
}
func (smsBroadcastSMSCacheStub) DeletePhoneVerificationCode(context.Context, string) error {
	panic("unexpected DeletePhoneVerificationCode call")
}

type smsBroadcastProviderStub struct {
	sent []struct {
		phone      string
		templateID string
		body       string
	}
}

func (s *smsBroadcastProviderStub) SendVerificationCode(_ context.Context, phoneNumber, code string) error {
	s.sent = append(s.sent, struct {
		phone      string
		templateID string
		body       string
	}{phone: phoneNumber, body: code})
	return nil
}

func (s *smsBroadcastProviderStub) SendTextMessage(_ context.Context, phoneNumber, message string) error {
	s.sent = append(s.sent, struct {
		phone      string
		templateID string
		body       string
	}{phone: phoneNumber, body: message})
	return nil
}

func (s *smsBroadcastProviderStub) SendTemplateMessage(_ context.Context, phoneNumber, templateID, message string) error {
	s.sent = append(s.sent, struct {
		phone      string
		templateID string
		body       string
	}{phone: phoneNumber, templateID: templateID, body: message})
	return nil
}

func (s *smsBroadcastProviderStub) SendActivityTemplateMessage(_ context.Context, phoneNumber, templateID string, vars []string) error {
	s.sent = append(s.sent, struct {
		phone      string
		templateID string
		body       string
	}{phone: phoneNumber, templateID: templateID, body: strings.Join(vars, "|")})
	return nil
}

type smsBroadcastRepoStub struct {
	campaign         *SMSBroadcastCampaign
	recipients       []SMSBroadcastRecipient
	recipientsPageFn func(ctx context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]SMSBroadcastRecipient, *pagination.PaginationResult, error)
}

func (s *smsBroadcastRepoStub) CreateCampaign(_ context.Context, campaign *SMSBroadcastCampaign) error {
	copied := *campaign
	copied.ID = 1
	s.campaign = &copied
	campaign.ID = copied.ID
	return nil
}

func (s *smsBroadcastRepoStub) UpdateCampaign(_ context.Context, campaign *SMSBroadcastCampaign) error {
	copied := *campaign
	s.campaign = &copied
	return nil
}

func (s *smsBroadcastRepoStub) GetCampaignByID(context.Context, int64) (*SMSBroadcastCampaign, error) {
	copied := *s.campaign
	return &copied, nil
}

func (s *smsBroadcastRepoStub) ListCampaigns(context.Context, pagination.PaginationParams) ([]SMSBroadcastCampaign, *pagination.PaginationResult, error) {
	panic("unexpected ListCampaigns call")
}

func (s *smsBroadcastRepoStub) AppendRecipients(_ context.Context, _ int64, recipients []SMSBroadcastRecipient) error {
	s.recipients = append([]SMSBroadcastRecipient(nil), recipients...)
	return nil
}

func (s *smsBroadcastRepoStub) ListRecipients(context.Context, int64) ([]SMSBroadcastRecipient, error) {
	return append([]SMSBroadcastRecipient(nil), s.recipients...), nil
}

func (s *smsBroadcastRepoStub) ListRecipientsPaginated(ctx context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]SMSBroadcastRecipient, *pagination.PaginationResult, error) {
	if s.recipientsPageFn != nil {
		return s.recipientsPageFn(ctx, campaignID, params, status)
	}
	return append([]SMSBroadcastRecipient(nil), s.recipients...), &pagination.PaginationResult{Total: int64(len(s.recipients)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (s *smsBroadcastRepoStub) UpdateRecipient(_ context.Context, _ int64, recipient *SMSBroadcastRecipient) error {
	for i := range s.recipients {
		if s.recipients[i].PhoneNumber == recipient.PhoneNumber {
			s.recipients[i] = *recipient
			return nil
		}
	}
	return nil
}

func TestSMSBroadcastService_ListRecipientsDeduplicatesPhoneNumbers(t *testing.T) {
	repo := &smsBroadcastUserRepoStub{
		users: []User{
			{ID: 1, PhoneNumber: "13800138000", Status: StatusActive, Role: RoleUser},
			{ID: 2, PhoneNumber: " 13800138000 ", Status: StatusActive, Role: RoleUser},
			{ID: 3, PhoneNumber: "", Status: StatusActive, Role: RoleUser},
			{ID: 4, PhoneNumber: "13900139000", Status: StatusDisabled, Role: RoleUser},
		},
	}
	svc := NewSMSBroadcastService(repo, &smsBroadcastSMSCacheStub{}, &smsBroadcastProviderStub{}, nil, nil)

	recipients, err := svc.ListRecipients(context.Background(), SMSBroadcastAudienceFilters{Status: StatusActive})
	require.NoError(t, err)
	require.Len(t, recipients, 1)
	require.Equal(t, "+8613800138000", recipients[0].PhoneNumber)
	require.Equal(t, int64(1), recipients[0].UserID)
	require.Equal(t, StatusActive, repo.filters.Status)
}

func TestSMSBroadcastService_ListRecipientsPassesAudienceFilters(t *testing.T) {
	repo := &smsBroadcastUserRepoStub{
		users: []User{
			{ID: 1, PhoneNumber: "13800138000", Status: StatusActive, Role: RoleUser},
		},
	}
	svc := NewSMSBroadcastService(repo, &smsBroadcastSMSCacheStub{}, &smsBroadcastProviderStub{}, nil, nil)
	includeSubscriptions := false
	attrs := map[int64]string{7: "vip"}

	_, err := svc.ListRecipients(context.Background(), SMSBroadcastAudienceFilters{
		Status:               StatusActive,
		Role:                 RoleUser,
		Search:               "alice",
		GroupName:            "premium",
		Attributes:           attrs,
		IncludeSubscriptions: &includeSubscriptions,
	})

	require.NoError(t, err)
	require.Equal(t, UserListFilters{
		Status:               StatusActive,
		Role:                 RoleUser,
		Search:               "alice",
		GroupName:            "premium",
		Attributes:           attrs,
		IncludeSubscriptions: &includeSubscriptions,
	}, repo.filters)
}

func TestSMSBroadcastService_CreateAndQueueRequiresSelectedUserIDs(t *testing.T) {
	svc := NewSMSBroadcastService(&smsBroadcastUserRepoStub{}, nil, nil, nil, &smsBroadcastRepoStub{})

	_, err := svc.CreateAndQueueCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title:      "Maintenance",
		TemplateID: "broadcast-template",
	})

	require.ErrorIs(t, err, ErrSMSBroadcastAudienceRequired)
}

func TestSMSBroadcastService_CreateAndQueueRejectsInvalidSelectedUsers(t *testing.T) {
	tests := []struct {
		name     string
		userIDs  []int64
		users    []User
		want     error
		wantGets []int64
	}{
		{
			name:    "invalid id",
			userIDs: []int64{1, 0},
			users: []User{
				{ID: 1, PhoneNumber: "13800138000", Status: StatusActive, Role: RoleUser},
			},
			want:     ErrSMSBroadcastAudienceUserInvalid,
			wantGets: []int64{1},
		},
		{
			name:    "missing user",
			userIDs: []int64{99},
			users:   []User{},
			want:    ErrSMSBroadcastAudienceUserInvalid,
		},
		{
			name:    "missing phone",
			userIDs: []int64{1},
			users: []User{
				{ID: 1, PhoneNumber: "", Status: StatusActive, Role: RoleUser},
			},
			want:     ErrSMSBroadcastAudiencePhoneRequired,
			wantGets: []int64{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &smsBroadcastUserRepoStub{users: tt.users}
			svc := NewSMSBroadcastService(userRepo, nil, nil, nil, &smsBroadcastRepoStub{})

			_, err := svc.CreateAndQueueCampaign(context.Background(), &SMSBroadcastCampaignInput{
				Title:      "Maintenance",
				TemplateID: "broadcast-template",
				Audience:   SMSBroadcastAudienceFilters{UserIDs: tt.userIDs},
			})

			require.ErrorIs(t, err, tt.want)
			if tt.wantGets != nil {
				require.Equal(t, tt.wantGets, userRepo.getByIDSeq)
			}
		})
	}
}

func TestSMSBroadcastService_CreateAndQueueSnapshotsSelectedUsersInOrder(t *testing.T) {
	userRepo := &smsBroadcastUserRepoStub{
		users: []User{
			{ID: 2, PhoneNumber: "13900139000", Status: StatusActive, Role: RoleUser},
			{ID: 1, PhoneNumber: "13800138000", Status: StatusActive, Role: RoleUser},
			{ID: 3, PhoneNumber: " 13900139000 ", Status: StatusActive, Role: RoleUser},
		},
	}
	repo := &smsBroadcastRepoStub{}
	svc := NewSMSBroadcastService(userRepo, &smsBroadcastSMSCacheStub{}, &smsBroadcastProviderStub{}, nil, repo)

	campaign, err := svc.CreateAndQueueCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title:      "Maintenance",
		TemplateID: "broadcast-template",
		Audience:   SMSBroadcastAudienceFilters{UserIDs: []int64{2, 1, 2, 3}},
	})

	require.NoError(t, err)
	require.Equal(t, []int64{2, 1, 3}, userRepo.getByIDSeq)
	require.Equal(t, []int64{2, 1, 3}, campaign.Audience.UserIDs)
	require.Len(t, repo.recipients, 2)
	require.Equal(t, int64(2), repo.recipients[0].UserID)
	require.Equal(t, "+8613900139000", repo.recipients[0].PhoneNumber)
	require.Equal(t, int64(1), repo.recipients[1].UserID)
	require.Equal(t, "+8613800138000", repo.recipients[1].PhoneNumber)
	require.Equal(t, int64(2), campaign.TotalRecipients)
}

func TestSMSBroadcastService_RenderTemplateUsesUserFieldsAndVars(t *testing.T) {
	svc := NewSMSBroadcastService(nil, nil, nil, nil, nil)

	rendered, err := svc.RenderMessage(SMSBroadcastRenderInput{
		Mode: SMSBroadcastModeTemplate,
		Body: "Hello {{.User.Username}}, {{.Vars.greeting}} {{.User.PhoneNumber}}",
		User: User{Username: "alice", PhoneNumber: "+8613800138000"},
		Vars: map[string]string{"greeting": "welcome"},
	})

	require.NoError(t, err)
	require.Equal(t, "Hello alice, welcome +8613800138000", rendered)
}

func TestSMSBroadcastService_CreateCampaignRequiresTemplateID(t *testing.T) {
	svc := NewSMSBroadcastService(nil, nil, nil, nil, nil)

	_, err := svc.CreateCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title: "Maintenance",
	})

	require.ErrorIs(t, err, ErrSMSBroadcastTemplateIDRequired)
}

func TestSMSBroadcastService_CreateCampaignNormalizesTemplateVarRows(t *testing.T) {
	repo := &smsBroadcastRepoStub{}
	svc := NewSMSBroadcastService(nil, nil, nil, nil, repo)

	campaign, err := svc.CreateCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title:      "Maintenance",
		TemplateID: " broadcast-template ",
		VarRows: []SMSBroadcastTemplateVarRow{
			{Key: "", Value: ""},
			{Key: " name ", Value: " Alice "},
			{Key: "window", Value: "tonight"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "broadcast-template", campaign.TemplateID)
	require.Equal(t, SMSBroadcastModeTemplate, campaign.Mode)
	require.Empty(t, campaign.Body)
	require.Empty(t, campaign.RenderedBody)
	require.Equal(t, []SMSBroadcastTemplateVarRow{
		{Key: "name", Value: "Alice"},
		{Key: "window", Value: "tonight"},
	}, campaign.TemplateVarRows)
	require.Equal(t, map[string]string{
		"name":   "Alice",
		"window": "tonight",
	}, campaign.TemplateVars)
}

func TestSMSBroadcastService_CreateCampaignNormalizesTemplateVarSources(t *testing.T) {
	repo := &smsBroadcastRepoStub{}
	svc := NewSMSBroadcastService(nil, nil, nil, nil, repo)

	campaign, err := svc.CreateCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title:      "Maintenance",
		TemplateID: "broadcast-template",
		VarRows: []SMSBroadcastTemplateVarRow{
			{Key: " phone ", Source: " phone_number "},
			{Key: "email", Source: "email"},
			{Key: "name", Source: "username"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, []SMSBroadcastTemplateVarRow{
		{Key: "phone", Source: SMSBroadcastTemplateVarSourcePhoneNumber},
		{Key: "email", Source: SMSBroadcastTemplateVarSourceEmail},
		{Key: "name", Source: SMSBroadcastTemplateVarSourceUsername},
	}, campaign.TemplateVarRows)
	require.Empty(t, campaign.TemplateVars)
}

func TestSMSBroadcastService_CreateCampaignRejectsInvalidTemplateVarRows(t *testing.T) {
	tests := []struct {
		name string
		rows []SMSBroadcastTemplateVarRow
		want error
	}{
		{
			name: "missing key",
			rows: []SMSBroadcastTemplateVarRow{{Value: "Alice"}},
			want: ErrSMSBroadcastTemplateVarKeyRequired,
		},
		{
			name: "missing value",
			rows: []SMSBroadcastTemplateVarRow{{Key: "name"}},
			want: ErrSMSBroadcastTemplateVarValueRequired,
		},
		{
			name: "duplicate key",
			rows: []SMSBroadcastTemplateVarRow{{Key: "name", Value: "Alice"}, {Key: " name ", Value: "Bob"}},
			want: ErrSMSBroadcastTemplateVarKeyDuplicate,
		},
		{
			name: "invalid source",
			rows: []SMSBroadcastTemplateVarRow{{Key: "name", Source: "balance"}},
			want: ErrSMSBroadcastTemplateVarSourceInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSMSBroadcastService(nil, nil, nil, nil, nil)

			_, err := svc.CreateCampaign(context.Background(), &SMSBroadcastCampaignInput{
				Title:      "Maintenance",
				TemplateID: "broadcast-template",
				VarRows:    tt.rows,
			})

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestSMSBroadcastService_CreateAndQueueRejectsEmptyUserFieldSource(t *testing.T) {
	userRepo := &smsBroadcastUserRepoStub{
		users: []User{
			{ID: 1, PhoneNumber: "13800138000", Email: "", Username: "alice", Status: StatusActive, Role: RoleUser},
		},
	}
	svc := NewSMSBroadcastService(userRepo, &smsBroadcastSMSCacheStub{}, &smsBroadcastProviderStub{}, nil, &smsBroadcastRepoStub{})

	_, err := svc.CreateAndQueueCampaign(context.Background(), &SMSBroadcastCampaignInput{
		Title:      "Maintenance",
		TemplateID: "broadcast-template",
		Audience:   SMSBroadcastAudienceFilters{UserIDs: []int64{1}},
		VarRows:    []SMSBroadcastTemplateVarRow{{Key: "email", Source: SMSBroadcastTemplateVarSourceEmail}},
	})

	require.ErrorIs(t, err, ErrSMSBroadcastTemplateVarUserFieldEmpty)
}

func TestSMSBroadcastService_ExecuteCampaignUsesCampaignTemplateIDAndOrderedVarRows(t *testing.T) {
	provider := &smsBroadcastProviderStub{}
	repo := &smsBroadcastRepoStub{
		campaign: &SMSBroadcastCampaign{
			ID:         12,
			Title:      "Maintenance",
			Mode:       SMSBroadcastModeTemplate,
			TemplateID: "broadcast-template",
			Status:     SMSBroadcastStatusQueued,
			TemplateVarRows: []SMSBroadcastTemplateVarRow{
				{Key: "order", Value: "20180515006"},
				{Key: "name", Value: "Alice"},
				{Key: "amount", Value: "100元"},
			},
		},
		recipients: []SMSBroadcastRecipient{{
			UserID:      1,
			PhoneNumber: "+8613800138000",
			RawPhone:    "13800138000",
		}},
	}
	svc := NewSMSBroadcastService(nil, &smsBroadcastSMSCacheStub{}, provider, nil, repo)

	err := svc.executeCampaign(context.Background(), 12)

	require.NoError(t, err)
	require.Len(t, provider.sent, 1)
	require.Equal(t, "+8613800138000", provider.sent[0].phone)
	require.Equal(t, "broadcast-template", provider.sent[0].templateID)
	require.Equal(t, "20180515006|Alice|100元", provider.sent[0].body)
	require.Equal(t, "succeeded", repo.recipients[0].Status)
	require.Empty(t, repo.recipients[0].RenderedBody)
}

func TestSMSBroadcastService_ExecuteCampaignResolvesTemplateVarSourcesPerRecipient(t *testing.T) {
	provider := &smsBroadcastProviderStub{}
	repo := &smsBroadcastRepoStub{
		campaign: &SMSBroadcastCampaign{
			ID:         12,
			Title:      "Maintenance",
			Mode:       SMSBroadcastModeTemplate,
			TemplateID: "broadcast-template",
			Status:     SMSBroadcastStatusQueued,
			TemplateVarRows: []SMSBroadcastTemplateVarRow{
				{Key: "phone", Source: SMSBroadcastTemplateVarSourcePhoneNumber},
				{Key: "email", Source: SMSBroadcastTemplateVarSourceEmail},
				{Key: "name", Source: SMSBroadcastTemplateVarSourceUsername},
			},
		},
		recipients: []SMSBroadcastRecipient{
			{
				UserID:      1,
				PhoneNumber: "+8613800138000",
				RawPhone:    "13800138000",
				User:        User{ID: 1, PhoneNumber: "13800138000", Email: "alice@example.com", Username: "alice"},
			},
			{
				UserID:      2,
				PhoneNumber: "+8613900139000",
				RawPhone:    "13900139000",
				User:        User{ID: 2, PhoneNumber: "13900139000", Email: "bob@example.com", Username: "bob"},
			},
		},
	}
	svc := NewSMSBroadcastService(nil, &smsBroadcastSMSCacheStub{}, provider, nil, repo)

	err := svc.executeCampaign(context.Background(), 12)

	require.NoError(t, err)
	require.Len(t, provider.sent, 2)
	require.Equal(t, "+8613800138000|alice@example.com|alice", provider.sent[0].body)
	require.Equal(t, "+8613900139000|bob@example.com|bob", provider.sent[1].body)
}

func TestSMSBroadcastService_ListRecipientsPaginatedPassesStatus(t *testing.T) {
	repo := &smsBroadcastRepoStub{
		recipientsPageFn: func(_ context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]SMSBroadcastRecipient, *pagination.PaginationResult, error) {
			require.Equal(t, int64(9), campaignID)
			require.Equal(t, "failed", status)
			require.Equal(t, 2, params.Page)
			require.Equal(t, 10, params.PageSize)
			return []SMSBroadcastRecipient{{UserID: 1, PhoneNumber: "+8613800138000", Status: "failed"}}, &pagination.PaginationResult{Total: 1, Page: 2, PageSize: 10, Pages: 1}, nil
		},
	}
	svc := NewSMSBroadcastService(nil, nil, nil, nil, repo)

	items, page, err := svc.ListRecipientsPaginated(context.Background(), 9, pagination.PaginationParams{Page: 2, PageSize: 10}, "failed")

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), page.Total)
}

func TestSMSBroadcastService_ValidateContentLength(t *testing.T) {
	svc := NewSMSBroadcastService(nil, nil, nil, nil, nil)

	err := svc.ValidateMessageBody(strings.Repeat("a", maxSMSBroadcastMessageRunes+1))
	require.ErrorIs(t, err, ErrSMSBroadcastMessageTooLong)
}
