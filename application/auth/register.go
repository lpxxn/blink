package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

type RegisterService struct {
	Users      domainuser.Repository
	Identities domainoauth.Repository
	Node       *snowflake.Node
}

// RegisterWithPassword creates a local user and links builtin IdP identity (subject = snowflake id string).
func (s *RegisterService) RegisterWithPassword(ctx context.Context, email, password, name string) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return 0, ErrInvalidEmail
	}
	if len(password) < minPasswordLen {
		return 0, ErrWeakPassword
	}
	if _, err := s.Users.FindByEmail(ctx, email); err == nil {
		return 0, ErrEmailTaken
	} else if !errors.Is(err, domainuser.ErrNotFound) {
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	n := strings.TrimSpace(name)
	if n == "" {
		n = email
	}
	uid := s.Node.Generate().Int64()
	oauthID := s.Node.Generate().Int64()
	u := &domainuser.User{
		SnowflakeID:  uid,
		Email:        email,
		Name:         n,
		WechatID:     "",
		Phone:        "",
		PasswordHash: string(hash),
		PasswordSalt: "",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := s.Users.Create(ctx, u); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrEmailTaken
		}
		return 0, err
	}
	sub := FormatBuiltinSubject(uid)
	oid := &domainoauth.Identity{
		SnowflakeID:     oauthID,
		Provider:        "builtin",
		ProviderSubject: sub,
		UserID:          uid,
	}
	if err := s.Identities.Create(ctx, oid); err != nil {
		return 0, err
	}
	return uid, nil
}

// FormatBuiltinSubject is oauth_identities.provider_subject and OIDC userinfo "sub" for the builtin IdP.
func FormatBuiltinSubject(userID int64) string {
	return strconv.FormatInt(userID, 10)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
