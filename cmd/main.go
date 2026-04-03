package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	appauth "github.com/lpxxn/blink/application/auth"
	appidp "github.com/lpxxn/blink/application/idp"
	appoauth "github.com/lpxxn/blink/application/oauth"
	"github.com/lpxxn/blink/internal/migrator"
	redisstore "github.com/lpxxn/blink/infrastructure/cache/redisstore"
	oauthadapter "github.com/lpxxn/blink/infrastructure/adapter/oauth2"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	httpidp "github.com/lpxxn/blink/infrastructure/interface/http/idp"
	httpoauth "github.com/lpxxn/blink/infrastructure/interface/http/oauth"
	sqlrepo "github.com/lpxxn/blink/infrastructure/persistence/sql"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()

	dsn := getenv("BLINK_DATABASE_DSN", "file:./data/blink.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	migDir := getenv("BLINK_MIGRATIONS_DIR", "platform/db")
	if err := migrator.Run(db, "sqlite", migDir); err != nil {
		log.Fatal(err)
	}

	redisAddr := getenv("BLINK_REDIS_ADDR", "127.0.0.1:6379")
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	nodeID, _ := strconv.ParseInt(getenv("BLINK_SNOWFLAKE_NODE", "1"), 10, 64)
	if nodeID < 0 || nodeID > 1023 {
		log.Fatal("BLINK_SNOWFLAKE_NODE must be 0..1023")
	}
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := &sqlrepo.UserRepository{DB: db}
	oauthRepo := &sqlrepo.OAuthRepository{DB: db}
	sessStore := &redisstore.SessionStore{Client: rdb}
	stateStore := &redisstore.OAuthStateStore{Client: rdb}

	providers := map[string]appoauth.OAuth2Provider{}
	if cid, secret, redir := os.Getenv("OAUTH_GOOGLE_CLIENT_ID"), os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"), os.Getenv("OAUTH_GOOGLE_REDIRECT_URL"); cid != "" && secret != "" && redir != "" {
		providers["google"] = &oauthadapter.Provider{
			Config: &oauth2.Config{
				ClientID:     cid,
				ClientSecret: secret,
				RedirectURL:  redir,
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
			},
			UserInfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
		}
	}

	var idpHTTP *httpidp.Handler
	var regHTTP *httpauth.RegisterHandler

	publicBase := strings.TrimSpace(getenv("BLINK_PUBLIC_BASE_URL", ""))
	publicBase = strings.TrimRight(publicBase, "/")
	oauthSecret := strings.TrimSpace(getenv("BLINK_OAUTH_CLIENT_SECRET", ""))
	if publicBase != "" && oauthSecret != "" {
		oauthClientID := getenv("BLINK_OAUTH_CLIENT_ID", "blink")
		redirectUR := strings.TrimSpace(getenv("BLINK_OAUTH_REDIRECT_URL", ""))
		if redirectUR == "" {
			redirectUR = publicBase + "/auth/oauth/builtin/callback"
		}
		providers["builtin"] = &oauthadapter.Provider{
			Config: &oauth2.Config{
				ClientID:     oauthClientID,
				ClientSecret: oauthSecret,
				RedirectURL:  redirectUR,
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  publicBase + "/auth/idp/authorize",
					TokenURL: publicBase + "/auth/idp/token",
				},
			},
			UserInfoURL: publicBase + "/auth/idp/userinfo",
		}
		idpTok := &redisstore.IdPTokenStore{Client: rdb}
		idpSvc := &appidp.Service{
			ClientID:     oauthClientID,
			ClientSecret: oauthSecret,
			AllowedRedirect: map[string]struct{}{
				redirectUR: {},
			},
			Users:      userRepo,
			Identities: oauthRepo,
			Codes:      idpTok,
			Access:     idpTok,
			Node:       node,
			CodeTTL:    5 * time.Minute,
			AccessTTL:  time.Hour,
		}
		idpHTTP = &httpidp.Handler{Svc: idpSvc, FormAction: "/auth/idp/authorize"}
		regHTTP = &httpauth.RegisterHandler{Svc: &appauth.RegisterService{
			Users:      userRepo,
			Identities: oauthRepo,
			Node:       node,
		}}
	}

	svc := &appoauth.LoginService{
		Users:      userRepo,
		Identities: oauthRepo,
		Sessions:   sessStore,
		States:     stateStore,
		Providers:  providers,
		Node:       node,
		StateTTL:   10 * time.Minute,
		SessionTTL: 7 * 24 * time.Hour,
	}

	h := &httpoauth.Handler{Svc: svc}

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Mount("/auth/oauth", h.Routes())
	if idpHTTP != nil {
		r.Mount("/auth/idp", idpHTTP.Routes())
		r.Post("/auth/register", regHTTP.Register)
	}

	addr := getenv("BLINK_HTTP_ADDR", ":8080")
	log.Printf("listening on %s (OAuth providers: %d, builtin IdP: %v)", addr, len(providers), idpHTTP != nil)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
