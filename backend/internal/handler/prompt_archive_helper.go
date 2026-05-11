package handler

import (
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func promptArchiveClientRequestID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if v, ok := c.Request.Context().Value(ctxkey.ClientRequestID).(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func promptArchiveRequestID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if v, ok := c.Request.Context().Value(ctxkey.RequestID).(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func enrichPromptArchiveEnvelope(c *gin.Context, env *service.PromptArchiveEnvelope) *service.PromptArchiveEnvelope {
	if env == nil {
		return nil
	}
	env.ClientRequestID = promptArchiveClientRequestID(c)
	env.RequestID = firstNonEmptyArchiveString(env.RequestID, promptArchiveRequestID(c), env.ClientRequestID)
	if env.CreatedAt.IsZero() {
		env.CreatedAt = time.Now().UTC()
	}
	return env
}

func firstNonEmptyArchiveString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
