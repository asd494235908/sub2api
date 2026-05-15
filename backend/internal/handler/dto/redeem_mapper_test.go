package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestRedeemCodeFromServiceAdminIncludesPaymentSource(t *testing.T) {
	rc := &service.RedeemCode{
		ID:     12,
		Code:   "PAY-TEST",
		Type:   service.RedeemTypeBalance,
		Value:  180,
		Status: service.StatusUsed,
		Source: "payment:wxpay",
	}

	got := RedeemCodeFromServiceAdmin(rc)
	if got == nil {
		t.Fatal("RedeemCodeFromServiceAdmin returned nil")
	}
	if got.Source != "payment:wxpay" {
		t.Fatalf("Source = %q, want payment:wxpay", got.Source)
	}
}
