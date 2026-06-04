//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestAffiliateIdentityQualificationGrantsOnlyPaidInvitees(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedInviteeCount:         0,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(1)
	inviteePaidID := int64(2)
	inviteeFreeID := int64(3)
	inviteeSubOnlyID := int64(4)

	repo.inviterByInvitee[inviteePaidID] = inviterID
	repo.inviterByInvitee[inviteeFreeID] = inviterID
	repo.inviterByInvitee[inviteeSubOnlyID] = inviterID

	require.NoError(t, svc.RecordSignupFingerprint(ctx, inviteePaidID, AffiliateSignupFingerprintInput{CompositeHash: "fp-paid", CanvasHash: "canvas-paid", WebGLHash: "webgl-paid"}))
	require.NoError(t, svc.RecordSignupFingerprint(ctx, inviteeFreeID, AffiliateSignupFingerprintInput{CompositeHash: "fp-free", CanvasHash: "canvas-free", WebGLHash: "webgl-free"}))
	require.NoError(t, svc.RecordSignupFingerprint(ctx, inviteeSubOnlyID, AffiliateSignupFingerprintInput{CompositeHash: "fp-sub", CanvasHash: "canvas-sub", WebGLHash: "webgl-sub"}))

	repo.paid[inviteePaidID] = 50
	repo.paid[inviteeSubOnlyID] = 0

	require.NoError(t, svc.RefreshAffiliateIdentitiesForInviter(ctx, inviterID))

	inviterState, err := svc.GetActiveAffiliateIdentity(ctx, inviterID)
	require.NoError(t, err)
	require.NotNil(t, inviterState)
	require.Equal(t, AffiliateIdentityTypeInviter, inviterState.Type)
	require.InDelta(t, 1.5, inviterState.RateMultiplier, 1e-9)

	paidState, err := svc.GetActiveAffiliateIdentity(ctx, inviteePaidID)
	require.NoError(t, err)
	require.NotNil(t, paidState)
	require.Equal(t, AffiliateIdentityTypeInvitee, paidState.Type)
	require.InDelta(t, 1.4, paidState.RateMultiplier, 1e-9)

	freeState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeFreeID)
	require.NoError(t, err)
	require.Nil(t, freeState)

	subOnlyState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeSubOnlyID)
	require.NoError(t, err)
	require.Nil(t, subOnlyState)
}

func TestAffiliateIdentityInviteeMustMeetOwnPayThreshold(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedInviteeCount:         0,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(5)
	inviteeAID := int64(6)
	inviteeBID := int64(7)

	repo.inviterByInvitee[inviteeAID] = inviterID
	repo.inviterByInvitee[inviteeBID] = inviterID
	repo.paid[inviteeAID] = 25
	repo.paid[inviteeBID] = 25

	require.NoError(t, svc.RefreshAffiliateIdentitiesForInviter(ctx, inviterID))

	inviterState, err := svc.GetActiveAffiliateIdentity(ctx, inviterID)
	require.NoError(t, err)
	require.NotNil(t, inviterState)
	require.Equal(t, AffiliateIdentityTypeInviter, inviterState.Type)
	require.InDelta(t, 1.5, inviterState.RateMultiplier, 1e-9)

	aState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeAID)
	require.NoError(t, err)
	require.Nil(t, aState)

	bState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeBID)
	require.NoError(t, err)
	require.Nil(t, bState)
}

