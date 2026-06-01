package service

import (
	"context"
	"time"
)

type AffiliateIdentityRepository interface {
	RecordSignupFingerprint(ctx context.Context, userID int64, input AffiliateSignupFingerprintInput, cfg *AffiliateIdentityConfig) error
	GetInviterIDForInvitee(ctx context.Context, inviteeUserID int64) (*int64, error)
	ListIdentityCandidates(ctx context.Context, inviterID int64, cfg *AffiliateIdentityConfig) ([]AffiliateIdentityCandidate, error)
	UpsertIdentity(ctx context.Context, userID int64, identityType string, rateMultiplier float64, sourceInviterID *int64, expiresAt time.Time, snapshot map[string]any) error
	RevokeSystemIdentitiesForInviter(ctx context.Context, inviterID int64) error
	RevokeStaleInviteeIdentities(ctx context.Context, inviterID int64, paidInvitees []AffiliateIdentityCandidate) error
	GetActiveIdentity(ctx context.Context, userID int64) (*AffiliateIdentityState, error)
}
