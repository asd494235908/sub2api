package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingKeyMemberLevelEnabled = "member_level_enabled"
	SettingKeyMemberLevelConfig  = "member_level_config"
)

type MemberLevelConfig struct {
	Levels []MemberLevel `json:"levels"`
}

type MemberLevel struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	MinRechargeAmount float64 `json:"min_recharge_amount"`
	RateMultiplier    float64 `json:"rate_multiplier"`
	Enabled           bool    `json:"enabled"`
	SortOrder         int     `json:"sort_order"`
}

type MemberLevelState struct {
	LevelID          string   `json:"level_id"`
	LevelName        string   `json:"level_name"`
	RateMultiplier   float64  `json:"rate_multiplier"`
	TotalRecharged   float64  `json:"total_recharged"`
	CurrentThreshold float64  `json:"current_threshold"`
	NextThreshold    *float64 `json:"next_threshold,omitempty"`
	Progress         float64  `json:"progress"`
	ProgressCurrent  float64  `json:"progress_current"`
	ProgressTarget   float64  `json:"progress_target"`
}

func (s *PaymentService) memberLevelEnabled(ctx context.Context) bool {
	if s == nil || s.configService == nil {
		return false
	}
	value, err := s.configService.settingRepo.GetValue(ctx, SettingKeyMemberLevelEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *PaymentService) GetMemberLevelConfig(ctx context.Context) (*MemberLevelConfig, bool, error) {
	enabled := s.memberLevelEnabled(ctx)
	if s == nil || s.configService == nil {
		return nil, enabled, fmt.Errorf("payment service is not initialized")
	}
	raw, err := s.configService.settingRepo.GetValue(ctx, SettingKeyMemberLevelConfig)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, enabled, fmt.Errorf("get member level config: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		cfg, normErr := normalizeMemberLevelConfig(nil)
		return cfg, enabled, normErr
	}
	var cfg MemberLevelConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, enabled, infraerrors.InternalServer("MEMBER_LEVEL_CONFIG_INVALID", "member level config is invalid")
	}
	normalized, normErr := normalizeMemberLevelConfig(&cfg)
	if normErr != nil {
		fallback, fallbackErr := normalizeMemberLevelConfig(nil)
		return fallback, enabled, fallbackErr
	}
	return normalized, enabled, nil
}

func (s *PaymentService) UpdateMemberLevelConfig(ctx context.Context, enabled bool, cfg *MemberLevelConfig) (*MemberLevelConfig, error) {
	if s == nil || s.configService == nil {
		return nil, fmt.Errorf("payment service is not initialized")
	}
	normalized, err := normalizeMemberLevelConfig(cfg)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal member level config: %w", err)
	}
	updates := map[string]string{
		SettingKeyMemberLevelEnabled: strconv.FormatBool(enabled),
		SettingKeyMemberLevelConfig:  string(raw),
	}
	if err := s.configService.settingRepo.SetMultiple(ctx, updates); err != nil {
		return nil, fmt.Errorf("save member level config: %w", err)
	}
	return normalized, nil
}

func (s *PaymentService) ResolveMemberLevel(ctx context.Context, user *User) (*MemberLevelState, error) {
	totalRecharged := 0.0
	if user != nil {
		totalRecharged = user.TotalRecharged
	}
	cfg, enabled, err := s.GetMemberLevelConfig(ctx)
	if err != nil {
		return nil, err
	}
	return resolveMemberLevelState(cfg, enabled, totalRecharged), nil
}

func (s *PaymentService) MemberMultiplierEnabledForKey(ctx context.Context, apiKey *APIKey, subscription *UserSubscription) bool {
	if s == nil || !s.isRechargeKeyBilling(apiKey, subscription) {
		return false
	}
	if s.memberLevelEnabled(ctx) {
		return true
	}
	return s.affiliateService != nil && s.affiliateService.affiliateIdentityEnabled(ctx)
}

