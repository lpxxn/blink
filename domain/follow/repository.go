package follow

import "context"

// Repository persists one-way follow edges (user_follows table).
type Repository interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
	// ListFollowing returns followee user ids for follower userID, newest first (by followee id desc).
	ListFollowing(ctx context.Context, userID int64, beforeUserID *int64, limit int) ([]int64, error)
	// ListFollowers returns follower user ids for followee userID, newest first (by follower id desc).
	ListFollowers(ctx context.Context, userID int64, beforeUserID *int64, limit int) ([]int64, error)
}
