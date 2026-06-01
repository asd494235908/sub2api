//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type affiliateWithdrawalMemoryRepo struct {
	nextID      int64
	available   map[int64]float64
	withdrawals map[int64]*AffiliateWithdrawal
}

func newAffiliateWithdrawalMemoryRepo() *affiliateWithdrawalMemoryRepo {
	return &affiliateWithdrawalMemoryRepo{
		nextID:      1,
		available:   map[int64]float64{},
		withdrawals: map[int64]*AffiliateWithdrawal{},
	}
}

func (r *affiliateWithdrawalMemoryRepo) CreateWithdrawal(_ context.Context, input AffiliateWithdrawalCreateInput) (*AffiliateWithdrawal, error) {
	if r.available[input.UserID]+1e-9 < input.Amount {
		return nil, ErrAffiliateWithdrawalInsufficientQuota
	}
	r.available[input.UserID] -= input.Amount
	w := &AffiliateWithdrawal{
		ID:                r.nextID,
		UserID:            input.UserID,
		Amount:            input.Amount,
		Status:            AffiliateWithdrawalStatusPendingReview,
		PayoutMethod:      input.PayoutMethod,
		PayoutAccountNote: input.PayoutAccountNote,
	}
	r.nextID++
	r.withdrawals[w.ID] = w
	return cloneAffiliateWithdrawal(w), nil
}

func (r *affiliateWithdrawalMemoryRepo) ListUserWithdrawals(_ context.Context, userID int64, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error) {
	out := []AffiliateWithdrawal{}
	for _, w := range r.withdrawals {
		if w.UserID == userID {
			out = append(out, *cloneAffiliateWithdrawal(w))
		}
	}
	return out, int64(len(out)), nil
}

func (r *affiliateWithdrawalMemoryRepo) ListAdminWithdrawals(_ context.Context, filter AffiliateWithdrawalListFilter) ([]AffiliateWithdrawal, int64, error) {
	out := make([]AffiliateWithdrawal, 0, len(r.withdrawals))
	for _, w := range r.withdrawals {
		out = append(out, *cloneAffiliateWithdrawal(w))
	}
	return out, int64(len(out)), nil
}

func (r *affiliateWithdrawalMemoryRepo) GetWithdrawal(_ context.Context, id int64) (*AffiliateWithdrawal, error) {
	w := r.withdrawals[id]
	if w == nil {
		return nil, ErrAffiliateWithdrawalNotFound
	}
	return cloneAffiliateWithdrawal(w), nil
}

func (r *affiliateWithdrawalMemoryRepo) ApproveWithdrawal(_ context.Context, id, adminID int64, note string) (*AffiliateWithdrawal, error) {
	w := r.withdrawals[id]
	if w == nil {
		return nil, ErrAffiliateWithdrawalNotFound
	}
	if w.Status != AffiliateWithdrawalStatusPendingReview {
		return nil, ErrAffiliateWithdrawalInvalidStatus
	}
	w.Status = AffiliateWithdrawalStatusApproved
	w.AdminNote = note
	return cloneAffiliateWithdrawal(w), nil
}

func (r *affiliateWithdrawalMemoryRepo) RejectWithdrawal(_ context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error) {
	w := r.withdrawals[id]
	if w == nil {
		return nil, ErrAffiliateWithdrawalNotFound
	}
	if w.Status != AffiliateWithdrawalStatusPendingReview {
		return nil, ErrAffiliateWithdrawalInvalidStatus
	}
	w.Status = AffiliateWithdrawalStatusRejected
	w.RejectReason = reason
	r.available[w.UserID] += w.Amount
	return cloneAffiliateWithdrawal(w), nil
}

func (r *affiliateWithdrawalMemoryRepo) MarkWithdrawalPaid(_ context.Context, id, adminID int64, input AffiliateWithdrawalPaidInput) (*AffiliateWithdrawal, error) {
	w := r.withdrawals[id]
	if w == nil {
		return nil, ErrAffiliateWithdrawalNotFound
	}
	if w.Status != AffiliateWithdrawalStatusApproved {
		return nil, ErrAffiliateWithdrawalInvalidStatus
	}
	w.Status = AffiliateWithdrawalStatusPaid
	w.PayoutChannel = input.PayoutChannel
	w.PayoutTradeNo = input.PayoutTradeNo
	w.AdminNote = input.AdminNote
	return cloneAffiliateWithdrawal(w), nil
}

