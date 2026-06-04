//go:build unit

package service

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

type affiliateRepoStub struct {
	summaries       map[int64]*AffiliateSummary
	invitees        []AffiliateInvitee
	transferAmount  float64
	transferBalance float64
	transferCalls   []struct {
		userID     int64
		multiplier float64
	}
	setRelationCalls []struct {
		inviterUserID int64
		inviteeUserID int64
		overwrite     bool
	}
	setRelationResult *AffiliateInviteRelationChange
	setRelationErr    error

	signupBonusCalls []struct {
		inviterID     int64
		inviteeUserID int64
		amount        float64
	}
	signupBonusApplied bool
	signupBonusBalance float64
	signupBonusErr     error

	userRecordCalls []struct {
		userID int64
		filter AffiliateRecordFilter
	}
	userRecords      []UserAffiliateRecord
	userRecordsTotal int64
	userRecordsErr   error
}

func (s *affiliateRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := s.summaries[userID]; ok {
		cloned := *summary
		return &cloned, nil
	}
	if s.summaries == nil {
		s.summaries = make(map[int64]*AffiliateSummary)
	}
	summary := &AffiliateSummary{UserID: userID}
	s.summaries[userID] = summary
	cloned := *summary
	return &cloned, nil
}

func (s *affiliateRepoStub) GetAffiliateByCode(_ context.Context, code string) (*AffiliateSummary, error) {
	for _, summary := range s.summaries {
		if summary.AffCode == code {
			cloned := *summary
			return &cloned, nil
		}
	}
	return nil, ErrAffiliateProfileNotFound
}

func (s *affiliateRepoStub) BindInviter(_ context.Context, userID, inviterID int64) (bool, error) {
	invitee, ok := s.summaries[userID]
	if !ok {
		return false, ErrAffiliateProfileNotFound
	}
	if invitee.InviterID != nil {
		return false, nil
	}
	invitee.InviterID = &inviterID
	return true, nil
}

func (s *affiliateRepoStub) AdminSetInviteRelation(_ context.Context, inviterUserID, inviteeUserID int64, overwrite bool) (*AffiliateInviteRelationChange, error) {
	s.setRelationCalls = append(s.setRelationCalls, struct {
		inviterUserID int64
		inviteeUserID int64
		overwrite     bool
	}{inviterUserID: inviterUserID, inviteeUserID: inviteeUserID, overwrite: overwrite})
	return s.setRelationResult, s.setRelationErr
}

func (s *affiliateRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	panic("unexpected AccrueQuota call")
}

func (s *affiliateRepoStub) AwardSignupBonus(_ context.Context, inviterID, inviteeUserID int64, amount float64) (bool, float64, error) {
	s.signupBonusCalls = append(s.signupBonusCalls, struct {
		inviterID     int64
		inviteeUserID int64
		amount        float64
	}{inviterID: inviterID, inviteeUserID: inviteeUserID, amount: amount})
	return s.signupBonusApplied, s.signupBonusBalance, s.signupBonusErr
}

func (s *affiliateRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	panic("unexpected GetAccruedRebateFromInvitee call")
}

func (s *affiliateRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	return 0, nil
}

func (s *affiliateRepoStub) TransferQuotaToBalance(_ context.Context, userID int64, multiplier float64) (float64, float64, error) {
	s.transferCalls = append(s.transferCalls, struct {
		userID     int64
		multiplier float64
	}{userID: userID, multiplier: multiplier})
	return s.transferAmount, s.transferBalance, nil
}

func (s *affiliateRepoStub) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	out := make([]AffiliateInvitee, len(s.invitees))
	copy(out, s.invitees)
	return out, nil
}

func (s *affiliateRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	panic("unexpected UpdateUserAffCode call")
}

func (s *affiliateRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	panic("unexpected ResetUserAffCode call")
}

func (s *affiliateRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	panic("unexpected SetUserRebateRate call")
}

func (s *affiliateRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	panic("unexpected BatchSetUserRebateRate call")
}

func (s *affiliateRepoStub) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	panic("unexpected ListUsersWithCustomSettings call")
}

func (s *affiliateRepoStub) ListInvitersWithInvitees(context.Context, AffiliateAdminFilter) ([]AffiliateInviterEntry, int64, error) {
	panic("unexpected ListInvitersWithInvitees call")
}

func (s *affiliateRepoStub) ListInviteesByInviter(context.Context, int64, int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInviteesByInviter call")
}

