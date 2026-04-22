package httpadmin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	appadmin "github.com/lpxxn/blink/application/admin"
	domainsettings "github.com/lpxxn/blink/domain/settings"
)

type stubAdminSettingsRepo struct {
	store map[string]string
}

func (s *stubAdminSettingsRepo) GetString(ctx context.Context, key string) (string, error) {
	if s.store == nil {
		return "", domainsettings.ErrNotFound
	}
	v, ok := s.store[key]
	if !ok {
		return "", domainsettings.ErrNotFound
	}
	return v, nil
}

func (s *stubAdminSettingsRepo) SetString(ctx context.Context, key, value string) error {
	if s.store == nil {
		s.store = make(map[string]string)
	}
	s.store[key] = value
	return nil
}

func TestServer_GetRegisterEmailVerificationRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{Admin: &appadmin.Service{Settings: &stubAdminSettingsRepo{}}}

	r := gin.New()
	r.GET("/", s.GetRegisterEmailVerificationRequired)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if got := out["required"]; got != false {
		t.Fatalf("expected false, got %v", got)
	}
}

func TestServer_SetRegisterEmailVerificationRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubAdminSettingsRepo{}
	s := &Server{Admin: &appadmin.Service{Settings: repo}}

	r := gin.New()
	r.PUT("/", s.SetRegisterEmailVerificationRequired)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"required":true}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if got := out["required"]; got != true {
		t.Fatalf("expected true, got %v", got)
	}
	if got := repo.store[appadmin.SettingRegisterEmailVerificationRequiredKey]; got != "true" {
		t.Fatalf("expected stored true, got %q", got)
	}
}

func TestServer_GetWhenSettingsNil_ReturnsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{Admin: &appadmin.Service{Settings: nil}}

	r := gin.New()
	r.GET("/", s.GetRegisterEmailVerificationRequired)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when admin settings nil, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestServer_PutWhenSettingsNil_ReturnsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{Admin: &appadmin.Service{Settings: nil}}

	r := gin.New()
	r.PUT("/", s.SetRegisterEmailVerificationRequired)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"required":true}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when admin settings nil on PUT, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// erroring repo to simulate unexpected errors from settings store
type erringRepo struct{}

func (e *erringRepo) GetString(ctx context.Context, key string) (string, error) {
	return "", errors.New("boom")
}

func (e *erringRepo) SetString(ctx context.Context, key, value string) error { return nil }

func TestServer_GetSettingsRepoError_ReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{Admin: &appadmin.Service{Settings: &erringRepo{}}}

	r := gin.New()
	r.GET("/", s.GetRegisterEmailVerificationRequired)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when settings repo errors, got %d body=%s", rr.Code, rr.Body.String())
	}
}
