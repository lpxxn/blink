package follow

import "context"

// Repository persists one-way follow edges (user_follows table).
type Repository interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
	// ListFollowing returns users that userID follows, newest follow first.
	ListFollowing(ctx context.Context, userID int64, cursor *PageCursor, limit int) ([]ListEntry, error)
	// ListFollowers returns users following userID, newest follow first.
	ListFollowers(ctx context.Context, userID int64, cursor *PageCursor, limit int) ([]ListEntry, error)
}