func (s *affiliateRepoStub) ListAffiliateInviteRecords(context.Context, AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error) {
	panic("unexpected ListAffiliateInviteRecords call")
}

func (s *affiliateRepoStub) ListAffiliateRebateRecords(context.Context, AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error) {
	panic("unexpected ListAffiliateRebateRecords call")
}

func (s *affiliateRepoStub) ListAffiliateTransferRecords(context.Context, AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error) {
	panic("unexpected ListAffiliateTransferRecords call")
}

func (s *affiliateRepoStub) ListUserAffiliateRecords(_ context.Context, userID int64, filter AffiliateRecordFilter) ([]UserAffiliateRecord, int64, error) {
	s.userRecordCalls = append(s.userRecordCalls, struct {
		userID int64
		filter AffiliateRecordFilter
	}{userID: userID, filter: filter})
	out := make([]UserAffiliateRecord, len(s.userRecords))
	copy(out, s.userRecords)
	return out, s.userRecordsTotal, s.userRecordsErr
}

func (s *affiliateRepoStub) GetAffiliateUserOverview(context.Context, int64) (*AffiliateUserOverview, error) {
	panic("unexpected GetAffiliateUserOverview call")
}

func TestGetAffiliateDetailReturnsCashAliases(t *testing.T) {
	t.Parallel()
	inviterID := int64(7)
	repo := &affiliateRepoStub{
		summaries: map[int64]*AffiliateSummary{
			11: {
				UserID:          11,
				AffCode:         "AFFCASH11",
				InviterID:       &inviterID,
				AffCount:        2,
				AffQuota:        9.9,
				AffFrozenQuota:  1.5,
				AffHistoryQuota: 20.4,
			},
		},
	}
	svc := &AffiliateService{repo: repo}

	detail, err := svc.GetAffiliateDetail(context.Background(), 11)

	require.NoError(t, err)
	require.InDelta(t, 9.9, detail.AffQuota, 1e-9)
	require.InDelta(t, 9.9, detail.RebateCashBalance, 1e-9)
	require.InDelta(t, 1.5, detail.FrozenRebateCash, 1e-9)
	require.InDelta(t, 20.4, detail.TotalRebateCash, 1e-9)
}

func TestListUserAffiliateRecordsNormalizesFilterAndScopesToUser(t *testing.T) {
	t.Parallel()
	orderID := int64(99)
	repo := &affiliateRepoStub{
		userRecords: []UserAffiliateRecord{
			{
				LedgerID:      1,
				Action:        "accrue",
				Amount:        9.9,
				SourceOrderID: &orderID,
			},
		},
		userRecordsTotal: 1,
	}
	svc := &AffiliateService{repo: repo}

	records, total, err := svc.ListUserAffiliateRecords(context.Background(), 11, AffiliateRecordFilter{Page: 0, PageSize: 500})

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, records, 1)
	require.Equal(t, int64(11), repo.userRecordCalls[0].userID)
	require.Equal(t, 1, repo.userRecordCalls[0].filter.Page)
	require.Equal(t, 100, repo.userRecordCalls[0].filter.PageSize)
	require.Equal(t, "accrue", records[0].Action)
	require.InDelta(t, 9.9, records[0].Amount, 1e-9)
}

// TestResolveRebateRatePercent_PerUserOverride verifies that per-inviter
// AffRebateRatePercent overrides the global rate, that NULL falls back to the
// global rate, and that out-of-range exclusive rates are clamped silently.
//
// SettingService is left nil here so globalRebateRatePercent returns the
// documented default (AffiliateRebateRateDefault = 20%) — this exercises the
// fallback path without spinning up a settings stub.
func TestResolveRebateRatePercent_PerUserOverride(t *testing.T) {
	t.Parallel()
	svc := &AffiliateService{}

	// nil exclusive rate → falls back to global default (20%)
	require.InDelta(t, AffiliateRebateRateDefault,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{}), 1e-9)

	// exclusive rate set → overrides global
	rate := 50.0
	require.InDelta(t, 50.0,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &rate}), 1e-9)

	// exclusive rate 0 → returns 0 (no rebate, intentional)
	zero := 0.0
	require.InDelta(t, 0.0,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &zero}), 1e-9)

	// exclusive rate above max → clamped to Max
	tooHigh := 250.0
	require.InDelta(t, AffiliateRebateRateMax,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooHigh}), 1e-9)

	// exclusive rate below min → clamped to Min
	tooLow := -5.0
	require.InDelta(t, AffiliateRebateRateMin,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooLow}), 1e-9)
}

