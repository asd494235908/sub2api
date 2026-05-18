package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type affiliateReaderStub struct {
	inviters      []service.AffiliateInviterEntry
	invitersTotal int64
	invitersErr   error
	invitees      []service.AffiliateInvitee
	inviteesErr   error
	lastFilter    service.AffiliateAdminFilter
	lastInviterID int64
}

func (s *affiliateReaderStub) ListInviters(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	search := c.Query("search")
	s.lastFilter = service.AffiliateAdminFilter{
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}
	if s.invitersErr != nil {
		response.ErrorFrom(c, s.invitersErr)
		return
	}
	response.Paginated(c, s.inviters, s.invitersTotal, page, pageSize)
}

func (s *affiliateReaderStub) ListInviterInvitees(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		response.BadRequest(c, "Invalid user_id")
		return
	}
	s.lastInviterID = userID
	if s.inviteesErr != nil {
		response.ErrorFrom(c, s.inviteesErr)
		return
	}
	response.Success(c, s.invitees)
}

func newAffiliateTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, rec
}

func TestAffiliateReaderStub_ListInviters(t *testing.T) {
	t.Parallel()

	stub := &affiliateReaderStub{
		inviters: []service.AffiliateInviterEntry{
			{
				UserID:      7,
				Email:       "owner@example.com",
				Username:    "owner",
				AffCode:     "AFFOWNER",
				AffCount:    2,
				TotalRebate: 9.9,
			},
		},
		invitersTotal: 1,
	}

	ctx, rec := newAffiliateTestContext(http.MethodGet, "/api/v1/admin/affiliates/inviters?page=2&page_size=15&search=owner")
	stub.ListInviters(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 2, stub.lastFilter.Page)
	require.Equal(t, 15, stub.lastFilter.PageSize)
	require.Equal(t, "owner", stub.lastFilter.Search)

	var resp struct {
		Data struct {
			Items []service.AffiliateInviterEntry `json:"items"`
			Total int64                           `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, int64(7), resp.Data.Items[0].UserID)
	require.Equal(t, int64(1), resp.Data.Total)
}

func TestAffiliateReaderStub_ListInviterInvitees(t *testing.T) {
	t.Parallel()

	joinedAt := time.Now().UTC().Truncate(time.Second)
	stub := &affiliateReaderStub{
		invitees: []service.AffiliateInvitee{
			{
				UserID:      17,
				Email:       "friend@example.com",
				Username:    "friend",
				CreatedAt:   &joinedAt,
				TotalRebate: 5.5,
			},
		},
	}

	ctx, rec := newAffiliateTestContext(http.MethodGet, "/api/v1/admin/affiliates/inviters/9/invitees")
	ctx.Params = gin.Params{{Key: "user_id", Value: "9"}}
	stub.ListInviterInvitees(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(9), stub.lastInviterID)

	var resp struct {
		Data []service.AffiliateInvitee `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, int64(17), resp.Data[0].UserID)
}

func TestAffiliateReaderStub_ListInviterInviteesRejectsBadUserID(t *testing.T) {
	t.Parallel()

	stub := &affiliateReaderStub{}
	ctx, rec := newAffiliateTestContext(http.MethodGet, "/api/v1/admin/affiliates/inviters/abc/invitees")
	ctx.Params = gin.Params{{Key: "user_id", Value: "abc"}}
	stub.ListInviterInvitees(ctx)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAffiliateHandler_CreateInviteRelationRejectsSameUser(t *testing.T) {
	t.Parallel()

	svc := &service.AffiliateService{}
	handler := NewAffiliateHandler(svc, nil)
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/affiliates/invites", strings.NewReader(`{"inviter_user_id":11,"invitee_user_id":11,"overwrite":true}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateInviteRelation(ctx)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

var _ = context.Background
