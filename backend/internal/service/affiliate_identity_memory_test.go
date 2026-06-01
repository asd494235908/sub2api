//go:build unit

package service

import (
	"context"
	"time"
)

type affiliateIdentityMemoryRepo struct {
	inviterByInvitee map[int64]int64
	fingerprints     map[int64]AffiliateSignupFingerprintInput
	risk             map[int64]bool
	paid             map[int64]float64
	identities       map[int64]*AffiliateIdentityState
}

func newAffiliateIdentityMemoryRepo() *affiliateIdentityMemoryRepo {
	return &affiliateIdentityMemoryRepo{
		inviterByInvitee: map[int64]int64{},
		fingerprints:     map[int64]AffiliateSignupFingerprintInput{},
		risk:             map[int64]bool{},
		paid:             map[int64]float64{},
		identities:       map[int64]*AffiliateIdentityState{},
	}
}

func (r *affiliateIdentityMemoryRepo) RecordSignupFingerprint(_ context.Context, userID int64, input AffiliateSignupFingerprintInput, cfg *AffiliateIdentityConfig) error {
	duplicateCount := 0
	for _, existing := range r.fingerprints {
		if existing.CompositeHash == input.CompositeHash && input.CompositeHash != "" {
			duplicateCount++
		}
	}
	r.fingerprints[userID] = input
	limit := 1
	if cfg != nil && cfg.MaxAccountsPerFingerprintHash > 0 {
		limit = cfg.MaxAccountsPerFingerprintHash
	}
	r.risk[userID] = cfg != nil && cfg.FingerprintEnforcementEnabled && duplicateCount >= limit
	return nil
}

func (r *affiliateIdentityMemoryRepo) GetInviterIDForInvitee(_ context.Context, inviteeUserID int64) (*int64, error) {
	id, ok := r.inviterByInvitee[inviteeUserID]
	if !ok {
		return nil, nil
	}
	return &id, nil
}

func (r *affiliateIdentityMemoryRepo) ListIdentityCandidates(_ context.Context, inviterID int64, _ *AffiliateIdentityConfig) ([]AffiliateIdentityCandidate, error) {
	out := []AffiliateIdentityCandidate{}
	for inviteeID, id := range r.inviterByInvitee {
		if id != inviterID {
			continue
		}
		out = append(out, AffiliateIdentityCandidate{UserID: inviteeID, PaidAmount: r.paid[inviteeID], RiskFlagged: r.risk[inviteeID]})
	}
	return out, nil
}

func (r *affiliateIdentityMemoryRepo) UpsertIdentity(_ context.Context, userID int64, identityType string, rateMultiplier float64, sourceInviterID *int64, expiresAt time.Time, _ map[string]any) error {
	source := (*int64)(nil)
	if sourceInviterID != nil {
		v := *sourceInviterID
		source = &v
	}
	r.identities[userID] = &AffiliateIdentityState{
		UserID:          userID,
		Type:            identityType,
		RateMultiplier:  rateMultiplier,
		SourceInviterID: source,
		GrantedAt:       time.Now().UTC(),
		ExpiresAt:       expiresAt,
		Status:          affiliateIdentityStatusActive,
	}
	return nil
}

func (r *affiliateIdentityMemoryRepo) RevokeSystemIdentitiesForInviter(_ context.Context, inviterID int64) error {
	delete(r.identities, inviterID)
	for inviteeID, id := range r.inviterByInvitee {
		if id == inviterID {
			delete(r.identities, inviteeID)
		}
	}
	return nil
}

func (r *affiliateIdentityMemoryRepo) RevokeStaleInviteeIdentities(_ context.Context, inviterID int64, paidInvitees []AffiliateIdentityCandidate) error {
	keep := map[int64]struct{}{}
	for _, invitee := range paidInvitees {
		keep[invitee.UserID] = struct{}{}
	}
	for inviteeID, id := range r.inviterByInvitee {
		if id != inviterID {
			continue
		}
		if _, ok := keep[inviteeID]; !ok {
			delete(r.identities, inviteeID)
		}
	}
	return nil
}

func (r *affiliateIdentityMemoryRepo) GetActiveIdentity(_ context.Context, userID int64) (*AffiliateIdentityState, error) {
	state := r.identities[userID]
	if state == nil || state.Status != affiliateIdentityStatusActive || !state.ExpiresAt.After(time.Now()) {
		return nil, nil
	}
	copy := *state
	return &copy, nil
}
