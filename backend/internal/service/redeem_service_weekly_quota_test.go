//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type weeklyQuotaSettingsProviderStub struct {
	settings *SystemSettings
	err      error
}

type weeklyQuotaRedeemRepoStub struct {
	client *dbent.Client
}

func (s *weeklyQuotaSettingsProviderStub) GetAllSettings(context.Context) (*SystemSettings, error) {
	return s.settings, s.err
}

func weeklyQuotaDerefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *weeklyQuotaRedeemRepoStub) Create(ctx context.Context, code *RedeemCode) error {
	created, err := r.client.RedeemCode.Create().
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays).
		SetNillableUsedBy(code.UsedBy).
		SetNillableUsedAt(code.UsedAt).
		SetNillableGroupID(code.GroupID).
		Save(ctx)
	if err == nil {
		code.ID = created.ID
		code.CreatedAt = created.CreatedAt
	}
	return err
}

func (r *weeklyQuotaRedeemRepoStub) CreateBatch(context.Context, []RedeemCode) error {
	panic("unexpected CreateBatch")
}

func (r *weeklyQuotaRedeemRepoStub) GetByID(ctx context.Context, id int64) (*RedeemCode, error) {
	m, err := r.client.RedeemCode.Get(ctx, id)
	if err != nil {
		return nil, ErrRedeemCodeNotFound
	}
	return &RedeemCode{
		ID:           m.ID,
		Code:         m.Code,
		Type:         m.Type,
		Value:        m.Value,
		Status:       m.Status,
		UsedBy:       m.UsedBy,
		UsedAt:       m.UsedAt,
		Notes:        weeklyQuotaDerefString(m.Notes),
		CreatedAt:    m.CreatedAt,
		GroupID:      m.GroupID,
		ValidityDays: m.ValidityDays,
	}, nil
}

func (r *weeklyQuotaRedeemRepoStub) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
	m, err := r.client.RedeemCode.Query().Where(redeemcode.CodeEQ(code)).Only(ctx)
	if err != nil {
		return nil, ErrRedeemCodeNotFound
	}
	return &RedeemCode{
		ID:           m.ID,
		Code:         m.Code,
		Type:         m.Type,
		Value:        m.Value,
		Status:       m.Status,
		UsedBy:       m.UsedBy,
		UsedAt:       m.UsedAt,
		Notes:        weeklyQuotaDerefString(m.Notes),
		CreatedAt:    m.CreatedAt,
		GroupID:      m.GroupID,
		ValidityDays: m.ValidityDays,
	}, nil
}

func (r *weeklyQuotaRedeemRepoStub) Update(context.Context, *RedeemCode) error {
	panic("unexpected Update")
}

func (r *weeklyQuotaRedeemRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete")
}

func (r *weeklyQuotaRedeemRepoStub) Use(ctx context.Context, id, userID int64) error {
	_, err := r.client.RedeemCode.UpdateOneID(id).
		SetStatus(StatusUsed).
		SetUsedBy(userID).
		SetUsedAt(time.Now().UTC()).
		Save(ctx)
	return err
}

func (r *weeklyQuotaRedeemRepoStub) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected List")
}

func (r *weeklyQuotaRedeemRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters")
}

func (r *weeklyQuotaRedeemRepoStub) ListByUser(ctx context.Context, userID int64, limit int) ([]RedeemCode, error) {
	if limit <= 0 {
		limit = 10
	}
	models, err := r.client.RedeemCode.Query().
		Where(redeemcode.UsedByEQ(userID)).
		Order(dbent.Desc(redeemcode.FieldUsedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RedeemCode, 0, len(models))
	for _, m := range models {
		out = append(out, RedeemCode{
			ID:           m.ID,
			Code:         m.Code,
			Type:         m.Type,
			Value:        m.Value,
			Status:       m.Status,
			UsedBy:       m.UsedBy,
			UsedAt:       m.UsedAt,
			Notes:        weeklyQuotaDerefString(m.Notes),
			CreatedAt:    m.CreatedAt,
			GroupID:      m.GroupID,
			ValidityDays: m.ValidityDays,
		})
	}
	return out, nil
}

func (r *weeklyQuotaRedeemRepoStub) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserPaginated")
}

func (r *weeklyQuotaRedeemRepoStub) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	panic("unexpected SumPositiveBalanceByUser")
}

func newWeeklyQuotaTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:redeem_service_weekly_quota?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

