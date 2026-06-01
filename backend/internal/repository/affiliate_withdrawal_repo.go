package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type affiliateWithdrawalRepository struct {
	client *dbent.Client
}

func NewAffiliateWithdrawalRepository(client *dbent.Client, _ *sql.DB) service.AffiliateWithdrawalRepository {
	return &affiliateWithdrawalRepository{client: client}
}

func (r *affiliateWithdrawalRepository) CreateWithdrawal(ctx context.Context, input service.AffiliateWithdrawalCreateInput) (*service.AffiliateWithdrawal, error) {
	var out *service.AffiliateWithdrawal
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, input.UserID); err != nil {
			return err
		}
		rows, err := txClient.QueryContext(txCtx, `
WITH locked AS (
    SELECT user_id, aff_quota::double precision AS available
    FROM user_affiliates
    WHERE user_id = $1
    FOR UPDATE
),
deducted AS (
    UPDATE user_affiliates ua
    SET aff_quota = ua.aff_quota - $2,
        updated_at = NOW()
    FROM locked l
    WHERE ua.user_id = l.user_id
      AND l.available + 0.000000001 >= $2
    RETURNING ua.user_id
),
created AS (
    INSERT INTO user_affiliate_withdrawals (
        user_id, amount, status, payout_method, payout_account_note, created_at, updated_at
    )
    SELECT user_id, $2, $3, $4, $5, NOW(), NOW()
    FROM deducted
    RETURNING id
)
SELECT id FROM created`,
			input.UserID,
			input.Amount,
			service.AffiliateWithdrawalStatusPendingReview,
			input.PayoutMethod,
			input.PayoutAccountNote,
		)
		if err != nil {
			return fmt.Errorf("create affiliate withdrawal: %w", err)
		}
		defer func() { _ = rows.Close() }()
		var id int64
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return service.ErrAffiliateWithdrawalInsufficientQuota
		}
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if err := insertAffiliateWithdrawalLedger(txCtx, txClient, input.UserID, service.AffiliateWithdrawActionRequest, input.Amount, &id, ""); err != nil {
			return err
		}
		w, err := queryAffiliateWithdrawalByID(txCtx, txClient, id)
		if err != nil {
			return err
		}
		out = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *affiliateWithdrawalRepository) ListUserWithdrawals(ctx context.Context, userID int64, filter service.AffiliateWithdrawalListFilter) ([]service.AffiliateWithdrawal, int64, error) {
	filter = normalizeWithdrawalFilterForRepo(filter)
	where := []string{"w.user_id = $1"}
	args := []any{userID}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where = append(where, fmt.Sprintf("w.status = $%d", len(args)))
	}
	return queryAffiliateWithdrawals(ctx, clientFromContext(ctx, r.client), strings.Join(where, " AND "), args, filter)
}

func (r *affiliateWithdrawalRepository) ListAdminWithdrawals(ctx context.Context, filter service.AffiliateWithdrawalListFilter) ([]service.AffiliateWithdrawal, int64, error) {
	filter = normalizeWithdrawalFilterForRepo(filter)
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where = append(where, fmt.Sprintf("w.status = $%d", len(args)))
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(u.email ILIKE $%d OR u.username ILIKE $%d OR w.payout_account_note ILIKE $%d OR w.payout_trade_no ILIKE $%d)", n, n, n, n))
	}
	return queryAffiliateWithdrawals(ctx, clientFromContext(ctx, r.client), strings.Join(where, " AND "), args, filter)
}

func (r *affiliateWithdrawalRepository) GetWithdrawal(ctx context.Context, id int64) (*service.AffiliateWithdrawal, error) {
	return queryAffiliateWithdrawalByID(ctx, clientFromContext(ctx, r.client), id)
}

