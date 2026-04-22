package admin

import (
	"context"
	"errors"
	"testing"

	domainsettings "github.com/lpxxn/blink/domain/settings"
)

type fakeSettingsRepo struct {
	store        map[string]string
	getErr       error
	setErr       error
	lastSetKey   string
	lastSetValue string
}

func (f *fakeSettingsRepo) GetString(ctx context.Context, key string) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	if f.store == nil {
		return "", domainsettings.ErrNotFound
	}
	v, ok := f.store[key]
	if !ok {
		return "", domainsettings.ErrNotFound
	}
	return v, nil
}

func (f *fakeSettingsRepo) SetString(ctx context.Context, key, value string) error {
	if f.setErr != nil {
		return f.setErr
	}
	if f.store == nil {
		f.store = make(map[string]string)
	}
	f.store[key] = value
	f.lastSetKey = key
	f.lastSetValue = value
	return nil
}

func TestGetRegisterEmailVerificationRequired_DefaultFalseWhenMissing(t *testing.T) {
	s := &Service{Settings: &fakeSettingsRepo{}}
	got, err := s.GetRegisterEmailVerificationRequired(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != false {
		t.Fatalf("expected false when missing, got %v", got)
	}
}

func TestSetRegisterEmailVerificationRequired_PersistTrueFalse(t *testing.T) {
	f := &fakeSettingsRepo{}
	s := &Service{Settings: f}
	// set true
	if err := s.SetRegisterEmailVerificationRequired(context.Background(), true); err != nil {
		t.Fatalf("set true err: %v", err)
	}
	if f.lastSetKey != SettingRegisterEmailVerificationRequiredKey {
		t.Fatalf("expected key %q, got %q", SettingRegisterEmailVerificationRequiredKey, f.lastSetKey)
	}
	if f.lastSetValue != "true" {
		t.Fatalf("expected stored true, got %q", f.lastSetValue)
	}
	got, err := s.GetRegisterEmailVerificationRequired(context.Background())
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if got != true {
		t.Fatalf("expected true, got %v", got)
	}
	// set false
	if err := s.SetRegisterEmailVerificationRequired(context.Background(), false); err != nil {
		t.Fatalf("set false err: %v", err)
	}
	if f.lastSetKey != SettingRegisterEmailVerificationRequiredKey {
		t.Fatalf("expected key %q, got %q", SettingRegisterEmailVerificationRequiredKey, f.lastSetKey)
	}
	if f.lastSetValue != "false" {
		t.Fatalf("expected stored false, got %q", f.lastSetValue)
	}
	got2, err := s.GetRegisterEmailVerificationRequired(context.Background())
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if got2 != false {
		t.Fatalf("expected false, got %v", got2)
	}
}

func TestGetSetRegisterEmailVerificationRequired_SettingsNil_ReturnsErrSettingsNotConfigured(t *testing.T) {
	s := &Service{Settings: nil}
	_, err := s.GetRegisterEmailVerificationRequired(context.Background())
	if !errors.Is(err, ErrSettingsNotConfigured) {
		t.Fatalf("expected ErrSettingsNotConfigured, got %v", err)
	}
	if err := s.SetRegisterEmailVerificationRequired(context.Background(), true); !errors.Is(err, ErrSettingsNotConfigured) {
		t.Fatalf("expected ErrSettingsNotConfigured, got %v", err)
	}
}
