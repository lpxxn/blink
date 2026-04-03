package session

import (
	"context"
	"time"
)

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

type Store interface {
	Create(ctx context.Context, userID int64, ttl time.Duration, ip, ua string) (token string, err error)
	Get(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
}
