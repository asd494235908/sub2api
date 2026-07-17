package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterInternalRoutes registers private integration endpoints.
func RegisterInternalRoutes(r *gin.Engine, h *handler.Handlers) {
	r.POST("/api/internal/casdoor/ihuyi-sms", h.Auth.SendCasdoorIHuyiSMS)
}
