package httpauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubRegisterConfigSettings struct {
	required bool
	err      error
}

func (s *stubRegisterConfigSettings) GetRegisterEmailVerificationRequired(context.Context) (bool, error) {
	return s.required, s.err
}

func TestRegisterConfigHandler_GetReturnsVerificationRequired(t *testing.T) {
	h := &RegisterConfigHandler{Settings: &stubRegisterConfigSettings{required: false}}

	req := httptest.NewRequest(http.MethodGet, "/auth/register/config", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if got := out["email_verification_required"]; got != false {
		t.Fatalf("expected false, got %v", got)
	}
}

func TestRegisterConfigHandler_GetReturnsVerificationRequiredTrue(t *testing.T) {
	h := &RegisterConfigHandler{Settings: &stubRegisterConfigSettings{required: true}}

	req := httptest.NewRequest(http.MethodGet, "/auth/register/config", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if got := out["email_verification_required"]; got != true {
		t.Fatalf("expected true, got %v", got)
	}
}

func TestRegisterConfigHandler_GetSettingsError(t *testing.T) {
	h := &RegisterConfigHandler{Settings: &stubRegisterConfigSettings{err: errors.New("boom")}}

	req := httptest.NewRequest(http.MethodGet, "/auth/register/config", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegisterConfigHandler_GetSettingsNil(t *testing.T) {
	h := &RegisterConfigHandler{Settings: nil}

	req := httptest.NewRequest(http.MethodGet, "/auth/register/config", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rr.Code, rr.Body.String())
	}
}
