package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type SMSBroadcastHandler struct {
	service *service.SMSBroadcastService
}

func NewSMSBroadcastHandler(service *service.SMSBroadcastService) *SMSBroadcastHandler {
	return &SMSBroadcastHandler{service: service}
}

type SMSBroadcastAudienceRequest struct {
	UserIDs    []int64          `json:"user_ids"`
	Status     string           `json:"status"`
	Role       string           `json:"role"`
	Search     string           `json:"search"`
	GroupName  string           `json:"group_name"`
	Attributes map[int64]string `json:"attributes"`
}

type SMSBroadcastCreateRequest struct {
	Title      string                              `json:"title" binding:"required"`
	TemplateID string                              `json:"template_id" binding:"required"`
	Audience   SMSBroadcastAudienceRequest         `json:"audience"`
	Vars       []SMSBroadcastTemplateVarRowRequest `json:"vars"`
}

type SMSBroadcastTemplateVarRowRequest struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

type SMSBroadcastPreviewRequest struct {
	Audience SMSBroadcastAudienceRequest `json:"audience"`
}

type SMSBroadcastRecipientResponse struct {
	UserID       int64   `json:"user_id"`
	PhoneNumber  string  `json:"phone_number"`
	RawPhone     string  `json:"raw_phone"`
	RenderedBody string  `json:"rendered_body"`
	Status       string  `json:"status,omitempty"`
	Error        *string `json:"error_message,omitempty"`
	SentAt       *string `json:"sent_at,omitempty"`
}

type SMSBroadcastPreviewResponse struct {
	Total  int64                           `json:"total"`
	Sample []SMSBroadcastRecipientResponse `json:"sample"`
}

type SMSBroadcastRecipientListResponse struct {
	Items    []SMSBroadcastRecipientResponse `json:"items"`
	Total    int64                           `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"page_size"`
	Pages    int                             `json:"pages"`
}

// POST /api/v1/admin/sms-broadcasts/preview
func (h *SMSBroadcastHandler) Preview(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	var req SMSBroadcastPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	total, sample, err := h.service.PreviewAudience(c.Request.Context(), smsAudienceFromRequest(req.Audience))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, SMSBroadcastPreviewResponse{Total: total, Sample: smsRecipientsToResponse(sample)})
}

// POST /api/v1/admin/sms-broadcasts
func (h *SMSBroadcastHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	var req SMSBroadcastCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	actorID := subject.UserID

	input := &service.SMSBroadcastCampaignInput{
		Title:      req.Title,
		TemplateID: req.TemplateID,
		Audience:   smsAudienceFromRequest(req.Audience),
		VarRows:    smsVarRowsFromRequest(req.Vars),
		ActorID:    &actorID,
	}
	campaign, err := h.service.CreateAndQueueCampaign(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, campaign)
}

// GET /api/v1/admin/sms-broadcasts
func (h *SMSBroadcastHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListCampaigns(c.Request.Context(), pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

// GET /api/v1/admin/sms-broadcasts/:id
func (h *SMSBroadcastHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid sms broadcast ID")
		return
	}
	campaign, err := h.service.GetCampaignByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, campaign)
}

// GET /api/v1/admin/sms-broadcasts/:id/recipients
func (h *SMSBroadcastHandler) ListRecipients(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid sms broadcast ID")
		return
	}
	page, pageSize := response.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))
	items, pageResult, err := h.service.ListRecipientsPaginated(c.Request.Context(), id, pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "id",
		SortOrder: "asc",
	}, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, SMSBroadcastRecipientListResponse{
		Items:    smsRecipientsToResponse(items),
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
		Pages:    pageResult.Pages,
	})
}

// POST /api/v1/admin/sms-broadcasts/:id/cancel
func (h *SMSBroadcastHandler) Cancel(c *gin.Context) {
	if h == nil || h.service == nil {
		response.ErrorFrom(c, service.ErrSMSBroadcastServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid sms broadcast ID")
		return
	}
	if err := h.service.CancelCampaign(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "SMS broadcast canceled"})
}

func smsAudienceFromRequest(req SMSBroadcastAudienceRequest) service.SMSBroadcastAudienceFilters {
	return service.SMSBroadcastAudienceFilters{
		UserIDs:    req.UserIDs,
		Status:     strings.TrimSpace(req.Status),
		Role:       strings.TrimSpace(req.Role),
		Search:     strings.TrimSpace(req.Search),
		GroupName:  strings.TrimSpace(req.GroupName),
		Attributes: req.Attributes,
	}
}

func smsVarRowsFromRequest(rows []SMSBroadcastTemplateVarRowRequest) []service.SMSBroadcastTemplateVarRow {
	out := make([]service.SMSBroadcastTemplateVarRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, service.SMSBroadcastTemplateVarRow{
			Key:    row.Key,
			Value:  row.Value,
			Source: service.SMSBroadcastTemplateVarSource(row.Source),
		})
	}
	return out
}

func smsRecipientsToResponse(items []service.SMSBroadcastRecipient) []SMSBroadcastRecipientResponse {
	out := make([]SMSBroadcastRecipientResponse, 0, len(items))
	for i := range items {
		var sentAt *string
		if items[i].SentAt != nil {
			v := items[i].SentAt.UTC().Format(time.RFC3339)
			sentAt = &v
		}
		out = append(out, SMSBroadcastRecipientResponse{
			UserID:       items[i].UserID,
			PhoneNumber:  items[i].PhoneNumber,
			RawPhone:     items[i].RawPhone,
			RenderedBody: items[i].RenderedBody,
			Status:       items[i].Status,
			Error:        items[i].ErrorMessage,
			SentAt:       sentAt,
		})
	}
	return out
}
