package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// OAuth2Provider abstracts a single OAuth2 authorization server + userinfo.
type OAuth2Provider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	UserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error)
}

// UserInfo is normalized from a provider userinfo response.
type UserInfo struct {
	Subject string
	Email   string
	Name    string
}
