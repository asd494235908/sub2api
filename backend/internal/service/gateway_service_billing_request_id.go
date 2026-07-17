package service

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const maxUsageBillingRequestIDLen = 64

func normalizeUsageBillingRequestID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if len(requestID) <= maxUsageBillingRequestIDLen {
		return requestID
	}

	prefix := "req-h:"
	switch {
	case strings.HasPrefix(requestID, "local:"):
		prefix = "local-h:"
	case strings.HasPrefix(requestID, "client:"):
		prefix = "client-h:"
	case strings.HasPrefix(requestID, "generated:"):
		prefix = "gen-h:"
	}
	sum := sha256.Sum256([]byte(requestID))
	hash := fmt.Sprintf("%x", sum)
	return prefix + hash[:maxUsageBillingRequestIDLen-len(prefix)]
}