func TestAffiliateIdentityQualificationMatrix(t *testing.T) {
	ctx := context.Background()
	type testCase struct {
		name            string
		aPaid           float64
		bPaid           float64
		aRisk           bool
		bRisk           bool
		seedActive      bool
		wantInviter     bool
		wantAInvitee    bool
		wantBInvitee    bool
		wantInviterRate float64
	}
	cases := []testCase{
		{
			name:            "single invitee reaches threshold",
			aPaid:           50,
			wantInviter:     true,
			wantAInvitee:    true,
			wantInviterRate: 1.5,
		},
		{
			name:            "two invitees cumulatively reach threshold but neither gets invitee identity",
			aPaid:           25,
			bPaid:           25,
			wantInviter:     true,
			wantInviterRate: 1.5,
		},
		{
			name:  "cumulative payment below threshold grants nobody",
			aPaid: 20,
			bPaid: 20,
		},
		{
			name:            "both invitees individually reach threshold",
			aPaid:           50,
			bPaid:           50,
			wantInviter:     true,
			wantAInvitee:    true,
			wantBInvitee:    true,
			wantInviterRate: 1.5,
		},
		{
			name:            "one invitee exceeds threshold and the other does not",
			aPaid:           80,
			bPaid:           10,
			wantInviter:     true,
			wantAInvitee:    true,
			wantInviterRate: 1.5,
		},
		{
			name:            "risk invitee is excluded from own identity and inviter total",
			aPaid:           50,
			bPaid:           50,
			bRisk:           true,
			wantInviter:     true,
			wantAInvitee:    true,
			wantInviterRate: 1.5,
		},
		{
			name:  "only risk invitee reaches threshold",
			bPaid: 50,
			bRisk: true,
		},
		{
			name:            "subscription effective payment counts when repository includes it",
			aPaid:           50,
			wantInviter:     true,
			wantAInvitee:    true,
			wantInviterRate: 1.5,
		},
		{
			name: "pending payment is excluded by repository effective amount",
		},
		{
			name:       "refund below threshold revokes seeded active identities",
			aPaid:      20,
			seedActive: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newAffiliateIdentityMemoryRepo()
			settingRepo := &settingPublicRepoStub{values: map[string]string{}}
			settingSvc := NewSettingService(settingRepo, nil)
			svc := NewAffiliateService(nil, settingSvc, nil, nil)
			svc.SetIdentityRepository(repo)

			_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
				InviterRateMultiplier:         1.5,
				InviteeRateMultiplier:         1.4,
				DurationHours:                 24,
				QualifiedInviteeCount:         0,
				QualifiedPayAmount:            50,
				EligibleOrderTypes:            []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
				FingerprintEnforcementEnabled: true,
				MaxAccountsPerFingerprintHash: 3,
			})
			require.NoError(t, err)

			inviterID := int64(30)
			inviteeAID := int64(31)
			inviteeBID := int64(32)
			repo.inviterByInvitee[inviteeAID] = inviterID
			repo.inviterByInvitee[inviteeBID] = inviterID
			repo.paid[inviteeAID] = tc.aPaid
			repo.paid[inviteeBID] = tc.bPaid
			repo.risk[inviteeAID] = tc.aRisk
			repo.risk[inviteeBID] = tc.bRisk

			if tc.seedActive {
				expiresAt := time.Now().Add(time.Hour)
				require.NoError(t, repo.UpsertIdentity(ctx, inviterID, AffiliateIdentityTypeInviter, 1.5, nil, expiresAt, nil))
				require.NoError(t, repo.UpsertIdentity(ctx, inviteeAID, AffiliateIdentityTypeInvitee, 1.4, &inviterID, expiresAt, nil))
			}

			require.NoError(t, svc.RefreshAffiliateIdentitiesForInviter(ctx, inviterID))

			inviterState, err := svc.GetActiveAffiliateIdentity(ctx, inviterID)
			require.NoError(t, err)
			if tc.wantInviter {
				require.NotNil(t, inviterState)
				require.Equal(t, AffiliateIdentityTypeInviter, inviterState.Type)
				require.InDelta(t, tc.wantInviterRate, inviterState.RateMultiplier, 1e-9)
			} else {
				require.Nil(t, inviterState)
			}

			aState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeAID)
			require.NoError(t, err)
			if tc.wantAInvitee {
				require.NotNil(t, aState)
				require.Equal(t, AffiliateIdentityTypeInvitee, aState.Type)
				require.InDelta(t, 1.4, aState.RateMultiplier, 1e-9)
			} else {
				require.Nil(t, aState)
			}

			bState, err := svc.GetActiveAffiliateIdentity(ctx, inviteeBID)
			require.NoError(t, err)
			if tc.wantBInvitee {
				require.NotNil(t, bState)
				require.Equal(t, AffiliateIdentityTypeInvitee, bState.Type)
				require.InDelta(t, 1.4, bState.RateMultiplier, 1e-9)
			} else {
				require.Nil(t, bState)
			}
		})
	}
}

func TestAffiliateIdentityFingerprintRiskExcludesDuplicateDevice(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 1,
	})
	require.NoError(t, err)

	inviterID := int64(10)
	firstID := int64(11)
	secondID := int64(12)

	repo.inviterByInvitee[firstID] = inviterID
	repo.inviterByInvitee[secondID] = inviterID

	require.NoError(t, svc.RecordSignupFingerprint(ctx, firstID, AffiliateSignupFingerprintInput{CompositeHash: "same-device"}))
	require.NoError(t, svc.RecordSignupFingerprint(ctx, secondID, AffiliateSignupFingerprintInput{CompositeHash: "same-device"}))
	repo.paid[firstID] = 50
	repo.paid[secondID] = 50

	require.NoError(t, svc.RefreshAffiliateIdentitiesForInviter(ctx, inviterID))

	inviterState, err := svc.GetActiveAffiliateIdentity(ctx, inviterID)
	require.NoError(t, err)
	require.NotNil(t, inviterState)

	firstState, err := svc.GetActiveAffiliateIdentity(ctx, firstID)
	require.NoError(t, err)
	require.NotNil(t, firstState)

	secondState, err := svc.GetActiveAffiliateIdentity(ctx, secondID)
	require.NoError(t, err)
	require.Nil(t, secondState)
}