func (r *affiliateWithdrawalRepository) ApproveWithdrawal(ctx context.Context, id, adminID int64, note string) (*service.AffiliateWithdrawal, error) {
	var out *service.AffiliateWithdrawal
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		res, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliate_withdrawals
SET status = $1, admin_note = $2, reviewed_by = $3, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $4 AND status = $5`,
			service.AffiliateWithdrawalStatusApproved,
			note,
			adminID,
			id,
			service.AffiliateWithdrawalStatusPendingReview,
		)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			if _, err := queryAffiliateWithdrawalByID(txCtx, txClient, id); err != nil {
				return err
			}
			return service.ErrAffiliateWithdrawalInvalidStatus
		}
		out, err = queryAffiliateWithdrawalByID(txCtx, txClient, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *affiliateWithdrawalRepository) RejectWithdrawal(ctx context.Context, id, adminID int64, reason string) (*service.AffiliateWithdrawal, error) {
	return r.refundAndClose(ctx, id, adminID, service.AffiliateWithdrawalStatusPendingReview, service.AffiliateWithdrawalStatusRejected, service.AffiliateWithdrawActionReject, reason)
}

func (r *affiliateWithdrawalRepository) MarkWithdrawalPaid(ctx context.Context, id, adminID int64, input service.AffiliateWithdrawalPaidInput) (*service.AffiliateWithdrawal, error) {
	var out *service.AffiliateWithdrawal
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		w, err := queryAffiliateWithdrawalByID(txCtx, txClient, id)
		if err != nil {
			return err
		}
		res, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliate_withdrawals
SET status = $1, payout_channel = $2, payout_trade_no = $3, admin_note = $4, paid_by = $5, paid_at = NOW(), updated_at = NOW()
WHERE id = $6 AND status = $7`,
			service.AffiliateWithdrawalStatusPaid,
			input.PayoutChannel,
			input.PayoutTradeNo,
			input.AdminNote,
			adminID,
			id,
			service.AffiliateWithdrawalStatusApproved,
		)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return service.ErrAffiliateWithdrawalInvalidStatus
		}
		if err := insertAffiliateWithdrawalLedger(txCtx, txClient, w.UserID, service.AffiliateWithdrawActionPaid, w.Amount, &id, input.PayoutTradeNo); err != nil {
			return err
		}
		out, err = queryAffiliateWithdrawalByID(txCtx, txClient, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *affiliateWithdrawalRepository) MarkWithdrawalFailed(ctx context.Context, id, adminID int64, reason string) (*service.AffiliateWithdrawal, error) {
	return r.refundAndClose(ctx, id, adminID, service.AffiliateWithdrawalStatusApproved, service.AffiliateWithdrawalStatusFailed, service.AffiliateWithdrawActionFail, reason)
}

func (r *affiliateWithdrawalRepository) refundAndClose(ctx context.Context, id, adminID int64, from, to, action, reason string) (*service.AffiliateWithdrawal, error) {
	var out *service.AffiliateWithdrawal
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		w, err := queryAffiliateWithdrawalByID(txCtx, txClient, id)
		if err != nil {
			return err
		}
		if w.Status != from {
			return service.ErrAffiliateWithdrawalInvalidStatus
		}
		if _, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_quota = aff_quota + $1, updated_at = NOW()
WHERE user_id = $2`, w.Amount, w.UserID); err != nil {
			return err
		}
		var updateSQL string
		if to == service.AffiliateWithdrawalStatusRejected {
			updateSQL = `UPDATE user_affiliate_withdrawals SET status = $1, reject_reason = $2, reviewed_by = $3, reviewed_at = NOW(), updated_at = NOW() WHERE id = $4 AND status = $5`
		} else {
			updateSQL = `UPDATE user_affiliate_withdrawals SET status = $1, failure_reason = $2, paid_by = $3, updated_at = NOW() WHERE id = $4 AND status = $5`
		}
		res, err := txClient.ExecContext(txCtx, updateSQL, to, reason, adminID, id, from)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return service.ErrAffiliateWithdrawalInvalidStatus
		}
		if err := insertAffiliateWithdrawalLedger(txCtx, txClient, w.UserID, action, w.Amount, &id, reason); err != nil {
			return err
		}
		out, err = queryAffiliateWithdrawalByID(txCtx, txClient, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *affiliateWithdrawalRepository) withTx(ctx context.Context, fn func(context.Context, *dbent.Client) error) error {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return fn(ctx, tx.Client())
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin affiliate withdrawal transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx, tx.Client()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit affiliate withdrawal transaction: %w", err)
	}
	return nil
}

func queryAffiliateWithdrawalByID(ctx context.Context, client affiliateQueryExecer, id int64) (*service.AffiliateWithdrawal, error) {
	rows, err := client.QueryContext(ctx, affiliateWithdrawalSelectSQL()+` WHERE w.id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrAffiliateWithdrawalNotFound
	}
	return scanAffiliateWithdrawal(rows)
}

