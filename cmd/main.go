package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	glsqlite "github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	apigen "github.com/lpxxn/blink/api/gen"
	appauth "github.com/lpxxn/blink/application/auth"
	appidp "github.com/lpxxn/blink/application/idp"
	appoauth "github.com/lpxxn/blink/application/oauth"
	oauthadapter "github.com/lpxxn/blink/infrastructure/adapter/oauth2"
	redisstore "github.com/lpxxn/blink/infrastructure/cache/redisstore"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	httpidp "github.com/lpxxn/blink/infrastructure/interface/http/idp"
	httpoauth "github.com/lpxxn/blink/infrastructure/interface/http/oauth"
	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/migrator"
)

func main() {
	ctx := context.Background()

	dsn := getenv("BLINK_DATABASE_DSN", "file:./data/blink.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	gdb, err := gorm.Open(glsqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	sqldb, err := gdb.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()
	sqldb.SetMaxOpenConns(1)
	if err := sqldb.Ping(); err != nil {
		log.Fatal(err)
	}
	migDir := getenv("BLINK_MIGRATIONS_DIR", "platform/db")
	if err := migrator.Run(sqldb, "sqlite", migDir); err != nil {
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

	userRepo := &gormdb.UserRepository{DB: gdb}
	oauthRepo := &gormdb.OAuthRepository{DB: gdb}
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
			Tx:         &gormdb.TxRunner{DB: gdb},
			Sessions:   sessStore,
			SessionTTL: 7 * 24 * time.Hour,
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	apigen.RegisterHandlers(r, openapiServer{})
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	oauthG := r.Group("/auth/oauth")
	h.Mount(oauthG)
	if idpHTTP != nil {
		idpHTTP.Mount(r.Group("/auth/idp"))
		r.POST("/auth/register", gin.WrapF(regHTTP.Register))
	}

	addr := getenv("BLINK_HTTP_ADDR", ":11110")
	log.Printf("listening on %s (OAuth providers: %d, builtin IdP: %v)", addr, len(providers), idpHTTP != nil)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

// openapiServer implements generated OpenAPI routes (see api/openapi/openapi.yaml).
type openapiServer struct{}

func (openapiServer) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, apigen.HealthResponse{Status: "ok"})
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
