package postlike

import (
	"context"
	"time"
)

// Repository persists post likes (post_likes table).
type Repository interface {
	Like(ctx context.Context, userID, postID int64) error
	Unlike(ctx context.Context, userID, postID int64) error
	IsLiked(ctx context.Context, userID, postID int64) (bool, error)
	CountByPostID(ctx context.Context, postID int64) (int64, error)
	CountByPostIDs(ctx context.Context, postIDs []int64) (map[int64]int64, error)
	LikedPostIDs(ctx context.Context, userID int64, postIDs []int64) (map[int64]bool, error)
	// TopLikedPosts ranks posts by like count in [since, until); only public published posts.
	TopLikedPosts(ctx context.Context, since, until time.Time, limit int) ([]PostLikeRank, error)
}

// PostLikeRank is a post ranked by likes received in a time window.
type PostLikeRank struct {
	PostID    int64
	LikeCount int64
}
