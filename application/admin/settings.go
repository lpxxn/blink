package admin

import (
	"context"
	"strings"

	domainsettings "github.com/lpxxn/blink/domain/settings"
)

const (
	SettingSensitivePostModeKey = "sensitive_post_mode"

	SensitivePostModeAutoRemove = "auto_remove"
	SensitivePostModeReview     = "review"
)

func validSensitivePostMode(v string) bool {
	switch v {
	case SensitivePostModeAutoRemove, SensitivePostModeReview:
		return true
	default:
		return false
	}
}

// GetSensitivePostMode returns the stored mode or default "review" when unset.
func (s *Service) GetSensitivePostMode(ctx context.Context) (string, error) {
	if s.Settings == nil {
		return SensitivePostModeReview, ErrSettingsNotConfigured
	}
	v, err := s.Settings.GetString(ctx, SettingSensitivePostModeKey)
	if err != nil {
		if err == domainsettings.ErrNotFound {
			return SensitivePostModeReview, nil
		}
		return "", err
	}
	v = strings.TrimSpace(strings.ToLower(v))
	if !validSensitivePostMode(v) {
		return SensitivePostModeReview, nil
	}
	return v, nil
}

func (s *Service) SetSensitivePostMode(ctx context.Context, mode string) error {
	if s.Settings == nil {
		return ErrSettingsNotConfigured
	}
	mode = strings.TrimSpace(strings.ToLower(mode))
	if !validSensitivePostMode(mode) {
		return ErrInvalidSetting
	}
	return s.Settings.SetString(ctx, SettingSensitivePostModeKey, mode)
}

