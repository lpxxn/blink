package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, id int64, ip, device string) error
	ListForAdmin(ctx context.Context, offset, limit int) ([]AdminListEntry, error)
	// ListSnowflakeIDsByRole returns user ids with the given role (e.g. RoleSuperAdmin).
	ListSnowflakeIDsByRole(ctx context.Context, role string) ([]int64, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatusRole(ctx context.Context, id int64, status *int, role *string) error
	// UpdateName sets display name (trimmed); used by profile settings.
	UpdateName(ctx context.Context, id int64, name string) error
}
