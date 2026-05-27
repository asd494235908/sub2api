package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"

	entsql "entgo.io/ent/dialect/sql"
)

func TestRedeemCodeListOrder_UsedAtDescUsesNullsLast(t *testing.T) {
	selector := entsql.Dialect("postgres").Select("*").From(entsql.Table("redeem_codes"))

	for _, order := range redeemCodeListOrder(pagination.PaginationParams{
		SortBy:    "used_at",
		SortOrder: pagination.SortOrderDesc,
	}) {
		order(selector)
	}

	query, _ := selector.Query()
	if !strings.Contains(query, redeemcode.FieldUsedAt) || !strings.Contains(query, "DESC NULLS LAST") {
		t.Fatalf("used_at desc order query = %q, want NULLS LAST", query)
	}
}
