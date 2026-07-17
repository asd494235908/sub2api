package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestOpenAIGatewayServiceRecordUsage_MemberLevelDisabledKeepsGroupRateForBalanceKey(t *testing.T) {
	groupID := int64(111)
	groupRate := 1.5
	usage := OpenAIUsage{InputTokens: 15, OutputTokens: 4, CacheReadInputTokens: 3}

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, userRepo, subRepo, nil)
	svc.paymentService = &PaymentService{
		configService: NewPaymentConfigService(nil, &openAIRecordUsageSettingRepoStub{values: map[string]string{
			SettingKeyMemberLevelEnabled: "false",
		}}, nil),
	}

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: "resp_member_disabled_group_rate",
			Usage:     usage,
			Model:     "gpt-5.1",
			Duration:  time.Second,
		},
		APIKey: &APIKey{
			ID:      1003,
			GroupID: i64p(groupID),
			Group: &Group{
				ID:               groupID,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   groupRate,
			},
		},
		User:    &User{ID: 2003, TotalRecharged: 3000},
		Account: &Account{ID: 3003},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, groupRate, usageRepo.lastLog.RateMultiplier)

	expected := expectedOpenAICost(t, svc, "gpt-5.1", usage, groupRate)
	require.InDelta(t, expected.ActualCost, usageRepo.lastLog.ActualCost, 1e-12)
	require.InDelta(t, expected.ActualCost, userRepo.lastAmount, 1e-12)
	require.Equal(t, 1, userRepo.deductCalls)
}

func TestOpenAIGatewayServiceRecordUsage_NormalizesLongBillingRequestID(t *testing.T) {
	usageRepo := &openAIRecordUsageLogRepoStub{}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, userRepo, subRepo, nil)

	longRequestID := "client:" + strings.Repeat("openai-long-billing-request-id-", 4)
	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: "upstream-openai-ignored",
			Usage: OpenAIUsage{
				InputTokens:  8,
				OutputTokens: 4,
			},
			Model:    "gpt-5.1",
			Duration: time.Second,
		},
		APIKey:           &APIKey{ID: 10050},
		User:             &User{ID: 20050},
		Account:          &Account{ID: 30050},
		BillingRequestID: longRequestID,
	})

	require.NoError(t, err)
	require.NotNil(t, billingRepo.lastCmd)
	require.NotNil(t, usageRepo.lastLog)
	require.LessOrEqual(t, len(billingRepo.lastCmd.RequestID), maxUsageBillingRequestIDLen)
	require.Equal(t, billingRepo.lastCmd.RequestID, usageRepo.lastLog.RequestID)
	require.True(t, strings.HasPrefix(billingRepo.lastCmd.RequestID, "client-h:"))
	require.Equal(t, billingRepo.lastCmd.RequestID, resolveUsageBillingRequestID(context.Background(), longRequestID, ""))
}