// TestIsEnabled_NilSettingServiceReturnsDefault verifies that IsEnabled
// safely handles a nil settingService dependency by returning the default
// (off). This protects callers from nil-pointer crashes in misconfigured
// environments.
func TestIsEnabled_NilSettingServiceReturnsDefault(t *testing.T) {
	t.Parallel()
	svc := &AffiliateService{}
	require.False(t, svc.IsEnabled(context.Background()))
	require.Equal(t, AffiliateEnabledDefault, svc.IsEnabled(context.Background()))
}

// TestValidateExclusiveRate_BoundaryAndInvalid covers the validator used by
// admin-facing rate setters: nil is always valid (clear), in-range values
// are accepted, NaN/Inf and out-of-range values produce a typed BadRequest.
func TestValidateExclusiveRate_BoundaryAndInvalid(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateExclusiveRate(nil))

	for _, v := range []float64{0, 0.01, 50, 99.99, 100} {
		v := v
		require.NoError(t, validateExclusiveRate(&v), "value %v should be valid", v)
	}

	for _, v := range []float64{-0.01, 100.01, -100, 200} {
		v := v
		require.Error(t, validateExclusiveRate(&v), "value %v should be rejected", v)
	}

	nan := math.NaN()
	require.Error(t, validateExclusiveRate(&nan))
	posInf := math.Inf(1)
	require.Error(t, validateExclusiveRate(&posInf))
	negInf := math.Inf(-1)
	require.Error(t, validateExclusiveRate(&negInf))
}

func TestMaskEmail(t *testing.T) {
	t.Parallel()
	require.Equal(t, "a***@g***.com", maskEmail("alice@gmail.com"))
	require.Equal(t, "x***@d***", maskEmail("x@domain"))
	require.Equal(t, "", maskEmail(""))
}

func TestApplySignupBonus_AwardsDirectBalanceAndReturnsBalance(t *testing.T) {
	t.Parallel()

	inviterID := int64(11)
	inviteeID := int64(22)
	repo := &affiliateRepoStub{
		summaries: map[int64]*AffiliateSummary{
			inviteeID: {UserID: inviteeID, InviterID: &inviterID},
			inviterID: {UserID: inviterID},
		},
		signupBonusApplied: true,
		signupBonusBalance: 66.5,
	}
	settings := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateSignupRewardEnabled: "true",
		SettingKeyAffiliateSignupRewardAmount:  "18.88",
	}}, nil)
	svc := &AffiliateService{repo: repo, settingService: settings}

	amount, balance, err := svc.ApplySignupBonus(context.Background(), inviteeID)
	require.NoError(t, err)
	require.InDelta(t, 18.88, amount, 1e-9)
	require.InDelta(t, 66.5, balance, 1e-9)
	require.Len(t, repo.signupBonusCalls, 1)
	require.Equal(t, inviterID, repo.signupBonusCalls[0].inviterID)
	require.Equal(t, inviteeID, repo.signupBonusCalls[0].inviteeUserID)
	require.InDelta(t, 18.88, repo.signupBonusCalls[0].amount, 1e-9)
}

func TestTransferAffiliateQuota_UsesBalanceRechargeMultiplierForPlatformAmount(t *testing.T) {
	t.Parallel()

	repo := &affiliateRepoStub{
		transferAmount:  10,
		transferBalance: 260,
	}
	settings := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingBalanceRechargeMult: "13",
	}}, nil)
	svc := &AffiliateService{repo: repo, settingService: settings}

	transferredCash, transferredQuota, balance, err := svc.TransferAffiliateQuota(context.Background(), 11)
	require.NoError(t, err)
	require.InDelta(t, 10, transferredCash, 1e-9)
	require.InDelta(t, 130, transferredQuota, 1e-9)
	require.InDelta(t, 260, balance, 1e-9)
	require.Len(t, repo.transferCalls, 1)
	require.Equal(t, int64(11), repo.transferCalls[0].userID)
	require.InDelta(t, 13, repo.transferCalls[0].multiplier, 1e-9)
}

