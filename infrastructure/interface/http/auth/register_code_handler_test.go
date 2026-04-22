package httpauth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lpxxn/blink/application/emailcode"
	domainuser "github.com/lpxxn/blink/domain/user"
)

type stubRegisterCodeSettings struct {
	required bool
	err      error
}

func (s *stubRegisterCodeSettings) GetRegisterEmailVerificationRequired(context.Context) (bool, error) {
	return s.required, s.err
}

type stubRegisterCodeUsers struct {
	calls []string
	found map[string]bool
}

func (s *stubRegisterCodeUsers) Create(context.Context, *domainuser.User) error { return nil }

func (s *stubRegisterCodeUsers) GetByID(context.Context, int64) (*domainuser.User, error) {
	return nil, domainuser.ErrNotFound
}

func (s *stubRegisterCodeUsers) FindByEmail(_ context.Context, email string) (*domainuser.User, error) {
	s.calls = append(s.calls, email)
	if s.found != nil && s.found[email] {
		return &domainuser.User{Email: email}, nil
	}
	return nil, domainuser.ErrNotFound
}

func (s *stubRegisterCodeUsers) UpdateLastLogin(context.Context, int64, string, string) error {
	return nil
}

func (s *stubRegisterCodeUsers) ListForAdmin(context.Context, int, int) ([]domainuser.AdminListEntry, error) {
	return nil, nil
}

func (s *stubRegisterCodeUsers) ListSnowflakeIDsByRole(context.Context, string) ([]int64, error) {
	return nil, nil
}

func (s *stubRegisterCodeUsers) Count(context.Context) (int64, error) { return 0, nil }

func (s *stubRegisterCodeUsers) UpdateStatusRole(context.Context, int64, *int, *string) error {
	return nil
}

func (s *stubRegisterCodeUsers) UpdateName(context.Context, int64, string) error { return nil }

func (s *stubRegisterCodeUsers) UpdatePasswordHash(context.Context, int64, string) error { return nil }

func TestRegisterCodeHandler_SendReturnsServiceUnavailableBeforeAccountLookupWhenCodesNil(t *testing.T) {
	users := &stubRegisterCodeUsers{found: map[string]bool{"existing@example.com": true}}
	h := &RegisterCodeHandler{
		Users:    users,
		Settings: &stubRegisterCodeSettings{required: true},
	}

	for _, tc := range []struct {
		name  string
		email string
	}{
		{name: "existing", email: "existing@example.com"},
		{name: "missing", email: "missing@example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/register/send_code", bytes.NewBufferString(`{"email":"`+tc.email+`"}`))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Send(rr, req)

			if rr.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503 when codes are nil, got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}

	if len(users.calls) != 0 {
		t.Fatalf("expected no account lookup before codes check, got calls=%v", users.calls)
	}
}

func TestRegisterCodeHandler_SendSkipsCodeFlowWhenVerificationDisabled(t *testing.T) {
	h := &RegisterCodeHandler{
		Codes:    &emailcode.Service{},
		Settings: &stubRegisterCodeSettings{required: false},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register/send_code", bytes.NewBufferString(`{"email":"skip@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Send(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when verification is disabled, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegisterCodeHandler_SendSettingsErrorReturnsInternalServerError(t *testing.T) {
	h := &RegisterCodeHandler{
		Codes:    &emailcode.Service{},
		Settings: &stubRegisterCodeSettings{err: context.Canceled},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register/send_code", bytes.NewBufferString(`{"email":"boom@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Send(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when settings read fails, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestClientIPKey_UsesFirstForwardedForEntry(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/register/send_code", nil)
	req.Header.Set("X-Forwarded-For", " 203.0.113.9 , 198.51.100.1, 198.51.100.2 ")

	if got := clientIPKey(req); got != "203.0.113.9" {
		t.Fatalf("expected first forwarded-for entry, got %q", got)
	}
}
