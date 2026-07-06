package replylike

import "context"

// Repository persists reply likes (reply_likes table).
type Repository interface {
	Like(ctx context.Context, userID, replyID int64) error
	Unlike(ctx context.Context, userID, replyID int64) error
	IsLiked(ctx context.Context, userID, replyID int64) (bool, error)
	CountByReplyID(ctx context.Context, replyID int64) (int64, error)
	CountByReplyIDs(ctx context.Context, replyIDs []int64) (map[int64]int64, error)
	LikedReplyIDs(ctx context.Context, userID int64, replyIDs []int64) (map[int64]bool, error)
}
