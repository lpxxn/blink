package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/domain/session"
	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	Users      domainuser.Repository
	Identities domainoauth.Repository
	Sessions   session.Store
	States     domainoauth.StateStore
	Providers  map[string]OAuth2Provider
	Node       *snowflake.Node
	StateTTL   time.Duration
	SessionTTL time.Duration
}

func (s *LoginService) provider(name string) (OAuth2Provider, error) {
	p, ok := s.Providers[strings.ToLower(strings.TrimSpace(name))]
	if !ok || p == nil {
		return nil, ErrUnknownProvider
	}
	return p, nil
}

// LoginRedirectURL returns the IdP authorization URL and the opaque state token.
func (s *LoginService) LoginRedirectURL(ctx context.Context, providerName, nextURL string) (authURL string, state string, err error) {
	p, err := s.provider(providerName)
	if err != nil {
		return "", "", err
	}
	state, err = newStateToken()
	if err != nil {
		return "", "", err
	}
	payload := domainoauth.RedirectState{
		Provider: strings.ToLower(strings.TrimSpace(providerName)),
		NextURL:  safeNextURL(nextURL),
	}
	if err := s.States.Save(ctx, state, payload, s.StateTTL); err != nil {
		return "", "", err
	}
	return p.AuthCodeURL(state), state, nil
}

// CompleteLogin exchanges the code, upserts user + oauth identity, and creates a Redis-backed session.
func (s *LoginService) CompleteLogin(ctx context.Context, providerName, code, state, ip, ua string) (sessionToken string, nextURL string, err error) {
	p, err := s.provider(providerName)
	if err != nil {
		return "", "", err
	}
	st, err := s.States.Consume(ctx, state)
	if err != nil {
		return "", "", err
	}
	if st.Provider != strings.ToLower(strings.TrimSpace(providerName)) {
		return "", "", domainoauth.ErrInvalidState
	}
	tok, err := p.Exchange(ctx, code)
	if err != nil {
		return "", "", err
	}
	info, err := p.UserInfo(ctx, tok)
	if err != nil {
		return "", "", err
	}
	subject := strings.TrimSpace(info.Subject)
	if subject == "" {
		return "", "", fmt.Errorf("oauth: empty subject from provider")
	}
	prov := strings.ToLower(strings.TrimSpace(providerName))

	ident, err := s.Identities.FindByProviderSubject(ctx, prov, subject)
	if err != nil && !errors.Is(err, domainoauth.ErrNotFound) {
		return "", "", err
	}

	var userID int64
	if ident != nil {
		userID = ident.UserID
	} else {
		userID, err = s.registerOAuthUser(ctx, prov, subject, info)
		if err != nil {
			return "", "", err
		}
	}

	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if u.Status != domainuser.StatusActive {
		return "", "", ErrUserSuspended
	}
	if err := s.Users.UpdateLastLogin(ctx, userID, ip, ua); err != nil {
		return "", "", err
	}
	sess, err := s.Sessions.Create(ctx, userID, s.SessionTTL, ip, ua)
	if err != nil {
		return "", "", err
	}
	return sess, st.NextURL, nil
}

func (s *LoginService) registerOAuthUser(ctx context.Context, provider, subject string, info UserInfo) (int64, error) {
	email := loginEmail(info, provider, subject)
	name := strings.TrimSpace(info.Name)
	if name == "" {
		name = email
	}
	pw, err := randomPasswordHash()
	if err != nil {
		return 0, err
	}
	uid := s.Node.Generate().Int64()
	oauthID := s.Node.Generate().Int64()
	u := &domainuser.User{
		SnowflakeID:  uid,
		Email:        email,
		Name:         name,
		WechatID:     "",
		Phone:        "",
		PasswordHash: pw,
		PasswordSalt: "",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := s.Users.Create(ctx, u); err != nil {
		return 0, err
	}
	oid := &domainoauth.Identity{
		SnowflakeID:     oauthID,
		Provider:        provider,
		ProviderSubject: subject,
		UserID:          uid,
	}
	if err := s.Identities.Create(ctx, oid); err != nil {
		return 0, err
	}
	return uid, nil
}

func loginEmail(info UserInfo, provider, subject string) string {
	if e := strings.TrimSpace(info.Email); e != "" {
		return strings.ToLower(e)
	}
	return fmt.Sprintf("oauth.%s.%s@oauth.local", provider, subject)
}

func randomPasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	h, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func newStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return base64.RawURLEncoding.EncodeToString(h[:]), nil
}

func safeNextURL(next string) string {
	n := strings.TrimSpace(next)
	if n == "" {
		return "/"
	}
	if !strings.HasPrefix(n, "/") || strings.HasPrefix(n, "//") {
		return "/"
	}
	return n
}
