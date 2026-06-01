package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type affiliateIdentityRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewAffiliateIdentityRepository(client *dbent.Client, db *sql.DB) service.AffiliateIdentityRepository {
	return &affiliateIdentityRepository{client: client, db: db}
}

func (r *affiliateIdentityRepository) execer(ctx context.Context) (affiliateQueryExecer, error) {
	if r == nil {
		return nil, fmt.Errorf("affiliate identity repository unavailable")
	}
	if client := clientFromContext(ctx, r.client); client != nil {
		return client, nil
	}
	if r.db != nil {
		return r.db, nil
	}
	return nil, fmt.Errorf("affiliate identity repository unavailable")
}

func (r *affiliateIdentityRepository) RecordSignupFingerprint(ctx context.Context, userID int64, input service.AffiliateSignupFingerprintInput, cfg *service.AffiliateIdentityConfig) error {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return err
	}
	composite := strings.TrimSpace(input.CompositeHash)
	canvas := strings.TrimSpace(input.CanvasHash)
	webgl := strings.TrimSpace(input.WebGLHash)
	if composite == "" && canvas == "" && webgl == "" {
		composite = fmt.Sprintf("missing:%d", userID)
	}

	duplicateCount := 0
	if composite != "" {
		rows, err := execCtx.QueryContext(ctx, `SELECT COUNT(*) FROM user_signup_fingerprints WHERE composite_hash = $1 AND user_id <> $2`, composite, userID)
		if err != nil {
			return err
		}
		if rows.Next() {
			if err := rows.Scan(&duplicateCount); err != nil {
				_ = rows.Close()
				return err
			}
		}
		if err := rows.Close(); err != nil {
			return err
		}
	}

	limit := 1
	if cfg != nil && cfg.MaxAccountsPerFingerprintHash > 0 {
		limit = cfg.MaxAccountsPerFingerprintHash
	}
	riskFlagged := cfg != nil && cfg.FingerprintEnforcementEnabled && composite != "" && duplicateCount >= limit
	riskReason := ""
	if riskFlagged {
		riskReason = "duplicate_device"
	}
	componentsRaw, _ := json.Marshal(input.Components)
	if string(componentsRaw) == "null" {
		componentsRaw = []byte("{}")
	}
	_, err = execCtx.ExecContext(ctx, `
INSERT INTO user_signup_fingerprints (
  user_id, composite_hash, canvas_hash, webgl_hash, components, duplicate_count, risk_flagged, risk_reason, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
ON CONFLICT (user_id) DO UPDATE SET
  composite_hash = EXCLUDED.composite_hash,
  canvas_hash = EXCLUDED.canvas_hash,
  webgl_hash = EXCLUDED.webgl_hash,
  components = EXCLUDED.components,
  duplicate_count = EXCLUDED.duplicate_count,
  risk_flagged = EXCLUDED.risk_flagged,
  risk_reason = EXCLUDED.risk_reason,
  updated_at = CURRENT_TIMESTAMP`,
		userID, composite, canvas, webgl, string(componentsRaw), duplicateCount, riskFlagged, riskReason)
	return err
}

