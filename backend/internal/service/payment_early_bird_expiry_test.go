package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestEarlyBirdSubscriptionPlanExpiresAfterPaidFulfillment(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("asd494235908@qq.com").
		SetPasswordHash("hash").
		SetUsername("early-bird-user").
		Save(ctx)
	require.NoError(t, err)

	groupID := int64(36501)
	totalLimit := 1560.0
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetName("早鸟套餐").
		SetDescription("early bird annual subscription").
		SetPrice(99).
		SetValidityDays(365).
		SetValidityUnit("days").
		SetFeatures("").
		SetProductName("").
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)

	groupRepo := &paymentEarlyBirdGroupRepo{
		group: &Group{ID: groupID, Status: payment.EntityStatusActive, SubscriptionType: SubscriptionTypeSubscription, SubscriptionTotalLimitUSD: &totalLimit},
	}
	svc := &PaymentService{entClient: client, groupRepo: groupRepo}
	order, err := svc.createOrderInTx(
		ctx,
		CreateOrderRequest{
			UserID:      user.ID,
			PaymentType: payment.TypeAlipay,
			OrderType:   payment.OrderTypeSubscription,
			ClientIP:    "127.0.0.1",
			SrcHost:     "app.example.com",
		},
		&User{ID: user.ID, Email: user.Email, Username: user.Username},
		plan,
		&PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30},
		99,
		99,
		0,
		99,
		&payment.InstanceSelection{ProviderKey: payment.TypeAlipay},
	)
	require.NoError(t, err)
	require.Equal(t, 365, paymentEarlyBirdIntValueOrZero(order.SubscriptionDays))
	require.NotNil(t, order.SubscriptionTotalLimitUsd)
	require.Equal(t, totalLimit, *order.SubscriptionTotalLimitUsd)

	subRepo := newPaymentEarlyBirdUserSubRepo()
	_, err = subRepo.GetByUserIDAndGroupID(ctx, user.ID, groupID)
	require.ErrorIs(t, err, ErrSubscriptionNotFound)

	svc.groupRepo = groupRepo
	svc.subscriptionSvc = NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)
	beforeFulfillment := time.Now()
	_, err = client.PaymentOrder.UpdateOneID(order.ID).
		SetStatus(OrderStatusPaid).
		SetPaymentTradeNo("paid-early-bird-1").
		SetPaidAt(beforeFulfillment).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.ExecuteSubscriptionFulfillment(ctx, order.ID))
	createdSub, err := subRepo.GetByUserIDAndGroupID(ctx, user.ID, groupID)
	require.NoError(t, err)
	require.WithinDuration(t, beforeFulfillment.AddDate(0, 0, 365), createdSub.ExpiresAt, 3*time.Second)
	require.NotNil(t, createdSub.TotalLimitUSD)
	require.Equal(t, totalLimit, *createdSub.TotalLimitUSD)
	require.Equal(t, 0.0, createdSub.TotalUsageUSD)

	secondOrder, err := svc.createOrderInTx(
		ctx,
		CreateOrderRequest{
			UserID:      user.ID,
			PaymentType: payment.TypeAlipay,
			OrderType:   payment.OrderTypeSubscription,
			ClientIP:    "127.0.0.1",
			SrcHost:     "app.example.com",
		},
		&User{ID: user.ID, Email: user.Email, Username: user.Username},
		plan,
		&PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30},
		99,
		99,
		0,
		99,
		&payment.InstanceSelection{ProviderKey: payment.TypeAlipay},
	)
	require.NoError(t, err)
	require.Equal(t, 365, paymentEarlyBirdIntValueOrZero(secondOrder.SubscriptionDays))
	require.NotNil(t, secondOrder.SubscriptionTotalLimitUsd)
	require.Equal(t, totalLimit, *secondOrder.SubscriptionTotalLimitUsd)

	_, err = client.PaymentOrder.UpdateOneID(secondOrder.ID).
		SetStatus(OrderStatusPaid).
		SetPaymentTradeNo("paid-early-bird-2").
		SetPaidAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.ExecuteSubscriptionFulfillment(ctx, secondOrder.ID))
	renewedSub, err := subRepo.GetByUserIDAndGroupID(ctx, user.ID, groupID)
	require.NoError(t, err)
	require.WithinDuration(t, createdSub.ExpiresAt.AddDate(0, 0, 365), renewedSub.ExpiresAt, time.Second)
	require.NotNil(t, renewedSub.TotalLimitUSD)
	require.Equal(t, totalLimit*2, *renewedSub.TotalLimitUSD)
}

