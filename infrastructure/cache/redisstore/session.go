package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	domainsession "github.com/lpxxn/blink/domain/session"
	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix      = "blink:session:"
	userSessionsPrefix = "blink:user_sessions:"
)

type SessionStore struct {
	Client *redis.Client
}

type sessionPayload struct {
	UserID int64  `json:"user_id"`
	IP     string `json:"ip,omitempty"`
	UA     string `json:"ua,omitempty"`
}

func (s *SessionStore) Create(ctx context.Context, userID int64, ttl time.Duration, ip, ua string) (token string, err error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	p := sessionPayload{UserID: userID, IP: ip, UA: ua}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	key := sessionPrefix + id
	userKey := userSessionsPrefix + strconv.FormatInt(userID, 10)
	_, err = s.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, key, b, ttl)
		pipe.SAdd(ctx, userKey, id)
		pipe.Expire(ctx, userKey, ttl)
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *SessionStore) Get(ctx context.Context, token string) (*domainsession.Session, error) {
	key := sessionPrefix + token
	raw, err := s.Client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, domainsession.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var p sessionPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	ttl, err := s.Client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if ttl < 0 {
		return nil, domainsession.ErrNotFound
	}
	exp := time.Now().Add(ttl)
	return &domainsession.Session{ID: token, UserID: p.UserID, ExpiresAt: exp}, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	key := sessionPrefix + token
	raw, err := s.Client.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return err
	}
	if err == nil {
		var p sessionPayload
		if json.Unmarshal(raw, &p) == nil && p.UserID != 0 {
			userKey := userSessionsPrefix + strconv.FormatInt(p.UserID, 10)
			_ = s.Client.SRem(ctx, userKey, token).Err()
		}
	}
	return s.Client.Del(ctx, key).Err()
}

func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID int64) error {
	userKey := userSessionsPrefix + strconv.FormatInt(userID, 10)
	tokens, err := s.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return s.Client.Del(ctx, userKey).Err()
	}
	pipe := s.Client.Pipeline()
	for _, tok := range tokens {
		pipe.Del(ctx, sessionPrefix+tok)
	}
	pipe.Del(ctx, userKey)
	_, err = pipe.Exec(ctx)
	return err
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
