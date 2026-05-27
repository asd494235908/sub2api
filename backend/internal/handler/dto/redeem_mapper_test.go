package dto

import (
	"testing"
	"time"

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

func TestRedeemCodeFromServiceAdminTreatsUsedMetadataAsUsed(t *testing.T) {
	usedBy := int64(42)
	usedAt := time.Now().UTC()
	rc := &service.RedeemCode{
		ID:     13,
		Code:   "LEGACY-USED",
		Type:   service.RedeemTypeBalance,
		Value:  50,
		Status: service.StatusUnused,
		UsedBy: &usedBy,
		UsedAt: &usedAt,
	}

	got := RedeemCodeFromServiceAdmin(rc)
	if got == nil {
		t.Fatal("RedeemCodeFromServiceAdmin returned nil")
	}
	if got.Status != service.StatusUsed {
		t.Fatalf("Status = %q, want %q", got.Status, service.StatusUsed)
	}
}
