package httpauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	domainsession "github.com/lpxxn/blink/domain/session"
)

type stubStore struct {
	sessions map[string]*domainsession.Session
}

func (s *stubStore) Create(_ context.Context, _ int64, _ time.Duration, _, _ string) (string, error) {
	panic("not used")
}

func (s *stubStore) Get(_ context.Context, token string) (*domainsession.Session, error) {
	if s.sessions == nil {
		return nil, domainsession.ErrNotFound
	}
	vs, ok := s.sessions[token]
	if !ok {
		return nil, domainsession.ErrNotFound
	}
	return vs, nil
}

func (s *stubStore) Delete(_ context.Context, _ string) error { return nil }

func (s *stubStore) DeleteAllForUser(context.Context, int64) error { return nil }

func TestRequireSession_AuthorizationBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"t1": {ID: "t1", UserID: 42, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}

	r := gin.New()
	r.GET("/me", RequireSession(store), func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer t1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRequireSession_CookieFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"c1": {ID: "c1", UserID: 7, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}

	r := gin.New()
	r.GET("/me", RequireSession(store), func(c *gin.Context) {
		uid, _ := UserIDFromContext(c)
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "blink_session", Value: "c1"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestOptionalSession_setsUserWhenTokenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"ot": {ID: "ot", UserID: 100, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	r := gin.New()
	r.GET("/", OptionalSession(store, nil), func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"user_id": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ot")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestOptionalSession_noUserWhenNoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &stubStore{}
	r := gin.New()
	r.GET("/", OptionalSession(store, nil), func(c *gin.Context) {
		_, ok := UserIDFromContext(c)
		if ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
}

