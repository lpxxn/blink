package search

import (
	"context"
	"errors"
	"testing"
	"time"

	domainpost "github.com/lpxxn/blink/domain/post"
	domainuser "github.com/lpxxn/blink/domain/user"
)

type stubPostSearchRepo struct {
	posts []*domainpost.Post
}

func (s *stubPostSearchRepo) Create(context.Context, *domainpost.Post) error { panic("ni") }
func (s *stubPostSearchRepo) Update(context.Context, *domainpost.Post) error { panic("ni") }
func (s *stubPostSearchRepo) SoftDelete(context.Context, int64) error        { panic("ni") }
func (s *stubPostSearchRepo) GetByID(context.Context, int64) (*domainpost.Post, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) ListPublicFeed(context.Context, *int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) ListFollowingFeed(context.Context, int64, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) SearchPublic(_ context.Context, query string, _ *int64, _ int) ([]*domainpost.Post, error) {
	if query != "go" {
		return nil, nil
	}
	return s.posts, nil
}
func (s *stubPostSearchRepo) ListPublicByUserID(context.Context, int64, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) ListByUserID(context.Context, int64, bool, *int64, int) ([]*domainpost.Post, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) AdminList(context.Context, domainpost.AdminListFilters, int, int) ([]*domainpost.Post, int64, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) CountAdmin(context.Context, domainpost.AdminListFilters) (int64, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) Count(context.Context) (int64, error) { panic("ni") }
func (s *stubPostSearchRepo) CountCreatedSince(context.Context, time.Time) (int64, error) {
	panic("ni")
}
func (s *stubPostSearchRepo) TopPosters(context.Context, time.Time, time.Time, int) ([]domainpost.UserPostCount, error) {
	panic("ni")
}

type stubUserSearchRepo struct {
	profiles []domainuser.PublicProfile
}

func (s *stubUserSearchRepo) Create(context.Context, *domainuser.User) error { panic("ni") }
func (s *stubUserSearchRepo) GetByID(context.Context, int64) (*domainuser.User, error) {
	panic("ni")
}
func (s *stubUserSearchRepo) FindByEmail(context.Context, string) (*domainuser.User, error) {
	panic("ni")
}
func (s *stubUserSearchRepo) UpdateLastLogin(context.Context, int64, string, string) error {
	panic("ni")
}
func (s *stubUserSearchRepo) ListForAdmin(context.Context, string, int, int) ([]domainuser.AdminListEntry, error) {
	panic("ni")
}
func (s *stubUserSearchRepo) CountForAdmin(context.Context, string) (int64, error) { panic("ni") }
func (s *stubUserSearchRepo) ListSnowflakeIDsByRole(context.Context, string) ([]int64, error) {
	panic("ni")
}
func (s *stubUserSearchRepo) Count(context.Context) (int64, error) { panic("ni") }
func (s *stubUserSearchRepo) UpdateStatusRole(context.Context, int64, *int, *string) error {
	panic("ni")
}
func (s *stubUserSearchRepo) UpdateName(context.Context, int64, string) error { panic("ni") }
func (s *stubUserSearchRepo) UpdatePasswordHash(context.Context, int64, string) error {
	panic("ni")
}
func (s *stubUserSearchRepo) TopActiveUsers(context.Context, time.Time, time.Time, int) ([]domainuser.UserActivity, error) {
	panic("ni")
}
func (s *stubUserSearchRepo) SearchPublic(_ context.Context, query string, _ int, _ int) ([]domainuser.PublicProfile, error) {
	if query != "ali" {
		return nil, nil
	}
	return s.profiles, nil
}

func TestService_SearchPosts_rejectsEmptyQuery(t *testing.T) {
	svc := &Service{Posts: &stubPostSearchRepo{}}
	_, err := svc.SearchPosts(context.Background(), "  ", nil, 20)
	if !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("err=%v", err)
	}
}

func TestService_SearchPosts_trimsQuery(t *testing.T) {
	want := []*domainpost.Post{{ID: 1, Body: "go"}}
	svc := &Service{Posts: &stubPostSearchRepo{posts: want}}
	list, err := svc.SearchPosts(context.Background(), "  go  ", nil, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("list=%+v", list)
	}
}

func TestService_SearchUsers_trimsQuery(t *testing.T) {
	want := []domainuser.PublicProfile{{SnowflakeID: 9, Name: "Ali"}}
	svc := &Service{Users: &stubUserSearchRepo{profiles: want}}
	list, err := svc.SearchUsers(context.Background(), " ali ", 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].SnowflakeID != 9 {
		t.Fatalf("list=%+v", list)
	}
}
