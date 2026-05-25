package user

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, id int64, ip, device string) error
	ListForAdmin(ctx context.Context, query string, offset, limit int) ([]AdminListEntry, error)
	CountForAdmin(ctx context.Context, query string) (int64, error)
	// ListSnowflakeIDsByRole returns user ids with the given role (e.g. RoleSuperAdmin).
	ListSnowflakeIDsByRole(ctx context.Context, role string) ([]int64, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatusRole(ctx context.Context, id int64, status *int, role *string) error
	// UpdateName sets display name (trimmed); used by profile settings.
	UpdateName(ctx context.Context, id int64, name string) error
	// UpdatePasswordHash sets bcrypt password hash (admin reset).
	UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error
	// TopActiveUsers returns users ranked by total activity (posts + replies + likes) in [since, until).
	TopActiveUsers(ctx context.Context, since, until time.Time, limit int) ([]UserActivity, error)
}

// UserActivity is a ranked user entry for the activity leaderboard.
type UserActivity struct {
	UserID     int64
	Name       string
	Email      string
	PostCount  int64
	ReplyCount int64
	LikeCount  int64
	Total      int64
}
