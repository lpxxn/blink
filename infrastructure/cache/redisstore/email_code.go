package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// EmailCodeStore is the Redis-backed implementation of emailcode.Store.
// Keys are namespaced by purpose to keep registration / change-password /
// reset-password codes separate.
type EmailCodeStore struct {
	Client *redis.Client
}

const (
	emailCodePrefix     = "email_code:"
	emailCodeCoolPrefix = "email_code_cool:"
	emailCodeEmailRL    = "email_code_rl:"
	emailCodeActorRL    = "email_code_iprl:"
	emailCodeFailPrefix = "email_code_fail:"
)

func codeKey(purpose, email string) string     { return emailCodePrefix + purpose + ":" + email }
func coolKey(purpose, email string) string     { return emailCodeCoolPrefix + purpose + ":" + email }
func emailRLKey(purpose, email string) string  { return emailCodeEmailRL + purpose + ":" + email }
func actorRLKey(purpose, actor string) string  { return emailCodeActorRL + purpose + ":" + actor }
func failKey(purpose, email string) string     { return emailCodeFailPrefix + purpose + ":" + email }

// IsCoolingDown returns true if the caller must wait before sending again.
func (s *EmailCodeStore) IsCoolingDown(ctx context.Context, purpose, email string) (bool, error) {
	n, err := s.Client.Exists(ctx, coolKey(purpose, email)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// IncrWindowAndGet increments a fixed-hour window counter and returns the new
// value. The window expires after windowTTL on first write.
func (s *EmailCodeStore) incrWindow(ctx context.Context, key string, windowTTL time.Duration) (int64, error) {
	n, err := s.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = s.Client.Expire(ctx, key, windowTTL).Err()
	}
	return n, nil
}

// HourlyEmailCount is the per-email send count in the current hour.
func (s *EmailCodeStore) HourlyEmailCount(ctx context.Context, purpose, email string) (int64, error) {
	v, err := s.Client.Get(ctx, emailRLKey(purpose, email)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}

// HourlyActorCount is the per-ip (or per-user-id) send count in the current hour.
func (s *EmailCodeStore) HourlyActorCount(ctx context.Context, purpose, actor string) (int64, error) {
	v, err := s.Client.Get(ctx, actorRLKey(purpose, actor)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}

// PutCode stores the verification code and primes cool/limit counters. It does
// NOT increment counters when called multiple times for the same email before
// TTL expiry; callers must check limits before invoking PutCode.
func (s *EmailCodeStore) PutCode(ctx context.Context, purpose, email, code, actor string, codeTTL, coolTTL, windowTTL time.Duration) error {
	pipe := s.Client.Pipeline()
	pipe.Set(ctx, codeKey(purpose, email), code, codeTTL)
	pipe.Del(ctx, failKey(purpose, email))
	pipe.Set(ctx, coolKey(purpose, email), "1", coolTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	if _, err := s.incrWindow(ctx, emailRLKey(purpose, email), windowTTL); err != nil {
		return err
	}
	if actor != "" {
		if _, err := s.incrWindow(ctx, actorRLKey(purpose, actor), windowTTL); err != nil {
			return err
		}
	}
	return nil
}

// GetCode returns the stored code for (purpose,email) or empty string if none.
func (s *EmailCodeStore) GetCode(ctx context.Context, purpose, email string) (string, error) {
	v, err := s.Client.Get(ctx, codeKey(purpose, email)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return v, err
}

// IncrFail increments the failure counter for (purpose,email) and returns the
// new value. TTL defaults to codeTTL (via EXPIRE on first increment).
func (s *EmailCodeStore) IncrFail(ctx context.Context, purpose, email string, ttl time.Duration) (int64, error) {
	n, err := s.Client.Incr(ctx, failKey(purpose, email)).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = s.Client.Expire(ctx, failKey(purpose, email), ttl).Err()
	}
	return n, nil
}

// ClearCode removes the code and failure counter (called after success / lockout).
func (s *EmailCodeStore) ClearCode(ctx context.Context, purpose, email string) error {
	return s.Client.Del(ctx, codeKey(purpose, email), failKey(purpose, email)).Err()
}
