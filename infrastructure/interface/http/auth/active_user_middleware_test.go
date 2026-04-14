package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestRequireActiveUser_allowsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.OpenSQLiteMemory(t)
	repo := &gormdb.UserRepository{DB: db}
	ctx := context.Background()
	if err := repo.Create(ctx, &domainuser.User{
		SnowflakeID:  91001,
		Email:        "active@example.com",
		Name:         "a",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         domainuser.RoleUser,
	}); err != nil {
		t.Fatal(err)
	}
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"tok": {ID: "tok", UserID: 91001, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	r := gin.New()
	r.GET("/", RequireSession(store), RequireActiveUser(store, repo), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestRequireActiveUser_rejectsBanned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.OpenSQLiteMemory(t)
	repo := &gormdb.UserRepository{DB: db}
	ctx := context.Background()
	if err := repo.Create(ctx, &domainuser.User{
		SnowflakeID:  91002,
		Email:        "banned@example.com",
		Name:         "b",
		PasswordHash: "h",
		Status:       domainuser.StatusBanned,
		Role:         domainuser.RoleUser,
	}); err != nil {
		t.Fatal(err)
	}
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"btok": {ID: "btok", UserID: 91002, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	r := gin.New()
	r.GET("/", RequireSession(store), RequireActiveUser(store, repo), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer btok")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestOptionalSession_skipsBannedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.OpenSQLiteMemory(t)
	repo := &gormdb.UserRepository{DB: db}
	ctx := context.Background()
	if err := repo.Create(ctx, &domainuser.User{
		SnowflakeID:  91003,
		Email:        "ban2@example.com",
		Name:         "c",
		PasswordHash: "h",
		Status:       domainuser.StatusBanned,
		Role:         domainuser.RoleUser,
	}); err != nil {
		t.Fatal(err)
	}
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"xtok": {ID: "xtok", UserID: 91003, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	r := gin.New()
	r.GET("/", OptionalSession(store, repo), func(c *gin.Context) {
		_, ok := UserIDFromContext(c)
		if ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer xtok")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
}
