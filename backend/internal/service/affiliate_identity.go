package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingKeyAffiliateIdentityEnabled = "affiliate_identity_enabled"
	SettingKeyAffiliateIdentityConfig  = "affiliate_identity_config"

	AffiliateIdentityTypeInviter = "inviter"
	AffiliateIdentityTypeInvitee = "invitee"

	affiliateIdentityStatusActive  = "active"
	affiliateIdentityStatusRevoked = "revoked"
)

type AffiliateIdentityConfig struct {
	InviterRateMultiplier         float64  `json:"inviter_rate_multiplier"`
	InviteeRateMultiplier         float64  `json:"invitee_rate_multiplier"`
	DurationHours                 int      `json:"duration_hours"`
	QualifiedInviteeCount         int      `json:"qualified_invitee_count"`
	QualifiedPayAmount            float64  `json:"qualified_pay_amount"`
	EligibleOrderTypes            []string `json:"eligible_order_types"`
	FingerprintEnforcementEnabled bool     `json:"fingerprint_enforcement_enabled"`
	MaxAccountsPerFingerprintHash int      `json:"max_accounts_per_fingerprint_hash"`
}

type AffiliateSignupFingerprintInput struct {
	CompositeHash string            `json:"composite_hash"`
	CanvasHash    string            `json:"canvas_hash"`
	WebGLHash     string            `json:"webgl_hash"`
	Components    map[string]string `json:"components,omitempty"`
}

