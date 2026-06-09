package gormdb

import (
	"context"
	"testing"
	"time"

	domainfollow "github.com/lpxxn/blink/domain/follow"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestFollowRepository_ListFollowing_OrderAndPagination(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &FollowRepository{DB: db}
	ctx := context.Background()

	me := int64(1000)
	t1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)

	for _, row := range []UserFollowModel{
		{FollowerID: me, FolloweeID: 2001, CreatedAt: t1, UpdatedAt: t1},
		{FollowerID: me, FolloweeID: 2002, CreatedAt: t3, UpdatedAt: t3},
		{FollowerID: me, FolloweeID: 2003, CreatedAt: t2, UpdatedAt: t2},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}

	page1, err := repo.ListFollowing(ctx, me, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len=%d", len(page1))
	}
	if page1[0].UserID != 2002 || page1[1].UserID != 2003 {
		t.Fatalf("page1 order: %+v", page1)
	}

	cursor := &domainfollow.PageCursor{CreatedAt: page1[1].CreatedAt, UserID: page1[1].UserID}
	page2, err := repo.ListFollowing(ctx, me, cursor, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 1 || page2[0].UserID != 2001 {
		t.Fatalf("page2: %+v", page2)
	}
}

func TestFollowRepository_ListFollowers(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &FollowRepository{DB: db}
	ctx := context.Background()

	target := int64(3000)
	now := time.Now().UTC()
	if err := db.Create(&UserFollowModel{
		FollowerID: 4001, FolloweeID: target, CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatal(err)
	}

	list, err := repo.ListFollowers(ctx, target, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].UserID != 4001 {
		t.Fatalf("followers: %+v", list)
	}
}
