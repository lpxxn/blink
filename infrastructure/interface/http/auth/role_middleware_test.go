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
)

type roleStubUsers struct {
	role string
}

func (roleStubUsers) Create(context.Context, *domainuser.User) error { panic("unimplemented") }
func (roleStubUsers) FindByEmail(context.Context, string) (*domainuser.User, error) {
	panic("unimplemented")
}
func (roleStubUsers) UpdateLastLogin(context.Context, int64, string, string) error {
	panic("unimplemented")
}
func (roleStubUsers) ListForAdmin(context.Context, int, int) ([]domainuser.AdminListEntry, error) {
	panic("unimplemented")
}
func (roleStubUsers) ListSnowflakeIDsByRole(context.Context, string) ([]int64, error) {
	panic("unimplemented")
}
func (roleStubUsers) Count(context.Context) (int64, error) { panic("unimplemented") }
func (roleStubUsers) UpdateStatusRole(context.Context, int64, *int, *string) error {
	panic("unimplemented")
}
func (roleStubUsers) UpdateName(context.Context, int64, string) error { panic("unimplemented") }
func (roleStubUsers) UpdatePasswordHash(context.Context, int64, string) error {
	panic("unimplemented")
}

func (r roleStubUsers) GetByID(_ context.Context, id int64) (*domainuser.User, error) {
	return &domainuser.User{SnowflakeID: id, Role: r.role}, nil
}

func TestRequireUserRole_allowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"tok": {ID: "tok", UserID: 99, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	users := roleStubUsers{role: domainuser.RoleSuperAdmin}

	r := gin.New()
	r.GET("/x", RequireSession(store), RequireUserRole(users, domainuser.RoleSuperAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestRequireUserRole_forbiddenWhenRoleMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &stubStore{
		sessions: map[string]*domainsession.Session{
			"tok": {ID: "tok", UserID: 1, ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	users := roleStubUsers{role: domainuser.RoleUser}

	r := gin.New()
	r.GET("/x", RequireSession(store), RequireUserRole(users, domainuser.RoleSuperAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want 403", rr.Code)
	}
}
