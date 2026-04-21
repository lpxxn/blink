package httpauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"

	appauth "github.com/lpxxn/blink/application/auth"
	appemailcode "github.com/lpxxn/blink/application/emailcode"
	mailinfra "github.com/lpxxn/blink/infrastructure/mail"
	"github.com/lpxxn/blink/infrastructure/cache/redisstore"
	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRegisterHandler_RegisterWithSession_SetsCookie(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	svc := &appauth.RegisterService{
		Users:      u,
		Identities: o,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		SessionTTL: time.Hour,
	}
	h := &RegisterHandler{Svc: svc}

	body, _ := json.Marshal(map[string]string{
		"email":    "handler@example.com",
		"password": "password12",
		"name":     "H",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["session_token"] == nil || out["session_token"] == "" {
		t.Fatalf("missing session_token: %#v", out)
	}
	var sid string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "blink_session" {
			sid = c.Value
			break
		}
	}
	if sid == "" {
		t.Fatal("expected blink_session cookie")
	}
}

// When Codes is configured, the handler MUST reject a request whose verification
// code does not match what was issued. Regression guard for the "wrong code still
// registers" bug.
func TestRegisterHandler_WrongCode_Rejected(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	codes := &appemailcode.Service{
		Store:  &redisstore.EmailCodeStore{Client: rdb},
		Mailer: &mailinfra.LogMailer{},
	}
	svc := &appauth.RegisterService{
		Users:      u,
		Identities: o,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		SessionTTL: time.Hour,
		Codes:      codes,
	}
	h := &RegisterHandler{Svc: svc}

	// 1) Send a real code so the store has something to compare against.
	if err := codes.Send(t.Context(), appemailcode.PurposeRegister, "wrong@example.com", ""); err != nil {
		t.Fatalf("seed send: %v", err)
	}

	// 2) Submit register with an obviously wrong code.
	body, _ := json.Marshal(map[string]string{
		"email":    "wrong@example.com",
		"password": "password12",
		"name":     "H",
		"code":     "000000",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong code, got %d body=%s", rr.Code, rr.Body.String())
	}

	// 3) User must NOT have been created.
	if _, err := u.FindByEmail(t.Context(), "wrong@example.com"); err == nil {
		t.Fatal("user was created despite wrong code — verification bypass bug")
	}
}

// Submitting register without ever requesting a code must also fail.
func TestRegisterHandler_NoCodeRequested_Rejected(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	codes := &appemailcode.Service{
		Store:  &redisstore.EmailCodeStore{Client: rdb},
		Mailer: &mailinfra.LogMailer{},
	}
	svc := &appauth.RegisterService{
		Users:      u,
		Identities: o,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		SessionTTL: time.Hour,
		Codes:      codes,
	}
	h := &RegisterHandler{Svc: svc}

	body, _ := json.Marshal(map[string]string{
		"email":    "nocode@example.com",
		"password": "password12",
		"name":     "H",
		"code":     "123456",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when no code was issued, got %d body=%s", rr.Code, rr.Body.String())
	}
	if _, err := u.FindByEmail(t.Context(), "nocode@example.com"); err == nil {
		t.Fatal("user was created despite no code ever issued")
	}
}
