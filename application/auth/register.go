package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	domainoauth "github.com/lpxxn/blink/domain/oauth"
	"github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

// CodeVerifier is the minimal subset of emailcode.Service needed to check a
// one-time verification code. Keeping it here avoids an application->application
// import cycle: register.go does not care how the code is stored, only that it
// can be validated for a given purpose+email.
type CodeVerifier interface {
	Verify(ctx context.Context, purpose, email, code string) error
}

type RegisterService struct {
	Users      domainuser.Repository
	Identities domainoauth.Repository
	Node       *snowflake.Node
	// Tx 非 nil 时 RegisterWithPassword 在单事务内写入 user + oauth；nil 时直接用 Users/Identities（测试或简单场景）。
	Tx TxRunner
	// Sessions + SessionTTL 非 nil 时 RegisterWithSession 可在注册成功后签发会话（通常在事务提交之后写 Redis）。
	Sessions   session.Store
	SessionTTL time.Duration
	// Codes 非 nil 时，*Verified 方法会在建号前校验邮箱验证码 (purpose=emailcode.PurposeRegister)。
	Codes CodeVerifier
}

var ErrSessionNotConfigured = errors.New("auth: session store or tx not configured for register-with-session")

// RegisterWithPassword creates a local user and links builtin IdP identity (subject = snowflake id string).
func (s *RegisterService) RegisterWithPassword(ctx context.Context, email, password, name string) (int64, error) {
	if s.Tx == nil {
		return s.registerOnce(ctx, s.Users, s.Identities, email, password, name)
	}
	var uid int64
	err := s.Tx.Run(ctx, func(ctx context.Context, u domainuser.Repository, o domainoauth.Repository) error {
		var e error
		uid, e = s.registerOnce(ctx, u, o, email, password, name)
		return e
	})
	return uid, err
}

// RegisterWithSessionVerified first validates an email verification code, then
// runs RegisterWithSession. Returns ErrCodesNotConfigured if s.Codes is nil and
// ErrInvalidCode if the code check fails.
func (s *RegisterService) RegisterWithSessionVerified(ctx context.Context, email, password, name, code, ip, ua string) (userID int64, sessionToken string, err error) {
	if s.Codes == nil {
		return 0, "", ErrCodesNotConfigured
	}
	if err := s.Codes.Verify(ctx, purposeRegister, email, code); err != nil {
		return 0, "", ErrInvalidCode
	}
	return s.RegisterWithSession(ctx, email, password, name, ip, ua)
}

// RegisterWithPasswordVerified verifies the email code and then creates the
// user WITHOUT issuing a session. Used as a fail-closed fallback when the
// session store is not configured but codes are — we still refuse to create
// an account unless the code is valid.
func (s *RegisterService) RegisterWithPasswordVerified(ctx context.Context, email, password, name, code string) (int64, error) {
	if s.Codes == nil {
		return 0, ErrCodesNotConfigured
	}
	if err := s.Codes.Verify(ctx, purposeRegister, email, code); err != nil {
		return 0, ErrInvalidCode
	}
	return s.RegisterWithPassword(ctx, email, password, name)
}

// RegisterWithSession runs registration in one DB transaction, then creates a Redis-backed session so the client
// does not need a separate login. Requires Tx, Sessions, and SessionTTL.
func (s *RegisterService) RegisterWithSession(ctx context.Context, email, password, name, ip, ua string) (userID int64, sessionToken string, err error) {
	if s.Tx == nil || s.Sessions == nil || s.SessionTTL <= 0 {
		return 0, "", ErrSessionNotConfigured
	}
	var uid int64
	err = s.Tx.Run(ctx, func(ctx context.Context, u domainuser.Repository, o domainoauth.Repository) error {
		var e error
		uid, e = s.registerOnce(ctx, u, o, email, password, name)
		return e
	})
	if err != nil {
		return 0, "", err
	}
	tok, err := s.Sessions.Create(ctx, uid, s.SessionTTL, ip, ua)
	if err != nil {
		return 0, "", err
	}
	return uid, tok, nil
}

func (s *RegisterService) registerOnce(ctx context.Context, users domainuser.Repository, identities domainoauth.Repository, email, password, name string) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return 0, ErrInvalidEmail
	}
	if len(password) < minPasswordLen {
		return 0, ErrWeakPassword
	}
	if _, err := users.FindByEmail(ctx, email); err == nil {
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
	if err := users.Create(ctx, u); err != nil {
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
	if err := identities.Create(ctx, oid); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrEmailTaken
		}
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
