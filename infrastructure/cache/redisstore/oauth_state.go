package redisstore

import (
	"context"
	"encoding/json"
	"time"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	"github.com/redis/go-redis/v9"
)

const statePrefix = "blink:oauth:state:"

type OAuthStateStore struct {
	Client *redis.Client
}

func (s *OAuthStateStore) Save(ctx context.Context, state string, p domainoauth.RedirectState, ttl time.Duration) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.Client.Set(ctx, statePrefix+state, b, ttl).Err()
}

func (s *OAuthStateStore) Consume(ctx context.Context, state string) (domainoauth.RedirectState, error) {
	key := statePrefix + state
	raw, err := s.Client.GetDel(ctx, key).Bytes()
	if err == redis.Nil {
		return domainoauth.RedirectState{}, domainoauth.ErrInvalidState
	}
	if err != nil {
		return domainoauth.RedirectState{}, err
	}
	var p domainoauth.RedirectState
	if err := json.Unmarshal(raw, &p); err != nil {
		return domainoauth.RedirectState{}, err
	}
	return p, nil
}