type AffiliateIdentityState struct {
	UserID          int64     `json:"user_id"`
	Type            string    `json:"type"`
	RateMultiplier  float64   `json:"rate_multiplier"`
	SourceInviterID *int64    `json:"source_inviter_id,omitempty"`
	GrantedAt       time.Time `json:"granted_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          string    `json:"status"`
}

type AffiliateIdentityCandidate struct {
	UserID      int64
	PaidAmount  float64
	RiskFlagged bool
}

func (s *AffiliateService) affiliateIdentityEnabled(ctx context.Context) bool {
	if s == nil || s.settingService == nil {
		return false
	}
	value, err := s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateIdentityEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *AffiliateService) GetAffiliateIdentityConfig(ctx context.Context) (*AffiliateIdentityConfig, bool, error) {
	enabled := s.affiliateIdentityEnabled(ctx)
	raw := ""
	var err error
	if s != nil && s.settingService != nil {
		raw, err = s.settingService.settingRepo.GetValue(ctx, SettingKeyAffiliateIdentityConfig)
	}
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, enabled, fmt.Errorf("get affiliate identity config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg, normErr := normalizeAffiliateIdentityConfig(nil)
		return cfg, enabled, normErr
	}
	var cfg AffiliateIdentityConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, enabled, infraerrors.InternalServer("AFFILIATE_IDENTITY_CONFIG_INVALID", "affiliate identity config is invalid")
	}
	normalized, normErr := normalizeAffiliateIdentityConfig(&cfg)
	if normErr != nil {
		return nil, enabled, normErr
	}
	return normalized, enabled, nil
}

func (s *AffiliateService) UpdateAffiliateIdentityConfig(ctx context.Context, enabled bool, cfg *AffiliateIdentityConfig) (*AffiliateIdentityConfig, error) {
	if s == nil || s.settingService == nil {
		return nil, fmt.Errorf("affiliate service is not initialized")
	}
	normalized, err := normalizeAffiliateIdentityConfig(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal affiliate identity config: %w", err)
	}
	if err := s.settingService.settingRepo.SetMultiple(ctx, map[string]string{
		SettingKeyAffiliateIdentityEnabled: strconvFormatBool(enabled),
		SettingKeyAffiliateIdentityConfig:  string(raw),
	}); err != nil {
		return nil, fmt.Errorf("save affiliate identity config: %w", err)
	}
	return normalized, nil
}

func normalizeAffiliateIdentityConfig(cfg *AffiliateIdentityConfig) (*AffiliateIdentityConfig, error) {
	out := &AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24 * 30,
		QualifiedInviteeCount:         0,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{payment.OrderTypeBalance, payment.OrderTypeSubscription},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	}
	if cfg != nil {
		*out = *cfg
	}
	if out.InviterRateMultiplier <= 0 || math.IsNaN(out.InviterRateMultiplier) || math.IsInf(out.InviterRateMultiplier, 0) {
		return nil, infraerrors.BadRequest("AFFILIATE_IDENTITY_CONFIG_INVALID", "inviter rate multiplier must be > 0")
	}
	if out.InviteeRateMultiplier <= 0 || math.IsNaN(out.InviteeRateMultiplier) || math.IsInf(out.InviteeRateMultiplier, 0) {
		return nil, infraerrors.BadRequest("AFFILIATE_IDENTITY_CONFIG_INVALID", "invitee rate multiplier must be > 0")
	}
	if out.DurationHours <= 0 {
		return nil, infraerrors.BadRequest("AFFILIATE_IDENTITY_CONFIG_INVALID", "duration hours must be > 0")
	}
	if out.QualifiedInviteeCount < 0 {
		return nil, infraerrors.BadRequest("AFFILIATE_IDENTITY_CONFIG_INVALID", "qualified invitee count must be >= 0")
	}
	if out.QualifiedPayAmount < 0 || math.IsNaN(out.QualifiedPayAmount) || math.IsInf(out.QualifiedPayAmount, 0) {
		return nil, infraerrors.BadRequest("AFFILIATE_IDENTITY_CONFIG_INVALID", "qualified pay amount must be >= 0")
	}
	if out.MaxAccountsPerFingerprintHash <= 0 {
		out.MaxAccountsPerFingerprintHash = 1
	}
	types := normalizeAffiliateIdentityOrderTypes(out.EligibleOrderTypes)
	if len(types) == 0 {
		types = []string{payment.OrderTypeBalance, payment.OrderTypeSubscription}
	}
	out.EligibleOrderTypes = types
	return out, nil
}

func normalizeAffiliateIdentityOrderTypes(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		switch strings.TrimSpace(value) {
		case payment.OrderTypeBalance:
			if _, ok := seen[payment.OrderTypeBalance]; !ok {
				seen[payment.OrderTypeBalance] = struct{}{}
				out = append(out, payment.OrderTypeBalance)
			}
		case payment.OrderTypeSubscription:
			if _, ok := seen[payment.OrderTypeSubscription]; !ok {
				seen[payment.OrderTypeSubscription] = struct{}{}
				out = append(out, payment.OrderTypeSubscription)
			}
		}
	}
	sort.Strings(out)
	return out
}

func (s *AffiliateService) RecordSignupFingerprint(ctx context.Context, userID int64, input AffiliateSignupFingerprintInput) error {
	if s == nil || s.identityRepo == nil || userID <= 0 {
		return nil
	}
	cfg, _, err := s.GetAffiliateIdentityConfig(ctx)
	if err != nil {
		return err
	}
	return s.identityRepo.RecordSignupFingerprint(ctx, userID, input, cfg)
}

func (s *AffiliateService) RefreshAffiliateIdentitiesForUser(ctx context.Context, userID int64) error {
	if s == nil || s.identityRepo == nil || userID <= 0 {
		return nil
	}
	inviterID, err := s.identityRepo.GetInviterIDForInvitee(ctx, userID)
	if err != nil {
		return err
	}
	if inviterID == nil || *inviterID <= 0 {
		return nil
	}
	return s.RefreshAffiliateIdentitiesForInviter(ctx, *inviterID)
}

func (s *AffiliateService) RefreshAffiliateIdentitiesForOrderInvitee(ctx context.Context, inviteeUserID int64) error {
	if s == nil || s.identityRepo == nil || inviteeUserID <= 0 {
		return nil
	}
	cfg, enabled, err := s.GetAffiliateIdentityConfig(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	inviterID, err := s.identityRepo.GetInviterIDForInvitee(ctx, inviteeUserID)
	if err != nil {
		return err
	}
	if inviterID == nil || *inviterID <= 0 {
		return nil
	}
	hasInviteeIdentity, err := s.identityRepo.HasIdentity(ctx, inviteeUserID, AffiliateIdentityTypeInvitee)
	if err != nil {
		return err
	}
	if hasInviteeIdentity {
		return nil
	}
	candidate, err := s.identityRepo.GetIdentityCandidate(ctx, inviteeUserID, cfg)
	if err != nil {
		return err
	}
	if candidate == nil || candidate.RiskFlagged || candidate.PaidAmount+1e-9 < cfg.QualifiedPayAmount {
		return nil
	}
	candidates, err := s.identityRepo.ListIdentityCandidates(ctx, *inviterID, cfg)
	if err != nil {
		return err
	}
	qualifiedCount, totalPaid, paidInviteeCount := affiliateIdentityQualificationSnapshotValues(candidates, cfg)
	if qualifiedCount < cfg.QualifiedInviteeCount || totalPaid+1e-9 < cfg.QualifiedPayAmount {
		return nil
	}
	expiresAt := time.Now().UTC().Add(time.Duration(cfg.DurationHours) * time.Hour)
	snapshot := map[string]any{
		"qualified_invitee_count": qualifiedCount,
		"qualified_pay_amount":    totalPaid,
		"paid_invitee_count":      paidInviteeCount,
	}
	if err := s.identityRepo.UpsertIdentity(ctx, *inviterID, AffiliateIdentityTypeInviter, cfg.InviterRateMultiplier, nil, expiresAt, snapshot); err != nil {
		return err
	}
	return s.identityRepo.UpsertIdentity(ctx, inviteeUserID, AffiliateIdentityTypeInvitee, cfg.InviteeRateMultiplier, inviterID, expiresAt, snapshot)
}

func (s *AffiliateService) RefreshAffiliateIdentitiesForInviter(ctx context.Context, inviterID int64) error {
	if s == nil || s.identityRepo == nil || inviterID <= 0 {
		return nil
	}
	cfg, enabled, err := s.GetAffiliateIdentityConfig(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		return s.identityRepo.RevokeSystemIdentitiesForInviter(ctx, inviterID)
	}
	candidates, err := s.identityRepo.ListIdentityCandidates(ctx, inviterID, cfg)
	if err != nil {
		return err
	}
	qualifiedCount, totalPaid, _ := affiliateIdentityQualificationSnapshotValues(candidates, cfg)
	paidInvitees := make([]AffiliateIdentityCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.RiskFlagged {
			continue
		}
		if candidate.PaidAmount+1e-9 >= cfg.QualifiedPayAmount {
			paidInvitees = append(paidInvitees, candidate)
		}
	}
	if qualifiedCount < cfg.QualifiedInviteeCount || totalPaid+1e-9 < cfg.QualifiedPayAmount {
		return s.identityRepo.RevokeSystemIdentitiesForInviter(ctx, inviterID)
	}
	expiresAt := time.Now().UTC().Add(time.Duration(cfg.DurationHours) * time.Hour)
	snapshot := map[string]any{
		"qualified_invitee_count": qualifiedCount,
		"qualified_pay_amount":    totalPaid,
		"paid_invitee_count":      len(paidInvitees),
	}
	if err := s.identityRepo.UpsertIdentity(ctx, inviterID, AffiliateIdentityTypeInviter, cfg.InviterRateMultiplier, nil, expiresAt, snapshot); err != nil {
		return err
	}
	for _, invitee := range paidInvitees {
		sourceInviterID := inviterID
		if err := s.identityRepo.UpsertIdentity(ctx, invitee.UserID, AffiliateIdentityTypeInvitee, cfg.InviteeRateMultiplier, &sourceInviterID, expiresAt, snapshot); err != nil {
			return err
		}
	}
	return s.identityRepo.RevokeStaleInviteeIdentities(ctx, inviterID, paidInvitees)
}

func affiliateIdentityQualificationSnapshotValues(candidates []AffiliateIdentityCandidate, cfg *AffiliateIdentityConfig) (int, float64, int) {
	qualifiedCount := 0
	totalPaid := 0.0
	paidInviteeCount := 0
	threshold := 0.0
	if cfg != nil {
		threshold = cfg.QualifiedPayAmount
	}
	for _, candidate := range candidates {
		if candidate.RiskFlagged {
			continue
		}
		qualifiedCount++
		totalPaid += candidate.PaidAmount
		if candidate.PaidAmount+1e-9 >= threshold {
			paidInviteeCount++
		}
	}
	return qualifiedCount, totalPaid, paidInviteeCount
}

func (s *AffiliateService) GetActiveAffiliateIdentity(ctx context.Context, userID int64) (*AffiliateIdentityState, error) {
	if s == nil || s.identityRepo == nil || userID <= 0 {
		return nil, nil
	}
	if !s.affiliateIdentityEnabled(ctx) {
		return nil, nil
	}
	return s.identityRepo.GetActiveIdentity(ctx, userID)
}

func (s *AffiliateService) ResolveAffiliateIdentityMultiplier(ctx context.Context, userID int64, currentMultiplier float64) float64 {
	if currentMultiplier <= 0 || math.IsNaN(currentMultiplier) || math.IsInf(currentMultiplier, 0) {
		currentMultiplier = 1
	}
	state, err := s.GetActiveAffiliateIdentity(ctx, userID)
	if err != nil || state == nil || state.RateMultiplier <= 0 {
		return currentMultiplier
	}
	if state.RateMultiplier < currentMultiplier {
		return state.RateMultiplier
	}
	return currentMultiplier
}
