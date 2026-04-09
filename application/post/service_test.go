package post

import (
	"context"
	"errors"
	"testing"
	"time"

	appcategory "github.com/lpxxn/blink/application/category"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainpost "github.com/lpxxn/blink/domain/post"
)

type stubCatRepo struct {
	err error
}

func (s *stubCatRepo) Create(context.Context, *domaincategory.Category) error { panic("ni") }
func (s *stubCatRepo) GetByID(context.Context, int64) (*domaincategory.Category, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &domaincategory.Category{ID: 1}, nil
}
func (s *stubCatRepo) ListActive(context.Context) ([]*domaincategory.Category, error) { panic("ni") }
func (s *stubCatRepo) Count(context.Context) (int64, error) { panic("ni") }

type stubPostRepo struct {
	created *domainpost.Post
}

func (stubPostRepo) Update(context.Context, *domainpost.Post) error   { panic("ni") }
func (stubPostRepo) SoftDelete(context.Context, int64) error          { panic("ni") }
func (stubPostRepo) GetByID(context.Context, int64) (*domainpost.Post, error) {
	panic("ni")
}
func (stubPostRepo) ListPublicFeed(context.Context, *int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (stubPostRepo) ListByUserID(context.Context, int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (stubPostRepo) AdminList(context.Context, domainpost.AdminListFilters, int, int) ([]*domainpost.Post, int64, error) {
	panic("ni")
}
func (stubPostRepo) Count(context.Context) (int64, error)         { panic("ni") }
func (stubPostRepo) CountCreatedSince(context.Context, time.Time) (int64, error) {
	panic("ni")
}

func (s *stubPostRepo) Create(_ context.Context, p *domainpost.Post) error {
	s.created = p
	return nil
}

func TestService_Create_rejectsUnknownCategory(t *testing.T) {
	ctx := context.Background()
	pr := &stubPostRepo{}
	svc := &Service{
		Posts: pr,
		Categories: &stubCatRepo{
			err: domaincategory.ErrNotFound,
		},
		NewID: func() int64 { return 42 },
	}
	cid := int64(9)
	_, err := svc.Create(ctx, 1, "hello", &cid, nil, false)
	if !errors.Is(err, appcategory.ErrInvalidCategory) {
		t.Fatalf("err=%v", err)
	}
}

func TestService_Create_persistsWhenCategoryOK(t *testing.T) {
	ctx := context.Background()
	pr := &stubPostRepo{}
	svc := &Service{
		Posts:      pr,
		Categories: &stubCatRepo{},
		NewID:      func() int64 { return 42 },
	}
	cid := int64(9)
	p, err := svc.Create(ctx, 1, "hello", &cid, []string{"/u/x.png"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != 42 || pr.created == nil || pr.created.CategoryID == nil || *pr.created.CategoryID != 9 {
		t.Fatalf("unexpected post: %+v", pr.created)
	}
}