func (r *affiliateWithdrawalMemoryRepo) MarkWithdrawalFailed(_ context.Context, id, adminID int64, reason string) (*AffiliateWithdrawal, error) {
	w := r.withdrawals[id]
	if w == nil {
		return nil, ErrAffiliateWithdrawalNotFound
	}
	if w.Status != AffiliateWithdrawalStatusApproved {
		return nil, ErrAffiliateWithdrawalInvalidStatus
	}
	w.Status = AffiliateWithdrawalStatusFailed
	w.FailureReason = reason
	r.available[w.UserID] += w.Amount
	return cloneAffiliateWithdrawal(w), nil
}

func cloneAffiliateWithdrawal(w *AffiliateWithdrawal) *AffiliateWithdrawal {
	if w == nil {
		return nil
	}
	cp := *w
	return &cp
}

func TestAffiliateWithdrawalRequestDeductsQuota(t *testing.T) {
	t.Parallel()
	repo := newAffiliateWithdrawalMemoryRepo()
	repo.available[11] = 80
	svc := NewAffiliateWithdrawalService(repo, affiliateWithdrawalEnabledSettings())

	w, err := svc.CreateWithdrawal(context.Background(), 11, AffiliateWithdrawalCreateInput{
		Amount:            30,
		PayoutMethod:      "wechat",
		PayoutAccountNote: "微信昵称：测试",
	})

	require.NoError(t, err)
	require.Equal(t, AffiliateWithdrawalStatusPendingReview, w.Status)
	require.InDelta(t, 30, w.Amount, 1e-9)
	require.InDelta(t, 50, repo.available[11], 1e-9)
}

func TestAffiliateWithdrawalRejectRefundsQuotaOnlyOnce(t *testing.T) {
	t.Parallel()
	repo := newAffiliateWithdrawalMemoryRepo()
	repo.available[11] = 80
	svc := NewAffiliateWithdrawalService(repo, affiliateWithdrawalEnabledSettings())
	w, err := svc.CreateWithdrawal(context.Background(), 11, AffiliateWithdrawalCreateInput{
		Amount:            30,
		PayoutMethod:      "wechat",
		PayoutAccountNote: "微信昵称：测试",
	})
	require.NoError(t, err)

	rejected, err := svc.RejectWithdrawal(context.Background(), w.ID, 99, "信息不完整")
	require.NoError(t, err)
	require.Equal(t, AffiliateWithdrawalStatusRejected, rejected.Status)
	require.InDelta(t, 80, repo.available[11], 1e-9)

	_, err = svc.RejectWithdrawal(context.Background(), w.ID, 99, "重复驳回")
	require.ErrorIs(t, err, ErrAffiliateWithdrawalInvalidStatus)
	require.InDelta(t, 80, repo.available[11], 1e-9)
}

func TestAffiliateWithdrawalPaidCannotBeFailedOrRejected(t *testing.T) {
	t.Parallel()
	repo := newAffiliateWithdrawalMemoryRepo()
	repo.available[11] = 80
	svc := NewAffiliateWithdrawalService(repo, affiliateWithdrawalEnabledSettings())
	w, err := svc.CreateWithdrawal(context.Background(), 11, AffiliateWithdrawalCreateInput{
		Amount:            30,
		PayoutMethod:      "wechat",
		PayoutAccountNote: "微信昵称：测试",
	})
	require.NoError(t, err)
	_, err = svc.ApproveWithdrawal(context.Background(), w.ID, 99, "通过")
	require.NoError(t, err)
	paid, err := svc.MarkWithdrawalPaid(context.Background(), w.ID, 99, AffiliateWithdrawalPaidInput{
		PayoutChannel: "wechat_manual",
		PayoutTradeNo: "wx123",
	})
	require.NoError(t, err)
	require.Equal(t, AffiliateWithdrawalStatusPaid, paid.Status)

	_, err = svc.MarkWithdrawalFailed(context.Background(), w.ID, 99, "失败")
	require.ErrorIs(t, err, ErrAffiliateWithdrawalInvalidStatus)
	_, err = svc.RejectWithdrawal(context.Background(), w.ID, 99, "驳回")
	require.ErrorIs(t, err, ErrAffiliateWithdrawalInvalidStatus)
	require.InDelta(t, 50, repo.available[11], 1e-9)
}

func affiliateWithdrawalEnabledSettings() *SettingService {
	return NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyAffiliateWithdrawEnabled:           "true",
		SettingKeyAffiliateWithdrawMinAmount:         "1",
		SettingKeyAffiliateWithdrawMaxAmount:         "0",
		SettingKeyAffiliateWithdrawDailyRequestLimit: "0",
	}}, nil)
}
