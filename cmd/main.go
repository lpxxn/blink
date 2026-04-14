package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	glsqlite "github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	apigen "github.com/lpxxn/blink/api/gen"
	appadmin "github.com/lpxxn/blink/application/admin"
	appauth "github.com/lpxxn/blink/application/auth"
	appbootstrap "github.com/lpxxn/blink/application/bootstrap"
	appcategory "github.com/lpxxn/blink/application/category"
	appidp "github.com/lpxxn/blink/application/idp"
	appnotification "github.com/lpxxn/blink/application/notification"
	appoauth "github.com/lpxxn/blink/application/oauth"
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domainuser "github.com/lpxxn/blink/domain/user"
	oauthadapter "github.com/lpxxn/blink/infrastructure/adapter/oauth2"
	redisstore "github.com/lpxxn/blink/infrastructure/cache/redisstore"
	httpadmin "github.com/lpxxn/blink/infrastructure/interface/http/admin"
	httpapi "github.com/lpxxn/blink/infrastructure/interface/http/api"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	httpidp "github.com/lpxxn/blink/infrastructure/interface/http/idp"
	httpoauth "github.com/lpxxn/blink/infrastructure/interface/http/oauth"
	"github.com/lpxxn/blink/infrastructure/messaging"
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
	postRepo := &gormdb.PostRepository{DB: gdb}
	replyRepo := &gormdb.PostReplyRepository{DB: gdb}
	notifRepo := &gormdb.NotificationRepository{DB: gdb}
	catRepo := &gormdb.CategoryRepository{DB: gdb}
	sessStore := &redisstore.SessionStore{Client: rdb}
	stateStore := &redisstore.OAuthStateStore{Client: rdb}

	if err := appcategory.SeedDefaults(ctx, catRepo, func() int64 { return node.Generate().Int64() }); err != nil {
		log.Fatalf("seed categories: %v", err)
	}
	if err := appbootstrap.PromoteSuperAdminFromEnv(ctx, userRepo, getenv("BLINK_BOOTSTRAP_SUPER_ADMIN_EMAIL", "")); err != nil {
		log.Fatalf("bootstrap super admin: %v", err)
	}

	postSvc := &apppost.Service{
		Posts:      postRepo,
		Categories: catRepo,
		NewID:      func() int64 { return node.Generate().Int64() },
	}
	replySvc := &apppostreply.Service{
		Posts:   postSvc,
		Replies: replyRepo,
		NewID:   func() int64 { return node.Generate().Int64() },
	}
	notifSvc := &appnotification.Service{
		Repo:  notifRepo,
		NewID: func() int64 { return node.Generate().Int64() },
		Users: userRepo,
	}

	wmLogger := watermill.NewStdLogger(false, false)
	wmPublisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:     rdb,
		Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
	}, wmLogger)
	if err != nil {
		log.Fatalf("watermill redis publisher: %v", err)
	}
	defer func() { _ = wmPublisher.Close() }()
	notifyEventBus := messaging.NewNotificationWatermillPublisher(wmPublisher)
	postSvc.NotifyEvents = notifyEventBus

	if strings.TrimSpace(getenv("BLINK_DISABLE_NOTIFICATION_CONSUMER", "")) == "" {
		consumer := strings.TrimSpace(getenv("BLINK_WATERMILL_CONSUMER", ""))
		if consumer == "" {
			h, _ := os.Hostname()
			consumer = fmt.Sprintf("%s-%d", h, os.Getpid())
		}
		wmSubscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
			Client:                        rdb,
			Unmarshaller:                  redisstream.DefaultMarshallerUnmarshaller{},
			ConsumerGroup:                 getenv("BLINK_WATERMILL_CONSUMER_GROUP", "blink-notify"),
			Consumer:                      consumer,
			OldestId:                      getenv("BLINK_NOTIFICATION_STREAM_FROM", "$"),
			DisableIndefiniteInitialBlock: true,
		}, wmLogger)
		if err != nil {
			log.Fatalf("watermill redis subscriber: %v", err)
		}
		defer func() { _ = wmSubscriber.Close() }()
		wmRouter, err := messaging.RunNotificationWatermillRouter(context.Background(), wmSubscriber, notifSvc, sessStore, wmLogger)
		if err != nil {
			log.Fatalf("watermill notification router: %v", err)
		}
		defer func() { _ = wmRouter.Close() }()
	}

	adminSvc := &appadmin.Service{
		Users:        userRepo,
		Posts:        postRepo,
		Sessions:     sessStore,
		NotifyEvents: notifyEventBus,
	}

	uploadRoot := getenv("BLINK_UPLOAD_DIR", "data/uploads")
	if err := os.MkdirAll(uploadRoot, 0750); err != nil {
		log.Fatalf("upload dir: %v", err)
	}
	apiSrv := &httpapi.Server{
		Posts:         postSvc,
		Replies:       replySvc,
		Notifications: notifSvc,
		NotifyEvents:  notifyEventBus,
		Categories:    catRepo,
		Users:         userRepo,
		Sessions:      sessStore,
		UploadRoot:    uploadRoot,
		UploadURLPath: "/uploads",
	}
	adminSrv := &httpadmin.Server{
		Admin:         adminSvc,
		CategoryCount: catRepo.Count,
		Users:         userRepo,
	}

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

	regHTTP := &httpauth.RegisterHandler{Svc: &appauth.RegisterService{
		Users:      userRepo,
		Identities: oauthRepo,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: gdb},
		Sessions:   sessStore,
		SessionTTL: 7 * 24 * time.Hour,
	}}

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

	gin.SetMode(gin.DebugMode)
	r := gin.New()
	apigen.RegisterHandlers(r, openapiServer{})
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	oauthG := r.Group("/auth/oauth")
	h.Mount(oauthG)
	r.POST("/auth/register", gin.WrapF(regHTTP.Register))
	if idpHTTP != nil {
		idpHTTP.Mount(r.Group("/auth/idp"))
	}

	r.Static("/uploads", uploadRoot)
	r.Static("/web", "./web")

	api := r.Group("/api")
	api.GET("/categories", apiSrv.ListCategories)
	api.GET("/posts", apiSrv.ListPosts)
	api.GET("/posts/:id/replies", apiSrv.ListReplies)
	api.POST("/logout", apiSrv.Logout)
	opt := api.Group("")
	opt.Use(httpauth.OptionalSession(sessStore, userRepo))
	opt.GET("/posts/:id", apiSrv.GetPost)

	authed := api.Group("")
	authed.Use(httpauth.RequireSession(sessStore))
	authed.Use(httpauth.RequireActiveUser(sessStore, userRepo))
	authed.GET("/me", apiSrv.GetMe)
	authed.PATCH("/me", apiSrv.PatchMe)
	authed.POST("/posts", apiSrv.CreatePost)
	authed.PATCH("/posts/:id", apiSrv.PatchPost)
	authed.DELETE("/posts/:id", apiSrv.DeletePost)
	authed.POST("/posts/:id/moderation_request", apiSrv.SubmitModerationRequest)
	authed.GET("/me/posts", apiSrv.ListMyPosts)
	authed.GET("/me/notifications", apiSrv.ListNotifications)
	authed.GET("/me/notifications/unread_count", apiSrv.UnreadNotificationCount)
	authed.POST("/me/notifications/:id/read", apiSrv.MarkNotificationRead)
	authed.POST("/me/notifications/read_all", apiSrv.MarkAllNotificationsRead)
	authed.POST("/posts/:id/replies", apiSrv.CreateReply)
	authed.POST("/uploads", apiSrv.UploadImage)
	authed.DELETE("/replies/:id", apiSrv.DeleteReply)

	adminG := r.Group("/admin/api")
	adminG.Use(httpauth.RequireSession(sessStore))
	adminG.Use(httpauth.RequireActiveUser(sessStore, userRepo))
	adminG.Use(httpauth.RequireUserRole(userRepo, domainuser.RoleSuperAdmin))
	adminG.GET("/overview", adminSrv.Overview)
	adminG.GET("/users", adminSrv.ListUsers)
	adminG.PATCH("/users/:id", adminSrv.PatchUser)
	adminG.POST("/users/:id/reset_password", adminSrv.ResetUserPassword)
	adminG.GET("/posts", adminSrv.ListPosts)
	adminG.PATCH("/posts/:id", adminSrv.PatchPost)
	adminG.POST("/posts/:id/resolve_appeal", adminSrv.ResolveAppeal)

	addr := getenv("BLINK_HTTP_ADDR", ":11110")
	log.Printf("listening on %s (OAuth providers: %d, builtin IdP: %v, POST /auth/register: on)", addr, len(providers), idpHTTP != nil)
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
