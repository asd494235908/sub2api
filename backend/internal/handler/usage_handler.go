package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UsageHandler handles usage-related requests
type UsageHandler struct {
	usageService   *service.UsageService
	apiKeyService  *service.APIKeyService
	settingService *service.SettingService
}

// NewUsageHandler creates a new UsageHandler
func NewUsageHandler(usageService *service.UsageService, apiKeyService *service.APIKeyService, extras ...*service.SettingService) *UsageHandler {
	var settingService *service.SettingService
	if len(extras) > 0 {
		settingService = extras[0]
	}
	return &UsageHandler{
		usageService:   usageService,
		apiKeyService:  apiKeyService,
		settingService: settingService,
	}
}

// List handles listing usage records with pagination
// GET /api/v1/usage
func (h *UsageHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	var apiKeyID int64
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}

		// [Security Fix] Verify API Key ownership to prevent horizontal privilege escalation
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's usage records")
			return
		}

		apiKeyID = id
	}

	// Parse additional filters
	model := c.Query("model")

	var requestType *int16
	var stream *bool
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := c.Query("stream"); streamStr != "" {
		val, err := strconv.ParseBool(streamStr)
		if err != nil {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
		stream = &val
	}

	var billingType *int8
	if billingTypeStr := c.Query("billing_type"); billingTypeStr != "" {
		val, err := strconv.ParseInt(billingTypeStr, 10, 8)
		if err != nil {
			response.BadRequest(c, "Invalid billing_type")
			return
		}
		bt := int8(val)
		billingType = &bt
	}

	// Parse date range
	var startTime, endTime *time.Time
	userTZ := c.Query("timezone") // Get user's timezone from request
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, err := timezone.ParseInUserLocation("2006-01-02", startDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		startTime = &t
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, err := timezone.ParseInUserLocation("2006-01-02", endDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		// Use half-open range [start, end), move to next calendar day start (DST-safe).
		t = t.AddDate(0, 0, 1)
		endTime = &t
	}

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	filters := usagestats.UsageLogFilters{
		UserID:      subject.UserID, // Always filter by current user for security
		APIKeyID:    apiKeyID,
		Model:       model,
		RequestType: requestType,
		Stream:      stream,
		BillingType: billingType,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	records, result, err := h.usageService.ListWithFilters(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UsageLog, 0, len(records))
	for i := range records {
		out = append(out, *dto.UsageLogFromService(&records[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// GetByID handles getting a single usage record
// GET /api/v1/usage/:id
func (h *UsageHandler) GetByID(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	usageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid usage ID")
		return
	}

	record, err := h.usageService.GetByID(c.Request.Context(), usageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 验证所有权
	if record.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this record")
		return
	}

	response.Success(c, dto.UsageLogFromService(record))
}

// Stats handles getting usage statistics
// GET /api/v1/usage/stats
func (h *UsageHandler) Stats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var apiKeyID int64
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}

		// [Security Fix] Verify API Key ownership to prevent horizontal privilege escalation
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "API key not found")
			return
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's statistics")
			return
		}

		apiKeyID = id
	}

	// 获取时间范围参数
	userTZ := c.Query("timezone") // Get user's timezone from request
	now := timezone.NowInUserLocation(userTZ)
	var startTime, endTime time.Time

	// 优先使用 start_date 和 end_date 参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		// 使用自定义日期范围
		var err error
		startTime, err = timezone.ParseInUserLocation("2006-01-02", startDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		endTime, err = timezone.ParseInUserLocation("2006-01-02", endDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		// 与 SQL 条件 created_at < end 对齐，使用次日 00:00 作为上边界（DST-safe）。
		endTime = endTime.AddDate(0, 0, 1)
	} else {
		// 使用 period 参数
		period := c.DefaultQuery("period", "today")
		switch period {
		case "today":
			startTime = timezone.StartOfDayInUserLocation(now, userTZ)
		case "week":
			startTime = now.AddDate(0, 0, -7)
		case "month":
			startTime = now.AddDate(0, -1, 0)
		default:
			startTime = timezone.StartOfDayInUserLocation(now, userTZ)
		}
		endTime = now
	}

	var stats *service.UsageStats
	var err error
	if apiKeyID > 0 {
		stats, err = h.usageService.GetStatsByAPIKey(c.Request.Context(), apiKeyID, startTime, endTime)
	} else {
		stats, err = h.usageService.GetStatsByUser(c.Request.Context(), subject.UserID, startTime, endTime)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

const (
	defaultUsageLeaderboardLimit = 20
	maxUsageLeaderboardLimit     = 100
)

type usageLeaderboardItemResponse struct {
	Rank        int64   `json:"rank"`
	UserID      int64   `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Tokens      int64   `json:"tokens"`
	Requests    int64   `json:"requests"`
	ActualCost  float64 `json:"actual_cost"`
}

func parseUsageLeaderboardLimit(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return defaultUsageLeaderboardLimit, true
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if limit <= 0 {
		return defaultUsageLeaderboardLimit, true
	}
	if limit > maxUsageLeaderboardLimit {
		return maxUsageLeaderboardLimit, true
	}
	return limit, true
}

func usageLeaderboardRange(period, userTZ string) (time.Time, time.Time, string, bool) {
	normalized := strings.TrimSpace(strings.ToLower(period))
	if normalized == "" {
		normalized = "today"
	}
	now := timezone.NowInUserLocation(userTZ)
	todayStart := timezone.StartOfDayInUserLocation(now, userTZ)
	switch normalized {
	case "today":
		return todayStart, todayStart.AddDate(0, 0, 1), normalized, true
	case "yesterday":
		start := todayStart.AddDate(0, 0, -1)
		return start, todayStart, normalized, true
	default:
		return time.Time{}, time.Time{}, "", false
	}
}

func maskLeaderboardDisplayName(username, email string, userID int64) string {
	if trimmed := strings.TrimSpace(username); trimmed != "" {
		return maskPlainIdentifier(trimmed)
	}
	if trimmed := strings.TrimSpace(email); trimmed != "" {
		return maskEmail(trimmed)
	}
	return fmt.Sprintf("User #%d", userID)
}

func maskPlainIdentifier(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return "***"
	}
	return string(runes[0]) + "***"
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return maskPlainIdentifier(email)
	}
	local := []rune(parts[0])
	return string(local[0]) + "***@" + parts[1]
}

// Leaderboard handles user-facing token consumption leaderboard.
// GET /api/v1/usage/leaderboard
func (h *UsageHandler) Leaderboard(c *gin.Context) {
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	limit, ok := parseUsageLeaderboardLimit(c.Query("limit"))
	if !ok {
		response.BadRequest(c, "Invalid limit")
		return
	}
	startTime, endTime, period, ok := usageLeaderboardRange(c.DefaultQuery("period", "today"), c.Query("timezone"))
	if !ok {
		response.BadRequest(c, "Invalid period, use today or yesterday")
		return
	}

	ignoredUserIDs := []int64{}
	if h.settingService != nil {
		settings, err := h.settingService.GetUsageLeaderboardSettings(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		ignoredUserIDs = settings.IgnoredUserIDs
	}

	leaderboard, err := h.usageService.GetUserTokenLeaderboard(c.Request.Context(), startTime, endTime, limit, ignoredUserIDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]usageLeaderboardItemResponse, 0, len(leaderboard.Ranking))
	for _, row := range leaderboard.Ranking {
		items = append(items, usageLeaderboardItemResponse{
			Rank:        row.Rank,
			UserID:      row.UserID,
			DisplayName: maskLeaderboardDisplayName(row.Username, row.Email, row.UserID),
			Tokens:      row.Tokens,
			Requests:    row.Requests,
			ActualCost:  row.ActualCost,
		})
	}

	response.Success(c, gin.H{
		"period":         period,
		"start_date":     startTime.Format("2006-01-02"),
		"end_date":       endTime.AddDate(0, 0, -1).Format("2006-01-02"),
		"ranking":        items,
		"total_tokens":   leaderboard.TotalTokens,
		"total_requests": leaderboard.TotalRequests,
	})
}

// parseUserTimeRange parses start_date, end_date query parameters for user dashboard
// Uses user's timezone if provided, otherwise falls back to server timezone
func parseUserTimeRange(c *gin.Context) (time.Time, time.Time) {
	userTZ := c.Query("timezone") // Get user's timezone from request
	now := timezone.NowInUserLocation(userTZ)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time

	if startDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ); err == nil {
			startTime = t
		} else {
			startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
		}
	} else {
		startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
	}

	if endDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ); err == nil {
			endTime = t.Add(24 * time.Hour) // Include the end date
		} else {
			endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
		}
	} else {
		endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	}

	return startTime, endTime
}

const (
	defaultAPIKeyDailyUsageDays = 30
	maxAPIKeyDailyUsageDays     = 90
)

func parseAPIKeyDailyUsageDays(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return defaultAPIKeyDailyUsageDays, true
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 || days > maxAPIKeyDailyUsageDays {
		return 0, false
	}
	return days, true
}

func apiKeyDailyUsageRange(days int, userTZ string) (time.Time, time.Time) {
	now := timezone.NowInUserLocation(userTZ)
	startTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -(days-1)), userTZ)
	endTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	return startTime, endTime
}

// DashboardStats handles getting user dashboard statistics
// GET /api/v1/usage/dashboard/stats
func (h *UsageHandler) DashboardStats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.usageService.GetUserDashboardStats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// DashboardTrend handles getting user usage trend data
// GET /api/v1/usage/dashboard/trend
func (h *UsageHandler) DashboardTrend(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	startTime, endTime := parseUserTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")

	trend, err := h.usageService.GetUserUsageTrendByUserID(c.Request.Context(), subject.UserID, startTime, endTime, granularity)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// DashboardModels handles getting user model usage statistics
// GET /api/v1/usage/dashboard/models
func (h *UsageHandler) DashboardModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	startTime, endTime := parseUserTimeRange(c)

	stats, err := h.usageService.GetUserModelStats(c.Request.Context(), subject.UserID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"models":     stats,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// BatchAPIKeysUsageRequest represents the request for batch API keys usage
type BatchAPIKeysUsageRequest struct {
	APIKeyIDs []int64 `json:"api_key_ids" binding:"required"`
}

// DashboardAPIKeysUsage handles getting usage stats for user's own API keys
// POST /api/v1/usage/dashboard/api-keys-usage
func (h *UsageHandler) DashboardAPIKeysUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req BatchAPIKeysUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.APIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	// Limit the number of API key IDs to prevent SQL parameter overflow
	if len(req.APIKeyIDs) > 100 {
		response.BadRequest(c, "Too many API key IDs (maximum 100 allowed)")
		return
	}

	validAPIKeyIDs, err := h.apiKeyService.VerifyOwnership(c.Request.Context(), subject.UserID, req.APIKeyIDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if len(validAPIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	stats, err := h.usageService.GetBatchAPIKeyUsageStats(c.Request.Context(), validAPIKeyIDs, time.Time{}, time.Time{})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

// GetMyAPIKeyDailyUsage handles getting daily usage details for the current user's API key.
// GET /api/v1/user/api-keys/:id/usage/daily?days=30
func (h *UsageHandler) GetMyAPIKeyDailyUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	apiKeyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	days, ok := parseAPIKeyDailyUsageDays(c.DefaultQuery("days", ""))
	if !ok {
		response.BadRequest(c, "Invalid days, allowed range is 1-90")
		return
	}

	if h.apiKeyService == nil {
		response.InternalError(c, "API key service is not configured")
		return
	}

	apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), apiKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if apiKey.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this API key's usage")
		return
	}

	userTZ := c.Query("timezone")
	startTime, endTime := apiKeyDailyUsageRange(days, userTZ)
	items, err := h.usageService.GetAPIKeyDailyUsage(c.Request.Context(), subject.UserID, apiKeyID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":      items,
		"days":       days,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.AddDate(0, 0, -1).Format("2006-01-02"),
	})
}
