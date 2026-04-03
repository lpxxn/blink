package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	domainuser "github.com/lpxxn/blink/domain/user"
	redisstore "github.com/lpxxn/blink/infrastructure/cache/redisstore"
	sqlrepo "github.com/lpxxn/blink/infrastructure/persistence/sql"
	"github.com/lpxxn/blink/internal/testutil"
)

type stubProvider struct {
	info UserInfo
}

func (s *stubProvider) AuthCodeURL(state string) string {
	_ = state
	return "https://example.com/oauth/authorize"
}

func (s *stubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	_ = code
	return &oauth2.Token{AccessToken: "test-token"}, nil
}

func (s *stubProvider) UserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	_ = ctx
	_ = token
	return s.info, nil
}

func TestLoginService_RegisterThenLogin(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	node, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatal(err)
	}
	svc := &LoginService{
		Users:      &sqlrepo.UserRepository{DB: db},
		Identities: &sqlrepo.OAuthRepository{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		States:     &redisstore.OAuthStateStore{Client: rdb},
		Providers: map[string]OAuth2Provider{
			"test": &stubProvider{info: UserInfo{Subject: "sub-1", Email: "a@example.com", Name: "Alice"}},
		},
		Node:       node,
		StateTTL:   10 * time.Minute,
		SessionTTL: 24 * time.Hour,
	}
	ctx := context.Background()

	authURL, state, err := svc.LoginRedirectURL(ctx, "test", "/dashboard")
	if err != nil {
		t.Fatal(err)
	}
	if authURL == "" || state == "" {
		t.Fatal("expected redirect url and state")
	}

	sessToken, next, err := svc.CompleteLogin(ctx, "test", "auth-code", state, "127.0.0.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if next != "/dashboard" {
		t.Fatalf("next: got %q", next)
	}
	if sessToken == "" {
		t.Fatal("expected session token")
	}

	s, err := svc.Sessions.Get(ctx, sessToken)
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.Users.GetByID(ctx, s.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "a@example.com" {
		t.Fatalf("email: %q", u.Email)
	}
	if u.Status != domainuser.StatusActive {
		t.Fatalf("status: %d", u.Status)
	}

	_, _, err = svc.CompleteLogin(ctx, "test", "auth-code", state, "127.0.0.1", "ua")
	if err == nil {
		t.Fatal("expected error reusing state")
	}

	authURL2, state2, err := svc.LoginRedirectURL(ctx, "test", "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = authURL2
	sessToken2, _, err := svc.CompleteLogin(ctx, "test", "auth-code", state2, "127.0.0.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := svc.Sessions.Get(ctx, sessToken2)
	if err != nil {
		t.Fatal(err)
	}
	if s2.UserID != s.UserID {
		t.Fatalf("login user mismatch: %d vs %d", s2.UserID, s.UserID)
	}
}

func TestLoginService_UnknownProvider(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	node, _ := snowflake.NewNode(1)
	svc := &LoginService{
		Users:      &sqlrepo.UserRepository{DB: db},
		Identities: &sqlrepo.OAuthRepository{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		States:     &redisstore.OAuthStateStore{Client: rdb},
		Providers:  map[string]OAuth2Provider{},
		Node:       node,
		StateTTL:   time.Minute,
		SessionTTL: time.Hour,
	}
	_, _, err = svc.LoginRedirectURL(context.Background(), "nope", "/")
	if err == nil {
		t.Fatal("expected error")
	}
}
