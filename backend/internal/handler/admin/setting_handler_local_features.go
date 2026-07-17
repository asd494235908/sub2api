package admin

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strings"
)

type leaderboardIgnoredUserResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

func (h *SettingHandler) GetLeaderboardSettings(c *gin.Context) {
	settings, err := h.settingService.GetUsageLeaderboardSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"ignored_user_ids": settings.IgnoredUserIDs,
		"ignored_users":    h.loadLeaderboardIgnoredUsers(c.Request.Context(), settings.IgnoredUserIDs),
	})
}

func (h *SettingHandler) UpdatePromptArchiveSettings(c *gin.Context) {
	var req UpdatePromptArchiveSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	view := &service.PromptArchiveSettingsView{
		Enabled:         req.Enabled,
		AllGroups:       req.AllGroups,
		GroupIDs:        req.GroupIDs,
		Bucket:          strings.TrimSpace(req.Bucket),
		UpdatedByUserID: subject.UserID,
	}
	if err := h.settingService.SetPromptArchiveSettings(c.Request.Context(), view); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if h.promptArchiveService != nil {
		_, _ = h.promptArchiveService.UpdateSettings(c.Request.Context(), &service.PromptArchiveSettings{
			Enabled:   req.Enabled,
			AllGroups: req.AllGroups,
			GroupIDs:  req.GroupIDs,
			Bucket:    strings.TrimSpace(req.Bucket),
		}, subject.UserID)
	}
	updated, err := h.settingService.GetPromptArchiveSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.PromptArchiveSettings{
		Enabled:   updated.Enabled,
		AllGroups: updated.AllGroups,
		GroupIDs:  updated.GroupIDs,
		Bucket:    updated.Bucket,
	})
}

func (h *SettingHandler) UpdateLeaderboardSettings(c *gin.Context) {
	var req service.UsageLeaderboardSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	settings, err := h.settingService.UpdateUsageLeaderboardSettings(c.Request.Context(), &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"ignored_user_ids": settings.IgnoredUserIDs,
		"ignored_users":    h.loadLeaderboardIgnoredUsers(c.Request.Context(), settings.IgnoredUserIDs),
	})
}

type UpdatePromptArchiveSettingsRequest struct {
	Enabled   bool    `json:"enabled"`
	AllGroups bool    `json:"all_groups"`
	GroupIDs  []int64 `json:"group_ids"`
	Bucket    string  `json:"bucket"`
}

func (h *SettingHandler) GetPromptArchiveSettings(c *gin.Context) {
	settings, err := h.settingService.GetPromptArchiveSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.PromptArchiveSettings{
		Enabled:   settings.Enabled,
		AllGroups: settings.AllGroups,
		GroupIDs:  settings.GroupIDs,
		Bucket:    settings.Bucket,
	})
}

func (h *SettingHandler) loadLeaderboardIgnoredUsers(ctx context.Context, userIDs []int64) []leaderboardIgnoredUserResponse {
	if h.adminService == nil || len(userIDs) == 0 {
		return []leaderboardIgnoredUserResponse{}
	}
	users := make([]leaderboardIgnoredUserResponse, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := h.adminService.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				continue
			}
			slog.Warn("leaderboard_ignored_user_load_failed", "user_id", userID, "error", err)
			continue
		}
		users = append(users, leaderboardIgnoredUserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		})
	}
	return users
}
