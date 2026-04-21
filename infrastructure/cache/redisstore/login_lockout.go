package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginLockoutStore is a simple fixed-window failure counter with a separate
// "locked" key. It implements application/auth.LoginLockout.
type LoginLockoutStore struct {
	Client *redis.Client
}

const (
	loginFailPrefix = "login_fail:"
	loginLockPrefix = "login_lock:"
)

func (s *LoginLockoutStore) IsLocked(ctx context.Context, email string) (bool, error) {
	n, err := s.Client.Exists(ctx, loginLockPrefix+email).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *LoginLockoutStore) Record(ctx context.Context, email string, threshold int64, window, lockout time.Duration) (bool, error) {
	key := loginFailPrefix + email
	n, err := s.Client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = s.Client.Expire(ctx, key, window).Err()
	}
	if n >= threshold {
		if err := s.Client.Set(ctx, loginLockPrefix+email, "1", lockout).Err(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (s *LoginLockoutStore) Reset(ctx context.Context, email string) error {
	return s.Client.Del(ctx, loginFailPrefix+email, loginLockPrefix+email).Err()
}
