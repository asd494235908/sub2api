//go:build unit

package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGatewayServiceRecordUsage_DroppedUsageLogSyncFallback(t *testing.T) {
	usageRepo := &openAIRecordUsageBestEffortLogRepoStub{
		bestEffortErr: MarkUsageLogCreateDropped(errors.New("usage log best-effort queue full")),
	}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	svc := newGatewayRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: "gateway_drop_usage_log",
			Usage: ClaudeUsage{
				InputTokens:  10,
				OutputTokens: 6,
			},
			Model:    "claude-sonnet-4",
			Duration: time.Second,
		},
		APIKey:  &APIKey{ID: 508},
		User:    &User{ID: 608},
		Account: &Account{ID: 708},
	})

	require.NoError(t, err)
	require.Equal(t, 1, usageRepo.bestEffortCalls)
	require.Equal(t, 1, usageRepo.createCalls)
}

func TestGatewayServiceRecordUsage_BalanceKeyUsesLowestMemberAndAffiliateMultiplier(t *testing.T) {
	groupID := int64(901)
	identityRepo := newAffiliateIdentityMemoryRepo()
	expiresAt := time.Now().Add(time.Hour)
	require.NoError(t, identityRepo.UpsertIdentity(context.Background(), 601, AffiliateIdentityTypeInvitee, 0.7, nil, expiresAt, nil))

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	svc := newGatewayRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
	svc.paymentService = &PaymentService{
		configService: NewPaymentConfigService(nil, &openAIRecordUsageSettingRepoStub{values: map[string]string{
			SettingKeyMemberLevelEnabled:       "true",
			SettingKeyMemberLevelConfig:        `{"levels":[{"id":"codex_level_08","name":"Codex 0.8","min_recharge_amount":100,"rate_multiplier":0.8,"enabled":true,"sort_order":0}]}`,
			SettingKeyAffiliateIdentityEnabled: "true",
			SettingKeyAffiliateIdentityConfig:  `{"inviter_rate_multiplier":0.7,"invitee_rate_multiplier":0.7,"duration_hours":720,"qualified_invitee_count":0,"qualified_pay_amount":50,"eligible_order_types":["balance","subscription"],"fingerprint_enforcement_enabled":false,"max_accounts_per_fingerprint_hash":10}`,
		}}, nil),
		affiliateService: &AffiliateService{
			settingService: &SettingService{settingRepo: &openAIRecordUsageSettingRepoStub{values: map[string]string{
				SettingKeyAffiliateIdentityEnabled: "true",
				SettingKeyAffiliateIdentityConfig:  `{"inviter_rate_multiplier":0.7,"invitee_rate_multiplier":0.7,"duration_hours":720,"qualified_invitee_count":0,"qualified_pay_amount":50,"eligible_order_types":["balance","subscription"],"fingerprint_enforcement_enabled":false,"max_accounts_per_fingerprint_hash":10}`,
			}}},
			identityRepo: identityRepo,
		},
	}

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: "gateway_member_affiliate_min",
			Usage: ClaudeUsage{
				InputTokens:  1000,
				OutputTokens: 1000,
			},
			Model:    "claude-3-5-haiku",
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:      501,
			GroupID: i64p(groupID),
			Group: &Group{
				ID:               groupID,
				RateMultiplier:   1.2,
				SubscriptionType: SubscriptionTypeStandard,
			},
			Quota: 100,
		},
		User:    &User{ID: 601, TotalRecharged: 200},
		Account: &Account{ID: 701},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, BillingTypeBalance, usageRepo.lastLog.BillingType)
	require.InDelta(t, 0.7, usageRepo.lastLog.RateMultiplier, 1e-12)
	require.InDelta(t, usageRepo.lastLog.TotalCost*0.7, usageRepo.lastLog.ActualCost, 1e-12)
	require.NotNil(t, billingRepo.lastCmd)
	require.InDelta(t, usageRepo.lastLog.ActualCost, billingRepo.lastCmd.BalanceCost, 1e-12)
}

func TestGatewayServiceRecordUsage_SubscriptionKeyIgnoresMemberAndAffiliateMultiplier(t *testing.T) {
	groupID := int64(902)
	subscriptionID := int64(77)
	identityRepo := newAffiliateIdentityMemoryRepo()
	expiresAt := time.Now().Add(time.Hour)
	require.NoError(t, identityRepo.UpsertIdentity(context.Background(), 602, AffiliateIdentityTypeInvitee, 0.7, nil, expiresAt, nil))

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	svc := newGatewayRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
	svc.paymentService = &PaymentService{
		configService: NewPaymentConfigService(nil, &openAIRecordUsageSettingRepoStub{values: map[string]string{
			SettingKeyMemberLevelEnabled:       "true",
			SettingKeyMemberLevelConfig:        `{"levels":[{"id":"codex_level_08","name":"Codex 0.8","min_recharge_amount":100,"rate_multiplier":0.8,"enabled":true,"sort_order":0}]}`,
			SettingKeyAffiliateIdentityEnabled: "true",
			SettingKeyAffiliateIdentityConfig:  `{"inviter_rate_multiplier":0.7,"invitee_rate_multiplier":0.7,"duration_hours":720,"qualified_invitee_count":0,"qualified_pay_amount":50,"eligible_order_types":["balance","subscription"],"fingerprint_enforcement_enabled":false,"max_accounts_per_fingerprint_hash":10}`,
		}}, nil),
		affiliateService: &AffiliateService{
			settingService: &SettingService{settingRepo: &openAIRecordUsageSettingRepoStub{values: map[string]string{
				SettingKeyAffiliateIdentityEnabled: "true",
				SettingKeyAffiliateIdentityConfig:  `{"inviter_rate_multiplier":0.7,"invitee_rate_multiplier":0.7,"duration_hours":720,"qualified_invitee_count":0,"qualified_pay_amount":50,"eligible_order_types":["balance","subscription"],"fingerprint_enforcement_enabled":false,"max_accounts_per_fingerprint_hash":10}`,
			}}},
			identityRepo: identityRepo,
		},
	}

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: "gateway_subscription_ignores_member_affiliate",
			Usage: ClaudeUsage{
				InputTokens:  1000,
				OutputTokens: 1000,
			},
			Model:    "claude-3-5-haiku",
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:      502,
			GroupID: i64p(groupID),
			Group: &Group{
				ID:               groupID,
				RateMultiplier:   1.2,
				SubscriptionType: SubscriptionTypeSubscription,
			},
			Quota: 100,
		},
		User:         &User{ID: 602, TotalRecharged: 200},
		Account:      &Account{ID: 702},
		Subscription: &UserSubscription{ID: subscriptionID, UserID: 602, GroupID: groupID},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, BillingTypeSubscription, usageRepo.lastLog.BillingType)
	require.InDelta(t, 1.2, usageRepo.lastLog.RateMultiplier, 1e-12)
	require.NotNil(t, billingRepo.lastCmd)
	require.Zero(t, billingRepo.lastCmd.BalanceCost)
	require.InDelta(t, usageRepo.lastLog.ActualCost, billingRepo.lastCmd.SubscriptionCost, 1e-12)
}
