package handler

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

const casdoorIHuyiSMSSecretEnv = "CASDOOR_IHUYI_SMS_SECRET"

type casdoorIHuyiSMSRequest struct {
	Mobile      string `form:"mobile" json:"mobile"`
	PhoneNumber string `form:"phoneNumber" json:"phoneNumber"`
	Content     string `form:"content" json:"content"`
}

// SendCasdoorIHuyiSMS proxies Casdoor Custom HTTP SMS requests through the
// existing IHuyi SMS service so phone normalization and provider error parsing stay centralized.
func (h *AuthHandler) SendCasdoorIHuyiSMS(c *gin.Context) {
	expectedSecret := strings.TrimSpace(os.Getenv(casdoorIHuyiSMSSecretEnv))
	if expectedSecret == "" {
		response.Error(c, http.StatusServiceUnavailable, "casdoor sms secret is not configured")
		return
	}
	if !constantTimeEqual(c.GetHeader("X-Casdoor-SMS-Secret"), expectedSecret) {
		response.Unauthorized(c, "invalid sms secret")
		return
	}

	var req casdoorIHuyiSMSRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	mobile := strings.TrimSpace(req.Mobile)
	if mobile == "" {
		mobile = strings.TrimSpace(req.PhoneNumber)
	}
	content := strings.TrimSpace(req.Content)
	if mobile == "" || content == "" {
		response.BadRequest(c, "mobile and content are required")
		return
	}

	if err := h.smsService.SendTemplateMessage(c.Request.Context(), mobile, "", content); response.ErrorFrom(c, err) {
		return
	}

	response.Success(c, gin.H{"sent": true})
}

func constantTimeEqual(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	if actual == "" || len(actual) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
