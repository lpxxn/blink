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
