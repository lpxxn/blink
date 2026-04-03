package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	domainsession "github.com/lpxxn/blink/domain/session"
	"github.com/redis/go-redis/v9"
)

const sessionPrefix = "blink:session:"

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
	if err := s.Client.Set(ctx, key, b, ttl).Err(); err != nil {
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
	return s.Client.Del(ctx, sessionPrefix+token).Err()
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
