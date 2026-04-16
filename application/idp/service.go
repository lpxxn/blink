package idp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	appauth "github.com/lpxxn/blink/application/auth"
	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRequest = errors.New("idp: invalid request")
	ErrInvalidClient  = errors.New("idp: invalid client")
	ErrInvalidGrant   = errors.New("idp: invalid grant")
	ErrUnauthorized   = errors.New("idp: unauthorized")
)

// Service is a minimal OAuth2 authorization server (authorization code + bearer userinfo).
type Service struct {
	ClientID        string
	ClientSecret    string
	AllowedRedirect map[string]struct{}
	Users           domainuser.Repository
	Identities      domainoauth.Repository
	Codes           AuthCodeStore
	Access          AccessTokenStore
	Node            *snowflake.Node
	CodeTTL         time.Duration
	AccessTTL       time.Duration
}

func (s *Service) ValidateAuthorizeQuery(clientID, redirectURI, responseType string) error {
	if responseType != "code" {
		return ErrInvalidRequest
	}
	if clientID != s.ClientID {
		return ErrInvalidClient
	}
	if redirectURI == "" {
		return ErrInvalidRequest
	}
	if _, ok := s.AllowedRedirect[redirectURI]; !ok {
		return ErrInvalidRequest
	}
	return nil
}

// LoginWithPassword validates the OAuth authorize request and issues an authorization code.
func (s *Service) LoginWithPassword(ctx context.Context, clientID, redirectURI, state, email, password string) (redirectURL string, err error) {
	if err := s.ValidateAuthorizeQuery(clientID, redirectURI, "code"); err != nil {
		return "", err
	}
	if state == "" {
		return "", ErrInvalidRequest
	}
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.Users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return "", ErrUnauthorized
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrUnauthorized
	}
	if u.Status != domainuser.StatusActive {
		return "", ErrUnauthorized
	}
	if err := s.ensureBuiltinIdentity(ctx, u); err != nil {
		return "", err
	}
	code, err := randomHex(16)
	if err != nil {
		return "", err
	}
	if err := s.Codes.SaveAuthCode(ctx, code, AuthCodePayload{
		UserID:      u.SnowflakeID,
		ClientID:    clientID,
		RedirectURI: redirectURI,
	}, s.CodeTTL); err != nil {
		return "", err
	}
	ru, err := url.Parse(redirectURI)
	if err != nil {
		return "", ErrInvalidRequest
	}
	q := ru.Query()
	q.Set("code", code)
	q.Set("state", state)
	ru.RawQuery = q.Encode()
	return ru.String(), nil
}

func (s *Service) ensureBuiltinIdentity(ctx context.Context, u *domainuser.User) error {
	sub := appauth.FormatBuiltinSubject(u.SnowflakeID)
	_, err := s.Identities.FindByProviderSubject(ctx, "builtin", sub)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domainoauth.ErrNotFound) {
		return err
	}
	if s.Node == nil {
		return errors.New("idp: snowflake node required to link builtin identity")
	}
	oid := &domainoauth.Identity{
		SnowflakeID:     s.Node.Generate().Int64(),
		Provider:        "builtin",
		ProviderSubject: sub,
		UserID:          u.SnowflakeID,
	}
	return s.Identities.Create(ctx, oid)
}

// Exchange turns an authorization code into an access token (RFC 6749).
func (s *Service) Exchange(ctx context.Context, grantType, code, redirectURI, clientID, clientSecret string) (accessToken string, expiresIn int, err error) {
	if grantType != "authorization_code" {
		return "", 0, ErrInvalidRequest
	}
	if clientID != s.ClientID || clientSecret != s.ClientSecret {
		return "", 0, ErrInvalidClient
	}
	p, err := s.Codes.ConsumeAuthCode(ctx, code)
	if err != nil {
		return "", 0, err
	}
	if p.ClientID != clientID || p.RedirectURI != redirectURI {
		return "", 0, ErrInvalidGrant
	}
	tok, err := randomHex(32)
	if err != nil {
		return "", 0, err
	}
	if err := s.Access.SaveAccessToken(ctx, tok, p.UserID, s.AccessTTL); err != nil {
		return "", 0, err
	}
	return tok, int(s.AccessTTL.Seconds()), nil
}

// UserInfo returns OIDC-style claims for the access token.
func (s *Service) UserInfo(ctx context.Context, accessToken string) (sub, email, name string, err error) {
	uid, err := s.Access.GetUserIDByAccessToken(ctx, accessToken)
	if err != nil {
		return "", "", "", err
	}
	u, err := s.Users.GetByID(ctx, uid)
	if err != nil {
		return "", "", "", err
	}
	return appauth.FormatBuiltinSubject(u.SnowflakeID), u.Email, u.Name, nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
