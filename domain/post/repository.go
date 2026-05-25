package post

import (
	"context"
	"time"
)

// AdminListFilters optional filters for admin post listing.
type AdminListFilters struct {
	UserID              *int64
	CategoryID          *int64
	ModerationFlag      *int
	AppealPending       bool // true → appeal_status = AppealPending（待处理申诉/复核）
	SensitiveHitPending bool // true → moderation_flag=违规 且备注为敏感词命中
	IncludeDeleted      bool
}

type Repository interface {
	Create(ctx context.Context, p *Post) error
	Update(ctx context.Context, p *Post) error
	SoftDelete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Post, error)
	ListPublicFeed(ctx context.Context, categoryID *int64, uncategorizedOnly bool, beforeID *int64, limit int) ([]*Post, error)
	ListByUserID(ctx context.Context, userID int64, includeDraft bool, beforeID *int64, limit int) ([]*Post, error)
	AdminList(ctx context.Context, f AdminListFilters, offset, limit int) ([]*Post, int64, error)
	CountAdmin(ctx context.Context, f AdminListFilters) (int64, error)
	Count(ctx context.Context) (int64, error)
	CountCreatedSince(ctx context.Context, t time.Time) (int64, error)
	// TopPosters returns user IDs ranked by post count in [since, until).
	TopPosters(ctx context.Context, since, until time.Time, limit int) ([]UserPostCount, error)
}

// UserPostCount is a (user_id, count) pair returned by ranking queries.
type UserPostCount struct {
	UserID    int64
	PostCount int64
}
