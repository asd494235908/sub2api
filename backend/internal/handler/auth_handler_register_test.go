//go:build unit

package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/stretchr/testify/require"
)

func TestRegisterRequest_UnmarshalAffiliateAlias(t *testing.T) {
	var req handler.RegisterRequest

	err := json.Unmarshal([]byte(`{"email":"invitee@example.com","password":"password","aff":"AFF_ALIAS"}`), &req)
	require.NoError(t, err)
	require.Equal(t, "AFF_ALIAS", req.AffCode)
}

func TestRegisterRequest_UnmarshalAffiliateCodeTakesPriorityOverAlias(t *testing.T) {
	var req handler.RegisterRequest

	err := json.Unmarshal([]byte(`{"email":"invitee@example.com","password":"password","aff":"AFF_ALIAS","aff_code":"AFF_EXPLICIT"}`), &req)
	require.NoError(t, err)
	require.Equal(t, "AFF_EXPLICIT", req.AffCode)
}
