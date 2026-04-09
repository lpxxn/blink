package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, id int64, ip, device string) error
	ListForAdmin(ctx context.Context, offset, limit int) ([]AdminListEntry, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatusRole(ctx context.Context, id int64, status *int, role *string) error
}
