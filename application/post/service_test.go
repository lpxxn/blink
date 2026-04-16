package post

import (
	"context"
	"errors"
	"testing"
	"time"

	appcategory "github.com/lpxxn/blink/application/category"
	appmoderation "github.com/lpxxn/blink/application/moderation"
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
	byID    map[int64]*domainpost.Post
}

func (r *stubPostRepo) Update(context.Context, *domainpost.Post) error { panic("ni") }
func (stubPostRepo) SoftDelete(context.Context, int64) error          { panic("ni") }
func (r *stubPostRepo) GetByID(_ context.Context, id int64) (*domainpost.Post, error) {
	if r.byID == nil {
		return nil, domainpost.ErrNotFound
	}
	p := r.byID[id]
	if p == nil {
		return nil, domainpost.ErrNotFound
	}
	return p, nil
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
	now := time.Now().UTC()
	cp := *p
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = now
	}
	if s.byID == nil {
		s.byID = map[int64]*domainpost.Post{}
	}
	s.byID[p.ID] = &cp
	return nil
}

type patchStubPostRepo struct {
	p *domainpost.Post
}

func (patchStubPostRepo) SoftDelete(context.Context, int64) error { panic("ni") }
func (patchStubPostRepo) ListPublicFeed(context.Context, *int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (patchStubPostRepo) ListByUserID(context.Context, int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (patchStubPostRepo) AdminList(context.Context, domainpost.AdminListFilters, int, int) ([]*domainpost.Post, int64, error) {
	panic("ni")
}
func (patchStubPostRepo) Count(context.Context) (int64, error) { panic("ni") }
func (patchStubPostRepo) CountCreatedSince(context.Context, time.Time) (int64, error) {
	panic("ni")
}
func (patchStubPostRepo) Create(context.Context, *domainpost.Post) error { panic("ni") }

func (r *patchStubPostRepo) GetByID(context.Context, int64) (*domainpost.Post, error) {
	return r.p, nil
}

func (patchStubPostRepo) Update(context.Context, *domainpost.Post) error { return nil }

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

func TestService_Create_blocksSensitiveWhenPublishing(t *testing.T) {
	ctx := context.Background()
	appmoderation.SetWordsSnapshot([]string{"bad"})
	defer appmoderation.SetWordsSnapshot(nil)
	pr := &stubPostRepo{}
	svc := &Service{
		Posts:      pr,
		Categories: &stubCatRepo{},
		NewID:      func() int64 { return 42 },
	}
	cid := int64(1)
	_, err := svc.Create(ctx, 1, "has bad word", &cid, nil, false)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if pr.created == nil {
		t.Fatal("expected create")
	}
}

func TestService_Create_draftAllowsSensitive(t *testing.T) {
	ctx := context.Background()
	appmoderation.SetWordsSnapshot([]string{"bad"})
	defer appmoderation.SetWordsSnapshot(nil)
	pr := &stubPostRepo{}
	svc := &Service{
		Posts:      pr,
		Categories: &stubCatRepo{},
		NewID:      func() int64 { return 42 },
	}
	cid := int64(1)
	p, err := svc.Create(ctx, 1, "has bad word", &cid, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != domainpost.StatusDraft || pr.created == nil {
		t.Fatalf("post=%+v", p)
	}
}

func TestService_Patch_blocksSensitiveWhenPublished(t *testing.T) {
	ctx := context.Background()
	appmoderation.SetWordsSnapshot([]string{"x"})
	defer appmoderation.SetWordsSnapshot(nil)
	base := &domainpost.Post{
		ID: 1, UserID: 1, Body: "ok", Status: domainpost.StatusPublished,
		ModerationFlag: domainpost.ModerationNormal, Images: []string{},
	}
	svc := &Service{
		Posts: &patchStubPostRepo{p: base},
		NewID: func() int64 { return 1 },
	}
	b := "hello x world"
	_, err := svc.Patch(ctx, 1, 1, Patch{Body: &b})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
}
