package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func (s *SettingService) IsAffiliateSignupRewardEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateSignupRewardEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func isSimpleSettingID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func (s *SettingService) IsPhoneVerifyEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyPhoneVerifyEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *SettingService) GetPromptArchiveSettings(ctx context.Context) (*PromptArchiveSettingsView, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyPromptArchiveSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return &PromptArchiveSettingsView{GroupIDs: []int64{}}, nil
		}
		return nil, fmt.Errorf("get prompt archive settings: %w", err)
	}
	if strings.TrimSpace(value) == "" {
		return &PromptArchiveSettingsView{GroupIDs: []int64{}}, nil
	}
	var settings PromptArchiveSettingsView
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return &PromptArchiveSettingsView{GroupIDs: []int64{}}, nil
	}
	if settings.GroupIDs == nil {
		settings.GroupIDs = []int64{}
	}
	return &settings, nil
}

func (s *SettingService) GetAffiliateSignupRewardAmount(ctx context.Context) float64 {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateSignupRewardAmount)
	if err != nil {
		return AffiliateSignupRewardAmountDefault
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return AffiliateSignupRewardAmountDefault
	}
	return amount
}

func (s *SettingService) SetPromptArchiveSettings(ctx context.Context, settings *PromptArchiveSettingsView) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if settings.GroupIDs == nil {
		settings.GroupIDs = []int64{}
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal prompt archive settings: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyPromptArchiveSettings, string(data))
}

func publicHomeLinksJSON(raw string) string {
	normalized, err := normalizeHomeLinks(raw)
	if err != nil || len(normalized) == 0 {
		return defaultHomeLinksJSON
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return defaultHomeLinksJSON
	}
	return string(b)
}

func normalizeHomeLinks(raw string) ([]HomeLink, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = defaultHomeLinksJSON
	}

	var links []HomeLink
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		return nil, infraerrors.BadRequest("HOME_LINKS_INVALID", "home links must be a JSON array")
	}
	if len(links) > 20 {
		return nil, infraerrors.BadRequest("HOME_LINKS_TOO_MANY", "home links cannot exceed 20 items")
	}
	if len(links) == 0 {
		return []HomeLink{}, nil
	}

	sort.SliceStable(links, func(i, j int) bool {
		return links[i].SortOrder < links[j].SortOrder
	})

	normalized := make([]HomeLink, 0, len(links))
	seenIDs := make(map[string]struct{}, len(links))
	for _, link := range links {
		label := strings.TrimSpace(link.Label)
		labelZH := strings.TrimSpace(link.LabelZH)
		labelEN := strings.TrimSpace(link.LabelEN)
		if label == "" {
			label = firstNonEmpty(labelZH, labelEN)
		}
		if labelZH == "" {
			labelZH = firstNonEmpty(label, labelEN)
		}
		if labelEN == "" {
			labelEN = firstNonEmpty(label, labelZH)
		}
		if label == "" && labelZH == "" && labelEN == "" {
			return nil, infraerrors.BadRequest("HOME_LINK_LABEL_REQUIRED", "home link label is required")
		}
		if len([]rune(label)) > 50 {
			return nil, infraerrors.BadRequest("HOME_LINK_LABEL_TOO_LONG", "home link label is too long")
		}
		if len([]rune(labelZH)) > 50 || len([]rune(labelEN)) > 50 {
			return nil, infraerrors.BadRequest("HOME_LINK_LABEL_TOO_LONG", "home link label is too long")
		}

		linkURL := strings.TrimSpace(link.URL)
		if linkURL == "" {
			return nil, infraerrors.BadRequest("HOME_LINK_URL_REQUIRED", "home link url is required")
		}
		parsed, err := url.ParseRequestURI(linkURL)
		if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, infraerrors.BadRequest("HOME_LINK_URL_INVALID", "home link url must be an absolute http(s) URL")
		}

		id := strings.TrimSpace(link.ID)
		if id == "" {
			id = fmt.Sprintf("home-link-%d", len(normalized)+1)
		}
		if len(id) > 32 || !isSimpleSettingID(id) {
			id = fmt.Sprintf("home-link-%d", len(normalized)+1)
		}
		if _, exists := seenIDs[id]; exists {
			id = fmt.Sprintf("home-link-%d", len(normalized)+1)
		}
		seenIDs[id] = struct{}{}

		normalized = append(normalized, HomeLink{
			ID:        id,
			Label:     label,
			LabelZH:   labelZH,
			LabelEN:   labelEN,
			URL:       linkURL,
			Enabled:   link.Enabled,
			SortOrder: len(normalized),
		})
	}

	return normalized, nil
}
