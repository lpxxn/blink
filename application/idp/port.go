package idp

import (
	"context"
	"time"
)

// AuthCodePayload is carried by the OAuth2 authorization code (RFC 6749).
type AuthCodePayload struct {
	UserID      int64  `json:"user_id"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
}

// AuthCodeStore persists one-time authorization codes.
type AuthCodeStore interface {
	SaveAuthCode(ctx context.Context, code string, p AuthCodePayload, ttl time.Duration) error
	ConsumeAuthCode(ctx context.Context, code string) (AuthCodePayload, error)
}

// AccessTokenStore persists bearer access tokens for userinfo.
type AccessTokenStore interface {
	SaveAccessToken(ctx context.Context, token string, userID int64, ttl time.Duration) error
	GetUserIDByAccessToken(ctx context.Context, token string) (int64, error)
}
