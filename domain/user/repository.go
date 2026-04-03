package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateLastLogin(ctx context.Context, id int64, ip, device string) error
}
