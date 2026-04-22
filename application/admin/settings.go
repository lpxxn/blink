package admin

import (
	"context"
	"errors"
	"strings"

	domainsettings "github.com/lpxxn/blink/domain/settings"
)

const (
	SettingSensitivePostModeKey = "sensitive_post_mode"

	SensitivePostModeAutoRemove = "auto_remove"
	SensitivePostModeReview     = "review"

	SettingRegisterEmailVerificationRequiredKey = "register_email_verification_required"
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
		if errors.Is(err, domainsettings.ErrNotFound) {
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

// GetRegisterEmailVerificationRequired returns whether registration requires email verification.
// Defaults to false when the key is missing.
func (s *Service) GetRegisterEmailVerificationRequired(ctx context.Context) (bool, error) {
	if s.Settings == nil {
		return false, ErrSettingsNotConfigured
	}
	v, err := s.Settings.GetString(ctx, SettingRegisterEmailVerificationRequiredKey)
	if err != nil {
		if errors.Is(err, domainsettings.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "true" {
		return true, nil
	}
	return false, nil
}

func (s *Service) SetRegisterEmailVerificationRequired(ctx context.Context, required bool) error {
	if s.Settings == nil {
		return ErrSettingsNotConfigured
	}
	str := "false"
	if required {
		str = "true"
	}
	return s.Settings.SetString(ctx, SettingRegisterEmailVerificationRequiredKey, str)
}