func TestAffiliateIdentityOrderRefreshGrantsOnlyCurrentInviteeAndInviter(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedInviteeCount:         0,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(40)
	currentInviteeID := int64(41)
	otherQualifiedInviteeID := int64(42)
	repo.inviterByInvitee[currentInviteeID] = inviterID
	repo.inviterByInvitee[otherQualifiedInviteeID] = inviterID
	repo.paid[currentInviteeID] = 50
	repo.paid[otherQualifiedInviteeID] = 100

	require.NoError(t, svc.RefreshAffiliateIdentitiesForOrderInvitee(ctx, currentInviteeID))

	require.Equal(t, 1, repo.upsertCount[inviterID])
	require.Equal(t, 1, repo.upsertCount[currentInviteeID])
	require.Zero(t, repo.upsertCount[otherQualifiedInviteeID])

	inviterState, err := svc.GetActiveAffiliateIdentity(ctx, inviterID)
	require.NoError(t, err)
	require.NotNil(t, inviterState)
	require.Equal(t, AffiliateIdentityTypeInviter, inviterState.Type)

	currentState, err := svc.GetActiveAffiliateIdentity(ctx, currentInviteeID)
	require.NoError(t, err)
	require.NotNil(t, currentState)
	require.Equal(t, AffiliateIdentityTypeInvitee, currentState.Type)

	otherState, err := svc.GetActiveAffiliateIdentity(ctx, otherQualifiedInviteeID)
	require.NoError(t, err)
	require.Nil(t, otherState)
}

func TestAffiliateIdentityOrderRefreshSkipsInviteeBelowThreshold(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(50)
	inviteeID := int64(51)
	repo.inviterByInvitee[inviteeID] = inviterID
	repo.paid[inviteeID] = 49.99

	require.NoError(t, svc.RefreshAffiliateIdentitiesForOrderInvitee(ctx, inviteeID))

	require.Zero(t, repo.upsertCount[inviterID])
	require.Zero(t, repo.upsertCount[inviteeID])
}

func TestAffiliateIdentityOrderRefreshDoesNotRefreshPreviouslyGrantedInvitee(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(60)
	inviteeID := int64(61)
	repo.inviterByInvitee[inviteeID] = inviterID
	repo.paid[inviteeID] = 50
	expiredAt := time.Now().Add(-time.Hour)
	require.NoError(t, repo.UpsertIdentity(ctx, inviteeID, AffiliateIdentityTypeInvitee, 1.4, &inviterID, expiredAt, nil))
	repo.upsertCount[inviteeID] = 0

	require.NoError(t, svc.RefreshAffiliateIdentitiesForOrderInvitee(ctx, inviteeID))

	require.Zero(t, repo.upsertCount[inviterID])
	require.Zero(t, repo.upsertCount[inviteeID])
}

func TestAffiliateIdentityOrderRefreshSkipsRiskFlaggedInvitee(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	inviterID := int64(70)
	inviteeID := int64(71)
	repo.inviterByInvitee[inviteeID] = inviterID
	repo.paid[inviteeID] = 50
	repo.risk[inviteeID] = true

	require.NoError(t, svc.RefreshAffiliateIdentitiesForOrderInvitee(ctx, inviteeID))

	require.Zero(t, repo.upsertCount[inviterID])
	require.Zero(t, repo.upsertCount[inviteeID])
}

func TestAffiliateIdentityMultiplierUsesMostFavorableRate(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateIdentityMemoryRepo()
	settingRepo := &settingPublicRepoStub{values: map[string]string{}}
	settingSvc := NewSettingService(settingRepo, nil)
	svc := NewAffiliateService(nil, settingSvc, nil, nil)
	svc.SetIdentityRepository(repo)

	_, err := svc.UpdateAffiliateIdentityConfig(ctx, true, &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance},
		MaxAccountsPerFingerprintHash: 3,
	})
	require.NoError(t, err)

	userID := int64(20)
	require.NoError(t, repo.UpsertIdentity(ctx, userID, AffiliateIdentityTypeInviter, 1.5, nil, time.Now().Add(time.Hour), nil))

	rate := svc.ResolveAffiliateIdentityMultiplier(ctx, userID, 1.8)
	require.InDelta(t, 1.5, rate, 1e-9)

	rate = svc.ResolveAffiliateIdentityMultiplier(ctx, userID, 1.2)
	require.InDelta(t, 1.2, rate, 1e-9)
}
