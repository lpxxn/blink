package gormdb

import (
	"context"
	"testing"
	"time"

	domainpost "github.com/lpxxn/blink/domain/post"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestPostRepository_ListFollowingFeed(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	postRepo := &PostRepository{DB: db}
	ctx := context.Background()

	viewer := int64(1000)
	followed := int64(2000)
	other := int64(3000)
	now := time.Now().UTC()

	for _, u := range []domainuser.User{
		{SnowflakeID: viewer, Email: "v@example.com", Name: "V", Status: domainuser.StatusActive, Role: "user"},
		{SnowflakeID: followed, Email: "f@example.com", Name: "F", Status: domainuser.StatusActive, Role: "user"},
		{SnowflakeID: other, Email: "o@example.com", Name: "O", Status: domainuser.StatusActive, Role: "user"},
	} {
		if err := (&UserRepository{DB: db}).Create(ctx, &u); err != nil {
			t.Fatal(err)
		}
	}

	followRepo := &FollowRepository{DB: db}
	if err := followRepo.Follow(ctx, viewer, followed); err != nil {
		t.Fatal(err)
	}

	makePost := func(id, author int64, body string) {
		p := &domainpost.Post{
			ID:             id,
			UserID:         author,
			PostType:       domainpost.TypeOriginal,
			Visibility:     domainpost.VisibilityPublic,
			Body:           body,
			Status:         domainpost.StatusPublished,
			ModerationFlag: domainpost.ModerationNormal,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := postRepo.Create(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	makePost(5003, other, "other user")
	makePost(5002, followed, "from followed")
	makePost(5001, followed, "older followed")

	list, err := postRepo.ListFollowingFeed(ctx, viewer, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(list), list)
	}
	if list[0].ID != 5002 || list[1].ID != 5001 {
		t.Fatalf("order: %+v", list)
	}

	cursor := list[1].ID
	page2, err := postRepo.ListFollowingFeed(ctx, viewer, &cursor, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 0 {
		t.Fatalf("page2: %+v", page2)
	}
}

func TestPostRepository_SearchPublic(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	postRepo := &PostRepository{DB: db}
	ctx := context.Background()
	now := time.Now().UTC()

	for _, p := range []*domainpost.Post{
		{
			ID: 6001, UserID: 1, PostType: domainpost.TypeOriginal, Visibility: domainpost.VisibilityPublic,
			Body: "hello golang world", Status: domainpost.StatusPublished, ModerationFlag: domainpost.ModerationNormal,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: 6002, UserID: 1, PostType: domainpost.TypeOriginal, Visibility: domainpost.VisibilityPublic,
			Body: "unrelated", Status: domainpost.StatusPublished, ModerationFlag: domainpost.ModerationNormal,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: 6003, UserID: 1, PostType: domainpost.TypeOriginal, Visibility: domainpost.VisibilityPublic,
			Body: "draft golang", Status: domainpost.StatusDraft, ModerationFlag: domainpost.ModerationNormal,
			CreatedAt: now, UpdatedAt: now,
		},
	} {
		if err := postRepo.Create(ctx, p); err != nil {
			t.Fatal(err)
		}
	}

	list, err := postRepo.SearchPublic(ctx, "golang", nil, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != 6001 {
		t.Fatalf("search posts: %+v", list)
	}
}

func TestUserRepository_SearchPublic(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	userRepo := &UserRepository{DB: db}
	ctx := context.Background()

	for _, u := range []domainuser.User{
		{SnowflakeID: 7001, Email: "a@example.com", Name: "Alice", Status: domainuser.StatusActive, Role: "user"},
		{SnowflakeID: 7002, Email: "b@example.com", Name: "Bob", Status: domainuser.StatusActive, Role: "user"},
		{SnowflakeID: 7003, Email: "c@example.com", Name: "Alicia", Status: domainuser.StatusBanned, Role: "user"},
	} {
		if err := userRepo.Create(ctx, &u); err != nil {
			t.Fatal(err)
		}
	}

	list, err := userRepo.SearchPublic(ctx, "Ali", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].SnowflakeID != 7001 {
		t.Fatalf("search users: %+v", list)
	}
}