type paymentEarlyBirdGroupRepo struct {
	groupRepoNoop
	group *Group
}

func (r *paymentEarlyBirdGroupRepo) GetByID(_ context.Context, id int64) (*Group, error) {
	if r.group == nil || r.group.ID != id {
		return nil, ErrGroupNotFound
	}
	group := *r.group
	return &group, nil
}

type paymentEarlyBirdUserSubRepo struct {
	userSubRepoNoop
	nextID int64
	subs   map[string]*UserSubscription
}

func newPaymentEarlyBirdUserSubRepo() *paymentEarlyBirdUserSubRepo {
	return &paymentEarlyBirdUserSubRepo{
		nextID: 1,
		subs:   map[string]*UserSubscription{},
	}
}

func (r *paymentEarlyBirdUserSubRepo) key(userID, groupID int64) string {
	return strconvFormatInt(userID) + ":" + strconvFormatInt(groupID)
}

func (r *paymentEarlyBirdUserSubRepo) GetByUserIDAndGroupID(_ context.Context, userID, groupID int64) (*UserSubscription, error) {
	sub := r.subs[r.key(userID, groupID)]
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	copy := *sub
	return &copy, nil
}

func (r *paymentEarlyBirdUserSubRepo) Create(_ context.Context, sub *UserSubscription) error {
	if sub == nil {
		return nil
	}
	copy := *sub
	if copy.ID == 0 {
		copy.ID = r.nextID
		r.nextID++
	}
	sub.ID = copy.ID
	r.subs[r.key(copy.UserID, copy.GroupID)] = &copy
	return nil
}

func (r *paymentEarlyBirdUserSubRepo) GetByID(_ context.Context, id int64) (*UserSubscription, error) {
	for _, sub := range r.subs {
		if sub.ID != id {
			continue
		}
		copy := *sub
		return &copy, nil
	}
	return nil, ErrSubscriptionNotFound
}

func (r *paymentEarlyBirdUserSubRepo) ExtendExpiry(_ context.Context, subscriptionID int64, newExpiresAt time.Time) error {
	for key, sub := range r.subs {
		if sub.ID != subscriptionID {
			continue
		}
		copy := *sub
		copy.ExpiresAt = newExpiresAt
		copy.UpdatedAt = time.Now()
		r.subs[key] = &copy
		return nil
	}
	return ErrSubscriptionNotFound
}

func (r *paymentEarlyBirdUserSubRepo) AddTotalLimit(_ context.Context, subscriptionID int64, amount float64) error {
	for key, sub := range r.subs {
		if sub.ID != subscriptionID {
			continue
		}
		copy := *sub
		if amount > 0 {
			if copy.TotalLimitUSD == nil {
				copy.TotalLimitUSD = &amount
			} else {
				total := *copy.TotalLimitUSD + amount
				copy.TotalLimitUSD = &total
			}
		}
		copy.UpdatedAt = time.Now()
		r.subs[key] = &copy
		return nil
	}
	return ErrSubscriptionNotFound
}

func (r *paymentEarlyBirdUserSubRepo) UpdateNotes(_ context.Context, subscriptionID int64, notes string) error {
	for key, sub := range r.subs {
		if sub.ID != subscriptionID {
			continue
		}
		copy := *sub
		copy.Notes = notes
		copy.UpdatedAt = time.Now()
		r.subs[key] = &copy
		return nil
	}
	return ErrSubscriptionNotFound
}

func paymentEarlyBirdIntValueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