func TestWeeklyQuotaWindowForUserUsesRollingSevenDayWindows(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 8, 30, 0, 0, time.UTC)
	now := createdAt.Add(15*24*time.Hour + 2*time.Hour)

	windowStart, windowEnd := weeklyQuotaWindowForUser(createdAt, now)
	require.Equal(t, createdAt.Add(14*24*time.Hour), windowStart)
	require.Equal(t, createdAt.Add(21*24*time.Hour), windowEnd)
}

func TestRedeemService_GetWeeklyQuotaInfo_Disabled(t *testing.T) {
	client := newWeeklyQuotaTestClient(t)
	userRepo := &userRepoStub{
		user: &User{
			ID:        1,
			Email:     "disabled@example.com",
			CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	svc := NewRedeemService(
		nil,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		&weeklyQuotaSettingsProviderStub{settings: &SystemSettings{WeeklyQuotaEnabled: false, WeeklyQuotaAmount: 10}},
	)

	info, err := svc.GetWeeklyQuotaInfo(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, info.Enabled)
	require.Equal(t, WeeklyQuotaStatusDisabled, info.Status)
	require.Equal(t, 10.0, info.Amount)
}

func TestRedeemService_ClaimWeeklyQuota_SucceedsAndWritesRedeemRecord(t *testing.T) {
	ctx := context.Background()
	client := newWeeklyQuotaTestClient(t)
	createdAt := time.Now().UTC().Add(-2 * 24 * time.Hour)
	userEnt, err := client.User.Create().
		SetEmail("weekly@example.com").
		SetPasswordHash("hash").
		SetUsername("weekly-user").
		SetBalance(5).
		SetCreatedAt(createdAt).
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:        userEnt.ID,
			Email:     userEnt.Email,
			Username:  userEnt.Username,
			Balance:   5,
			CreatedAt: createdAt,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, userEnt.ID, id)
		userRepo.getByIDUser.Balance += amount
		return nil
	}

	redeemRepo := &weeklyQuotaRedeemRepoStub{client: client}
	svc := NewRedeemService(
		redeemRepo,
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		&weeklyQuotaSettingsProviderStub{settings: &SystemSettings{WeeklyQuotaEnabled: true, WeeklyQuotaAmount: 12.5}},
	)

	result, err := svc.ClaimWeeklyQuota(ctx, userEnt.ID)
	require.NoError(t, err)
	require.Equal(t, RedeemTypeWeeklyBalance, result.Type)
	require.Equal(t, 12.5, result.Value)
	require.Equal(t, 17.5, result.NewBalance)

	info, err := svc.GetWeeklyQuotaInfo(ctx, userEnt.ID)
	require.NoError(t, err)
	require.True(t, info.Enabled)
	require.Equal(t, WeeklyQuotaStatusClaimed, info.Status)
	require.NotNil(t, info.ClaimedAt)
	require.Equal(t, int64(1), info.TotalClaimCount)
	require.Equal(t, 12.5, info.TotalClaimAmount)
}

func TestRedeemService_ClaimWeeklyQuota_RejectsSecondClaimInSameWindow(t *testing.T) {
	ctx := context.Background()
	client := newWeeklyQuotaTestClient(t)
	createdAt := time.Now().UTC().Add(-24 * time.Hour)
	userEnt, err := client.User.Create().
		SetEmail("claimed@example.com").
		SetPasswordHash("hash").
		SetUsername("claimed-user").
		SetBalance(3).
		SetCreatedAt(createdAt).
		Save(ctx)
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:        userEnt.ID,
			Email:     userEnt.Email,
			Username:  userEnt.Username,
			Balance:   3,
			CreatedAt: createdAt,
		},
	}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		userRepo.getByIDUser.Balance += amount
		return nil
	}

	svc := NewRedeemService(
		&weeklyQuotaRedeemRepoStub{client: client},
		userRepo,
		nil,
		nil,
		nil,
		client,
		nil,
		&weeklyQuotaSettingsProviderStub{settings: &SystemSettings{WeeklyQuotaEnabled: true, WeeklyQuotaAmount: 9}},
	)

	_, err = svc.ClaimWeeklyQuota(ctx, userEnt.ID)
	require.NoError(t, err)

	_, err = svc.ClaimWeeklyQuota(ctx, userEnt.ID)
	require.ErrorIs(t, err, ErrWeeklyQuotaClaimed)
}
