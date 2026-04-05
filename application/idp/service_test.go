package idp

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"

	appauth "github.com/lpxxn/blink/application/auth"
	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"
)

// memStore is an in-memory AuthCodeStore + AccessTokenStore for tests (avoids importing infrastructure).
type memStore struct {
	codes  map[string]AuthCodePayload
	tokens map[string]int64
}

func (m *memStore) SaveAuthCode(_ context.Context, code string, p AuthCodePayload, _ time.Duration) error {
	if m.codes == nil {
		m.codes = make(map[string]AuthCodePayload)
	}
	m.codes[code] = p
	return nil
}

func (m *memStore) ConsumeAuthCode(_ context.Context, code string) (AuthCodePayload, error) {
	p, ok := m.codes[code]
	if !ok {
		return AuthCodePayload{}, ErrInvalidGrant
	}
	delete(m.codes, code)
	return p, nil
}

func (m *memStore) SaveAccessToken(_ context.Context, token string, userID int64, _ time.Duration) error {
	if m.tokens == nil {
		m.tokens = make(map[string]int64)
	}
	m.tokens[token] = userID
	return nil
}

func (m *memStore) GetUserIDByAccessToken(_ context.Context, token string) (int64, error) {
	uid, ok := m.tokens[token]
	if !ok {
		return 0, ErrUnauthorized
	}
	return uid, nil
}

func TestService_LoginAndExchange(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	userRepo := &gormdb.UserRepository{DB: db}
	oauthRepo := &gormdb.OAuthRepository{DB: db}
	_, err := (&appauth.RegisterService{
		Users:      userRepo,
		Identities: oauthRepo,
		Node:       node,
	}).RegisterWithPassword(context.Background(), "idp@example.com", "password12", "I")
	if err != nil {
		t.Fatal(err)
	}

	store := &memStore{}
	svc := &Service{
		ClientID:     "blink",
		ClientSecret: "sec",
		AllowedRedirect: map[string]struct{}{
			"http://localhost/cb": {},
		},
		Users:      userRepo,
		Identities: oauthRepo,
		Codes:      store,
		Access:     store,
		Node:       node,
		CodeTTL:    time.Minute,
		AccessTTL:  time.Hour,
	}
	ctx := context.Background()
	loc, err := svc.LoginWithPassword(ctx, "blink", "http://localhost/cb", "st1", "idp@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	pu, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	code := pu.Query().Get("code")
	if code == "" {
		t.Fatalf("no code in %q", loc)
	}
	tok, exp, err := svc.Exchange(ctx, "authorization_code", code, "http://localhost/cb", "blink", "sec")
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || exp <= 0 {
		t.Fatalf("token %q exp %d", tok, exp)
	}
	sub, email, name, err := svc.UserInfo(ctx, tok)
	if err != nil {
		t.Fatal(err)
	}
	if email != "idp@example.com" || name != "I" || sub == "" {
		t.Fatalf("userinfo sub=%q email=%q name=%q", sub, email, name)
	}
}