func TestAdminSetInviteRelationDelegatesOverwrite(t *testing.T) {
	t.Parallel()

	previousID := int64(33)
	repo := &affiliateRepoStub{
		setRelationResult: &AffiliateInviteRelationChange{
			InviterUserID:         11,
			InviteeUserID:         22,
			Overwritten:           true,
			PreviousInviterUserID: &previousID,
		},
	}
	svc := &AffiliateService{repo: repo}

	result, err := svc.AdminSetInviteRelation(context.Background(), 11, 22, true)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(11), result.InviterUserID)
	require.Equal(t, int64(22), result.InviteeUserID)
	require.True(t, result.Overwritten)
	require.Equal(t, previousID, *result.PreviousInviterUserID)
	require.Len(t, repo.setRelationCalls, 1)
	require.Equal(t, int64(11), repo.setRelationCalls[0].inviterUserID)
	require.Equal(t, int64(22), repo.setRelationCalls[0].inviteeUserID)
	require.True(t, repo.setRelationCalls[0].overwrite)
}

func TestAdminSetInviteRelationRejectsInvalidIDs(t *testing.T) {
	t.Parallel()

	svc := &AffiliateService{repo: &affiliateRepoStub{}}
	_, err := svc.AdminSetInviteRelation(context.Background(), 0, 22, true)
	require.Error(t, err)

	_, err = svc.AdminSetInviteRelation(context.Background(), 11, 0, true)
	require.Error(t, err)

	_, err = svc.AdminSetInviteRelation(context.Background(), 11, 11, true)
	require.Error(t, err)
}

func TestApplySignupBonus_SkipsWhenFeatureDisabledOrNoInviter(t *testing.T) {
	t.Parallel()

	inviteeID := int64(33)
	repo := &affiliateRepoStub{
		summaries: map[int64]*AffiliateSummary{
			inviteeID: {UserID: inviteeID},
		},
	}
	settings := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateSignupRewardEnabled: "false",
		SettingKeyAffiliateSignupRewardAmount:  "10",
	}}, nil)
	svc := &AffiliateService{repo: repo, settingService: settings}

	amount, balance, err := svc.ApplySignupBonus(context.Background(), inviteeID)
	require.NoError(t, err)
	require.InDelta(t, 0.0, amount, 1e-9)
	require.InDelta(t, 0.0, balance, 1e-9)
	require.Empty(t, repo.signupBonusCalls)

	settings = NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateSignupRewardEnabled: "true",
		SettingKeyAffiliateSignupRewardAmount:  "10",
	}}, nil)
	svc = &AffiliateService{repo: repo, settingService: settings}
	amount, balance, err = svc.ApplySignupBonus(context.Background(), inviteeID)
	require.NoError(t, err)
	require.InDelta(t, 0.0, amount, 1e-9)
	require.InDelta(t, 0.0, balance, 1e-9)
	require.Empty(t, repo.signupBonusCalls)
}

func TestIsValidAffiliateCodeFormat(t *testing.T) {
	t.Parallel()

	// 邀请码格式校验同时服务于：
	// 1) 系统自动生成的 12 位随机码（A-Z 去 I/O，2-9 去 0/1）
	// 2) 管理员设置的自定义专属码（如 "VIP2026"、"NEW_USER-1"）
	// 因此校验放宽到 [A-Z0-9_-]{4,32}（要求调用方先 ToUpper）。
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid canonical 12-char", "ABCDEFGHJKLM", true},
		{"valid all digits 2-9", "234567892345", true},
		{"valid mixed", "A2B3C4D5E6F7", true},
		{"valid admin custom short", "VIP1", true},
		{"valid admin custom with hyphen", "NEW-USER", true},
		{"valid admin custom with underscore", "VIP_2026", true},
		{"valid 32-char max", "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345", true},
		// Previously-excluded chars (I/O/0/1) are now allowed since admins may use them.
		{"letter I now allowed", "IBCDEFGHJKLM", true},
		{"letter O now allowed", "OBCDEFGHJKLM", true},
		{"digit 0 now allowed", "0BCDEFGHJKLM", true},
		{"digit 1 now allowed", "1BCDEFGHJKLM", true},
		{"too short (3 chars)", "ABC", false},
		{"too long (33 chars)", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456", false},
		{"lowercase rejected (caller must ToUpper first)", "abcdefghjklm", false},
		{"empty", "", false},
		{"utf8 non-ascii", "ÄÄÄÄÄÄ", false}, // bytes out of charset
		{"ascii punctuation .", "ABCDEFGHJK.M", false},
		{"whitespace", "ABCDEFGHJK M", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, isValidAffiliateCodeFormat(tc.in))
		})
	}
}
