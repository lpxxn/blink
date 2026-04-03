package redisstore

import (
	"context"
	"encoding/json"
	"time"

	appidp "github.com/lpxxn/blink/application/idp"
	"github.com/redis/go-redis/v9"
)

const (
	idpCodePrefix   = "blink:idp:code:"
	idpAccessPrefix = "blink:idp:access:"
)

// IdPTokenStore implements application/idp AuthCodeStore and AccessTokenStore.
type IdPTokenStore struct {
	Client *redis.Client
}

func (s *IdPTokenStore) SaveAuthCode(ctx context.Context, code string, p appidp.AuthCodePayload, ttl time.Duration) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.Client.Set(ctx, idpCodePrefix+code, b, ttl).Err()
}

func (s *IdPTokenStore) ConsumeAuthCode(ctx context.Context, code string) (appidp.AuthCodePayload, error) {
	key := idpCodePrefix + code
	raw, err := s.Client.GetDel(ctx, key).Bytes()
	if err == redis.Nil {
		return appidp.AuthCodePayload{}, appidp.ErrInvalidGrant
	}
	if err != nil {
		return appidp.AuthCodePayload{}, err
	}
	var p appidp.AuthCodePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return appidp.AuthCodePayload{}, err
	}
	return p, nil
}

func (s *IdPTokenStore) SaveAccessToken(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	b, err := json.Marshal(struct {
		UserID int64 `json:"user_id"`
	}{UserID: userID})
	if err != nil {
		return err
	}
	return s.Client.Set(ctx, idpAccessPrefix+token, b, ttl).Err()
}

func (s *IdPTokenStore) GetUserIDByAccessToken(ctx context.Context, token string) (int64, error) {
	raw, err := s.Client.Get(ctx, idpAccessPrefix+token).Bytes()
	if err == redis.Nil {
		return 0, appidp.ErrUnauthorized
	}
	if err != nil {
		return 0, err
	}
	var p struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return 0, err
	}
	return p.UserID, nil
}