func (r *affiliateIdentityRepository) GetInviterIDForInvitee(ctx context.Context, inviteeUserID int64) (*int64, error) {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := execCtx.QueryContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id = $1`, inviteeUserID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, rows.Err()
	}
	var inviter sql.NullInt64
	if err := rows.Scan(&inviter); err != nil {
		return nil, err
	}
	if !inviter.Valid {
		return nil, nil
	}
	return &inviter.Int64, nil
}

func (r *affiliateIdentityRepository) ListIdentityCandidates(ctx context.Context, inviterID int64, cfg *service.AffiliateIdentityConfig) ([]service.AffiliateIdentityCandidate, error) {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return nil, err
	}
	orderTypes := map[string]bool{}
	if cfg != nil {
		for _, t := range cfg.EligibleOrderTypes {
			orderTypes[t] = true
		}
	}
	includeBalance := orderTypes[payment.OrderTypeBalance]
	includeSubscription := orderTypes[payment.OrderTypeSubscription]
	rows, err := execCtx.QueryContext(ctx, `
SELECT ua.user_id,
       COALESCE(SUM(CASE
         WHEN po.status = $2 THEN GREATEST(po.pay_amount - COALESCE(po.refund_amount, 0), 0)
         WHEN po.status = $3 THEN po.pay_amount
         ELSE 0
       END), 0)::double precision AS paid_amount,
       COALESCE(BOOL_OR(COALESCE(f.risk_flagged, false)), false) AS risk_flagged
FROM user_affiliates ua
LEFT JOIN user_signup_fingerprints f ON f.user_id = ua.user_id
LEFT JOIN payment_orders po ON po.user_id = ua.user_id
  AND (($4 AND po.order_type = $5) OR ($6 AND po.order_type = $7))
  AND po.status IN ($2, $3)
WHERE ua.inviter_id = $1
GROUP BY ua.user_id
ORDER BY ua.user_id ASC`,
		inviterID,
		service.OrderStatusPartiallyRefunded,
		service.OrderStatusCompleted,
		includeBalance,
		payment.OrderTypeBalance,
		includeSubscription,
		payment.OrderTypeSubscription,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []service.AffiliateIdentityCandidate{}
	for rows.Next() {
		var item service.AffiliateIdentityCandidate
		if err := rows.Scan(&item.UserID, &item.PaidAmount, &item.RiskFlagged); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *affiliateIdentityRepository) UpsertIdentity(ctx context.Context, userID int64, identityType string, rateMultiplier float64, sourceInviterID *int64, expiresAt time.Time, snapshot map[string]any) error {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return err
	}
	snapshotRaw, _ := json.Marshal(snapshot)
	if string(snapshotRaw) == "null" {
		snapshotRaw = []byte("{}")
	}
	res, err := execCtx.ExecContext(ctx, `
UPDATE user_affiliate_identities
SET rate_multiplier = $1, source_inviter_id = $2, granted_at = CURRENT_TIMESTAMP, expires_at = $3,
    status = $4, qualification_snapshot = $5, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $6 AND identity_type = $7`,
		rateMultiplier, nullableInt64Arg(sourceInviterID), expiresAt, "active", string(snapshotRaw), userID, identityType)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected > 0 {
		return nil
	}
	_, err = execCtx.ExecContext(ctx, `
INSERT INTO user_affiliate_identities (
  user_id, identity_type, rate_multiplier, source_inviter_id, granted_at, expires_at,
  status, qualification_snapshot, created_at, updated_at
) VALUES ($1,$2,$3,$4,CURRENT_TIMESTAMP,$5,$6,$7,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		userID, identityType, rateMultiplier, nullableInt64Arg(sourceInviterID), expiresAt, "active", string(snapshotRaw))
	return err
}

func (r *affiliateIdentityRepository) RevokeSystemIdentitiesForInviter(ctx context.Context, inviterID int64) error {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return err
	}
	_, err = execCtx.ExecContext(ctx, `
UPDATE user_affiliate_identities
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE status = $2
  AND (user_id = $3 OR source_inviter_id = $3)`,
		"revoked", "active", inviterID)
	return err
}

func (r *affiliateIdentityRepository) RevokeStaleInviteeIdentities(ctx context.Context, inviterID int64, paidInvitees []service.AffiliateIdentityCandidate) error {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return err
	}
	keep := make(map[int64]struct{}, len(paidInvitees))
	for _, invitee := range paidInvitees {
		keep[invitee.UserID] = struct{}{}
	}
	rows, err := execCtx.QueryContext(ctx, `SELECT user_id FROM user_affiliate_identities WHERE identity_type = $1 AND source_inviter_id = $2 AND status = $3`, service.AffiliateIdentityTypeInvitee, inviterID, "active")
	if err != nil {
		return err
	}
	stale := []int64{}
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return err
		}
		if _, ok := keep[userID]; !ok {
			stale = append(stale, userID)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, userID := range stale {
		if _, err := execCtx.ExecContext(ctx, `UPDATE user_affiliate_identities SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2 AND identity_type = $3 AND source_inviter_id = $4`, "revoked", userID, service.AffiliateIdentityTypeInvitee, inviterID); err != nil {
			return err
		}
	}
	return nil
}

func (r *affiliateIdentityRepository) GetActiveIdentity(ctx context.Context, userID int64) (*service.AffiliateIdentityState, error) {
	execCtx, err := r.execer(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := execCtx.QueryContext(ctx, `
SELECT user_id, identity_type, rate_multiplier, source_inviter_id, granted_at, expires_at, status
FROM user_affiliate_identities
WHERE user_id = $1 AND status = $2 AND expires_at > CURRENT_TIMESTAMP
ORDER BY rate_multiplier ASC, expires_at DESC
LIMIT 1`, userID, "active")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, rows.Err()
	}
	var state service.AffiliateIdentityState
	var source sql.NullInt64
	if err := rows.Scan(&state.UserID, &state.Type, &state.RateMultiplier, &source, &state.GrantedAt, &state.ExpiresAt, &state.Status); err != nil {
		return nil, err
	}
	if source.Valid {
		state.SourceInviterID = &source.Int64
	}
	return &state, nil
}
