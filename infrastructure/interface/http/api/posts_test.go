package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	apppost "github.com/lpxxn/blink/application/post"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainpost "github.com/lpxxn/blink/domain/post"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

type createPostTestCatRepo struct{}

func (createPostTestCatRepo) Create(context.Context, *domaincategory.Category) error { panic("ni") }
func (createPostTestCatRepo) GetByID(_ context.Context, id int64) (*domaincategory.Category, error) {
	return &domaincategory.Category{ID: id}, nil
}
func (createPostTestCatRepo) ListActive(context.Context) ([]*domaincategory.Category, error) {
	panic("ni")
}
func (createPostTestCatRepo) Count(context.Context) (int64, error) { panic("ni") }

type createPostTestPostRepo struct {
	lastCreated *domainpost.Post
	byID        map[int64]*domainpost.Post
}

func (createPostTestPostRepo) Update(context.Context, *domainpost.Post) error   { panic("ni") }
func (createPostTestPostRepo) SoftDelete(context.Context, int64) error            { panic("ni") }
func (r *createPostTestPostRepo) GetByID(_ context.Context, id int64) (*domainpost.Post, error) {
	if r.byID == nil || r.byID[id] == nil {
		return nil, domainpost.ErrNotFound
	}
	return r.byID[id], nil
}
func (createPostTestPostRepo) ListPublicFeed(context.Context, *int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (createPostTestPostRepo) ListByUserID(context.Context, int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (createPostTestPostRepo) AdminList(context.Context, domainpost.AdminListFilters, int, int) ([]*domainpost.Post, int64, error) {
	panic("ni")
}
func (createPostTestPostRepo) Count(context.Context) (int64, error) { panic("ni") }
func (createPostTestPostRepo) CountCreatedSince(context.Context, time.Time) (int64, error) {
	panic("ni")
}

func (r *createPostTestPostRepo) Create(_ context.Context, p *domainpost.Post) error {
	r.lastCreated = p
	now := time.Now().UTC()
	cp := *p
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = now
	}
	if r.byID == nil {
		r.byID = map[int64]*domainpost.Post{}
	}
	r.byID[p.ID] = &cp
	return nil
}

func injectTestUserID(uid int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(httpauth.ContextUserIDKey, uid)
		c.Next()
	}
}

func newCreatePostTestServer(t *testing.T) (*Server, *createPostTestPostRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	postRepo := &createPostTestPostRepo{}
	svc := &apppost.Service{
		Posts:      postRepo,
		Categories: createPostTestCatRepo{},
		NewID:      func() int64 { return 9001 },
	}
	return &Server{Posts: svc}, postRepo
}

func TestCreatePost_singleImage(t *testing.T) {
	srv, postRepo := newCreatePostTestServer(t)
	r := gin.New()
	r.POST("/api/posts", injectTestUserID(42), srv.CreatePost)

	body := map[string]any{
		"body":         "hello single",
		"category_id":  "0",
		"images":       []string{"/uploads/a.jpg"},
		"draft":        false,
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var got PostJSON
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Images) != 1 || got.Images[0] != "/uploads/a.jpg" {
		t.Fatalf("response images: %#v", got.Images)
	}
	if got.UserID != 42 {
		t.Fatalf("user_id: got %d", got.UserID)
	}
	if postRepo.lastCreated == nil || len(postRepo.lastCreated.Images) != 1 {
		t.Fatalf("persisted images: %+v", postRepo.lastCreated)
	}
}

func TestCreatePost_multipleImages(t *testing.T) {
	srv, postRepo := newCreatePostTestServer(t)
	r := gin.New()
	r.POST("/api/posts", injectTestUserID(42), srv.CreatePost)

	want := []string{"/uploads/1.png", "/uploads/2.png", "/uploads/3.png"}
	body := map[string]any{
		"body":         "hello multi",
		"category_id":  "0",
		"images":       want,
		"draft":        false,
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var got PostJSON
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Images) != len(want) {
		t.Fatalf("len(images)=%d want %d: %#v", len(got.Images), len(want), got.Images)
	}
	for i := range want {
		if got.Images[i] != want[i] {
			t.Fatalf("images[%d]=%q want %q", i, got.Images[i], want[i])
		}
	}
	if postRepo.lastCreated == nil || len(postRepo.lastCreated.Images) != len(want) {
		t.Fatalf("persisted images: %+v", postRepo.lastCreated)
	}
}
