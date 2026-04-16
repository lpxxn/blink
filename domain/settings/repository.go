package settings

import "context"

type Repository interface {
	GetString(ctx context.Context, key string) (string, error)
	SetString(ctx context.Context, key, value string) error
}