func normalizeMemberLevelConfig(cfg *MemberLevelConfig) (*MemberLevelConfig, error) {
	if cfg == nil {
		return &MemberLevelConfig{Levels: []MemberLevel{}}, nil
	}
	out := &MemberLevelConfig{Levels: make([]MemberLevel, 0, len(cfg.Levels))}
	seen := map[string]struct{}{}
	for _, level := range cfg.Levels {
		level.ID = strings.TrimSpace(level.ID)
		level.Name = strings.TrimSpace(level.Name)
		if level.ID == "" {
			return nil, infraerrors.BadRequest("MEMBER_LEVEL_INVALID", "member level id is required")
		}
		if _, ok := seen[level.ID]; ok {
			return nil, infraerrors.BadRequest("MEMBER_LEVEL_INVALID", "member level id must be unique")
		}
		seen[level.ID] = struct{}{}
		if level.Name == "" {
			return nil, infraerrors.BadRequest("MEMBER_LEVEL_INVALID", "member level name is required")
		}
		if math.IsNaN(level.MinRechargeAmount) || math.IsInf(level.MinRechargeAmount, 0) || level.MinRechargeAmount < 0 {
			return nil, infraerrors.BadRequest("MEMBER_LEVEL_INVALID", "member level threshold must be >= 0")
		}
		if math.IsNaN(level.RateMultiplier) || math.IsInf(level.RateMultiplier, 0) || level.RateMultiplier < 0 {
			return nil, infraerrors.BadRequest("MEMBER_LEVEL_INVALID", "member level rate multiplier must be >= 0")
		}
		out.Levels = append(out.Levels, level)
	}
	sort.SliceStable(out.Levels, func(i, j int) bool {
		if out.Levels[i].SortOrder == out.Levels[j].SortOrder {
			return out.Levels[i].MinRechargeAmount < out.Levels[j].MinRechargeAmount
		}
		return out.Levels[i].SortOrder < out.Levels[j].SortOrder
	})
	return out, nil
}

func resolveMemberLevelState(cfg *MemberLevelConfig, enabled bool, totalRecharged float64) *MemberLevelState {
	if totalRecharged < 0 {
		totalRecharged = 0
	}
	defaultLevel := MemberLevel{
		ID:                "default",
		Name:              "普通会员",
		MinRechargeAmount: 0,
		RateMultiplier:    1,
		Enabled:           true,
	}
	if cfg == nil || !enabled {
		return memberLevelState(defaultLevel, nil, totalRecharged)
	}
	levels := make([]MemberLevel, 0, len(cfg.Levels))
	for _, level := range cfg.Levels {
		if level.Enabled {
			levels = append(levels, level)
		}
	}
	sort.SliceStable(levels, func(i, j int) bool {
		if levels[i].MinRechargeAmount == levels[j].MinRechargeAmount {
			return levels[i].SortOrder < levels[j].SortOrder
		}
		return levels[i].MinRechargeAmount < levels[j].MinRechargeAmount
	})
	if len(levels) == 0 {
		return memberLevelState(defaultLevel, nil, totalRecharged)
	}
	var (
		current    MemberLevel
		hasCurrent bool
	)
	for _, level := range levels {
		if level.MinRechargeAmount <= totalRecharged {
			current = level
			hasCurrent = true
			continue
		}
		if hasCurrent {
			return memberLevelState(current, &level.MinRechargeAmount, totalRecharged)
		}
		return memberLevelState(defaultLevel, &level.MinRechargeAmount, totalRecharged)
	}
	if hasCurrent {
		return memberLevelState(current, nil, totalRecharged)
	}
	return memberLevelState(defaultLevel, &levels[0].MinRechargeAmount, totalRecharged)
}

func memberLevelState(level MemberLevel, nextThreshold *float64, totalRecharged float64) *MemberLevelState {
	current := level.MinRechargeAmount
	target := current
	progress := 100.0
	if nextThreshold != nil && *nextThreshold > current {
		target = *nextThreshold
		progress = ((totalRecharged - current) / (target - current)) * 100
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
	}
	return &MemberLevelState{
		LevelID:          level.ID,
		LevelName:        level.Name,
		RateMultiplier:   level.RateMultiplier,
		TotalRecharged:   totalRecharged,
		CurrentThreshold: current,
		NextThreshold:    nextThreshold,
		Progress:         progress,
		ProgressCurrent:  totalRecharged,
		ProgressTarget:   target,
	}
}

func (s *PaymentService) isRechargeKeyBilling(apiKey *APIKey, subscription *UserSubscription) bool {
	if subscription != nil {
		return false
	}
	if apiKey == nil || apiKey.Group == nil {
		return true
	}
	return !apiKey.Group.IsSubscriptionType()
}
