package gormdb

import (
	"context"
	"testing"

	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestPostReplyRepository_HideSubtree(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &PostReplyRepository{DB: db}
	ctx := context.Background()
	postID := int64(100)
	root := &domainpostreply.Reply{
		ID:    1,
		PostID: postID,
		UserID: 10,
		Body:   "root",
		Status: domainpostreply.StatusVisible,
	}
	child := &domainpostreply.Reply{
		ID:            2,
		PostID:        postID,
		UserID:        11,
		ParentReplyID: &root.ID,
		Body:          "child",
		Status:        domainpostreply.StatusVisible,
	}
	grand := &domainpostreply.Reply{
		ID:            3,
		PostID:        postID,
		UserID:        12,
		ParentReplyID: &child.ID,
		Body:          "grand",
		Status:        domainpostreply.StatusVisible,
	}
	for _, r := range []*domainpostreply.Reply{root, child, grand} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.HideSubtree(ctx, root.ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{1, 2, 3} {
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != domainpostreply.StatusHidden {
			t.Fatalf("id %d status=%d want hidden", id, got.Status)
		}
	}
}

func TestPostReplyRepository_ListByPostIDAllStatuses_And_UnhideSubtree(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &PostReplyRepository{DB: db}
	ctx := context.Background()
	postID := int64(101)
	root := &domainpostreply.Reply{
		ID:    11,
		PostID: postID,
		UserID: 10,
		Body:   "root",
		Status: domainpostreply.StatusVisible,
	}
	child := &domainpostreply.Reply{
		ID:            12,
		PostID:        postID,
		UserID:        11,
		ParentReplyID: &root.ID,
		Body:          "child",
		Status:        domainpostreply.StatusVisible,
	}
	grand := &domainpostreply.Reply{
		ID:            13,
		PostID:        postID,
		UserID:        12,
		ParentReplyID: &child.ID,
		Body:          "grand",
		Status:        domainpostreply.StatusVisible,
	}
	for _, r := range []*domainpostreply.Reply{root, child, grand} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	if err := repo.HideSubtree(ctx, root.ID); err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListByPostIDAllStatuses(ctx, postID, nil, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("len(list)=%d want 3", len(list))
	}
	for _, r := range list {
		if r.Status != domainpostreply.StatusHidden {
			t.Fatalf("id %d status=%d want hidden", r.ID, r.Status)
		}
	}

	if err := repo.UnhideSubtree(ctx, root.ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{11, 12, 13} {
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != domainpostreply.StatusVisible {
			t.Fatalf("id %d status=%d want visible", id, got.Status)
		}
	}
}
