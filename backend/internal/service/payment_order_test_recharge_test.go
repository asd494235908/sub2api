package service

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"testing"
)

func TestValidateOrderInputTestRecharge(t *testing.T) {
	t.Parallel()

	svc := &PaymentService{}
	baseCfg := &PaymentConfig{
		MinAmount: 1,
		MaxAmount: 100,
	}

	t.Run("rejects when disabled", func(t *testing.T) {
		t.Parallel()
		_, err := svc.validateOrderInput(context.Background(), CreateOrderRequest{
			Amount:       0.01,
			PaymentType:  payment.TypeWxpay,
			OrderType:    payment.OrderTypeBalance,
			TestRecharge: true,
		}, baseCfg)
		if err == nil {
			t.Fatal("expected disabled test recharge to fail")
		}
		if appErr := infraerrors.FromError(err); appErr.Reason != "TEST_RECHARGE_DISABLED" {
			t.Fatalf("reason = %q, want TEST_RECHARGE_DISABLED", appErr.Reason)
		}
	})

	t.Run("allows wxpay balance amount below global minimum when enabled", func(t *testing.T) {
		t.Parallel()
		cfg := *baseCfg
		cfg.TestRechargeEnabled = true
		_, err := svc.validateOrderInput(context.Background(), CreateOrderRequest{
			Amount:       0.01,
			PaymentType:  payment.TypeWxpay,
			OrderType:    payment.OrderTypeBalance,
			TestRecharge: true,
		}, &cfg)
		if err != nil {
			t.Fatalf("expected enabled test recharge to pass: %v", err)
		}
	})

	tests := []struct {
		name string
		req  CreateOrderRequest
	}{
		{
			name: "rejects non wxpay",
			req: CreateOrderRequest{
				Amount:       0.01,
				PaymentType:  payment.TypeAlipay,
				OrderType:    payment.OrderTypeBalance,
				TestRecharge: true,
			},
		},
		{
			name: "rejects non test amount",
			req: CreateOrderRequest{
				Amount:       0.02,
				PaymentType:  payment.TypeWxpay,
				OrderType:    payment.OrderTypeBalance,
				TestRecharge: true,
			},
		},
		{
			name: "rejects subscription order",
			req: CreateOrderRequest{
				Amount:       0.01,
				PaymentType:  payment.TypeWxpay,
				OrderType:    payment.OrderTypeSubscription,
				PlanID:       1,
				TestRecharge: true,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := *baseCfg
			cfg.TestRechargeEnabled = true
			_, err := svc.validateOrderInput(context.Background(), tt.req, &cfg)
			if err == nil {
				t.Fatal("expected invalid test recharge to fail")
			}
			if appErr := infraerrors.FromError(err); appErr.Reason != "INVALID_TEST_RECHARGE" {
				t.Fatalf("reason = %q, want INVALID_TEST_RECHARGE", appErr.Reason)
			}
		})
	}
}
