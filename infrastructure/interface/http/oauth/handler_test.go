package httpoauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	appoauth "github.com/lpxxn/blink/application/oauth"
	redisstore "github.com/lpxxn/blink/infrastructure/cache/redisstore"
	sqlrepo "github.com/lpxxn/blink/infrastructure/persistence/sql"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestHandler_OAuthCallback_SetsCookie(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	node, _ := snowflake.NewNode(1)
	stub := &stubProviderHTTP{
		authBase: "https://idp.example/oauth?code=fake&state=",
		info:     appoauth.UserInfo{Subject: "http-1", Email: "h@example.com", Name: "H"},
	}
	svc := &appoauth.LoginService{
		Users:      &sqlrepo.UserRepository{DB: db},
		Identities: &sqlrepo.OAuthRepository{DB: db},
		Sessions:   &redisstore.SessionStore{Client: rdb},
		States:     &redisstore.OAuthStateStore{Client: rdb},
		Providers: map[string]appoauth.OAuth2Provider{
			"test": stub,
		},
		Node:       node,
		StateTTL:   10 * time.Minute,
		SessionTTL: 24 * time.Hour,
	}

	h := &Handler{Svc: svc}
	r := chi.NewRouter()
	r.Mount("/auth/oauth", h.Routes())

	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/test/login?next=/home", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("login: %d %s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	state := u.Query().Get("state")
	if state == "" {
		t.Fatalf("missing state in %q", loc)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/auth/oauth/test/callback?code=c&state="+url.QueryEscape(state), nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusFound {
		t.Fatalf("callback: %d %s", rr2.Code, rr2.Body.String())
	}
	if rr2.Header().Get("Location") != "/home" {
		t.Fatalf("redirect: %q", rr2.Header().Get("Location"))
	}
	var sid string
	for _, c := range rr2.Result().Cookies() {
		if c.Name == "blink_session" {
			sid = c.Value
			break
		}
	}
	if sid == "" || strings.TrimSpace(sid) == "" {
		t.Fatal("expected blink_session cookie")
	}
}

type stubProviderHTTP struct {
	authBase string
	info     appoauth.UserInfo
}

func (s *stubProviderHTTP) AuthCodeURL(state string) string {
	return s.authBase + url.QueryEscape(state)
}

func (s *stubProviderHTTP) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	_ = ctx
	_ = code
	return &oauth2.Token{AccessToken: "tok"}, nil
}

func (s *stubProviderHTTP) UserInfo(ctx context.Context, token *oauth2.Token) (appoauth.UserInfo, error) {
	_ = ctx
	_ = token
	return s.info, nil
}