func queryAffiliateWithdrawals(ctx context.Context, client affiliateQueryExecer, where string, args []any, filter service.AffiliateWithdrawalListFilter) ([]service.AffiliateWithdrawal, int64, error) {
	countSQL := `SELECT COUNT(*) FROM user_affiliate_withdrawals w JOIN users u ON u.id = w.user_id WHERE ` + where
	rows, err := client.QueryContext(ctx, countSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			_ = rows.Close()
			return nil, 0, err
		}
	}
	if err := rows.Close(); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.PageSize, (filter.Page-1)*filter.PageSize)
	limitPos := len(queryArgs) - 1
	offsetPos := len(queryArgs)
	querySQL := affiliateWithdrawalSelectSQL() + ` WHERE ` + where + fmt.Sprintf(` ORDER BY w.created_at DESC, w.id DESC LIMIT $%d OFFSET $%d`, limitPos, offsetPos)
	rows, err = client.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := []service.AffiliateWithdrawal{}
	for rows.Next() {
		item, err := scanAffiliateWithdrawal(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *item)
	}
	return out, total, rows.Err()
}

func affiliateWithdrawalSelectSQL() string {
	return `
SELECT w.id, w.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''),
       w.amount::double precision, w.status, w.payout_method, w.payout_account_note,
       w.admin_note, w.payout_channel, w.payout_trade_no, w.reject_reason, w.failure_reason,
       w.reviewed_by, w.reviewed_at, w.paid_by, w.paid_at, w.created_at, w.updated_at
FROM user_affiliate_withdrawals w
JOIN users u ON u.id = w.user_id`
}

func scanAffiliateWithdrawal(rows *sql.Rows) (*service.AffiliateWithdrawal, error) {
	var w service.AffiliateWithdrawal
	var reviewedBy, paidBy sql.NullInt64
	var reviewedAt, paidAt sql.NullTime
	if err := rows.Scan(
		&w.ID,
		&w.UserID,
		&w.UserEmail,
		&w.Username,
		&w.Amount,
		&w.Status,
		&w.PayoutMethod,
		&w.PayoutAccountNote,
		&w.AdminNote,
		&w.PayoutChannel,
		&w.PayoutTradeNo,
		&w.RejectReason,
		&w.FailureReason,
		&reviewedBy,
		&reviewedAt,
		&paidBy,
		&paidAt,
		&w.CreatedAt,
		&w.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		w.ReviewedBy = &reviewedBy.Int64
	}
	if reviewedAt.Valid {
		w.ReviewedAt = &reviewedAt.Time
	}
	if paidBy.Valid {
		w.PaidBy = &paidBy.Int64
	}
	if paidAt.Valid {
		w.PaidAt = &paidAt.Time
	}
	return &w, nil
}

func insertAffiliateWithdrawalLedger(ctx context.Context, client affiliateQueryExecer, userID int64, action string, amount float64, withdrawalID *int64, note string) error {
	_, err := client.ExecContext(ctx, `
INSERT INTO user_affiliate_ledger (
    user_id, action, amount, source_user_id, aff_quota_after, aff_frozen_quota_after,
    aff_history_quota_after, created_at, updated_at
)
SELECT ua.user_id, $2, $3, NULL, ua.aff_quota, ua.aff_frozen_quota, ua.aff_history_quota, NOW(), NOW()
FROM user_affiliates ua
WHERE ua.user_id = $1`,
		userID,
		action,
		amount,
	)
	if err != nil {
		return fmt.Errorf("insert affiliate withdrawal ledger: %w", err)
	}
	return nil
}

func normalizeWithdrawalFilterForRepo(filter service.AffiliateWithdrawalListFilter) service.AffiliateWithdrawalListFilter {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = strings.TrimSpace(filter.Status)
	return filter
}
