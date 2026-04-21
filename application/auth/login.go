package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// LoginLockout is the optional Redis-backed counter used to rate-limit brute-
// force login attempts. The Redis implementation lives under
// infrastructure/cache/redisstore.
type LoginLockout interface {
	// Record increments the failure counter for email. threshold/window define
	// when a lockout activates; lockout sets how long the key stays blocked.
	Record(ctx context.Context, email string, threshold int64, window, lockout time.Duration) (locked bool, err error)
	// IsLocked returns true if email is currently locked out.
	IsLocked(ctx context.Context, email string) (bool, error)
	// Reset clears counters (after a successful login).
	Reset(ctx context.Context, email string) error
}

// LoginService authenticates users against local password hashes and issues
// a session token. OAuth-based login paths stay in application/oauth.
type LoginService struct {
	Users      domainuser.Repository
	Sessions   session.Store
	SessionTTL time.Duration
	// Lockout is optional. When set, after LockoutThreshold consecutive failed
	// attempts within LockoutWindow the email is blocked for LockoutDuration.
	Lockout          LoginLockout
	LockoutThreshold int64
	LockoutWindow    time.Duration
	LockoutDuration  time.Duration
}

func (s *LoginService) lockThreshold() int64 {
	if s.LockoutThreshold > 0 {
		return s.LockoutThreshold
	}
	return 5
}
func (s *LoginService) lockWindow() time.Duration {
	if s.LockoutWindow > 0 {
		return s.LockoutWindow
	}
	return 5 * time.Minute
}
func (s *LoginService) lockDuration() time.Duration {
	if s.LockoutDuration > 0 {
		return s.LockoutDuration
	}
	return 5 * time.Minute
}

// LoginWithPassword authenticates and creates a session. All error paths map
// to ErrInvalidCredentials except infrastructure errors (which bubble up) to
// avoid leaking which of (email, password, status) was wrong.
func (s *LoginService) LoginWithPassword(ctx context.Context, email, password, ip, ua string) (userID int64, token string, err error) {
	if s.Users == nil || s.Sessions == nil || s.SessionTTL <= 0 {
		return 0, "", ErrSessionNotConfigured
	}
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" || !strings.Contains(e, "@") {
		return 0, "", ErrInvalidCredentials
	}
	if s.Lockout != nil {
		locked, err := s.Lockout.IsLocked(ctx, e)
		if err != nil {
			return 0, "", err
		}
		if locked {
			return 0, "", ErrInvalidCredentials
		}
	}
	u, err := s.Users.FindByEmail(ctx, e)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			s.recordFailure(ctx, e)
			return 0, "", ErrInvalidCredentials
		}
		return 0, "", err
	}
	if u.PasswordHash == "" {
		s.recordFailure(ctx, e)
		return 0, "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		s.recordFailure(ctx, e)
		return 0, "", ErrInvalidCredentials
	}
	if u.Status != domainuser.StatusActive {
		s.recordFailure(ctx, e)
		return 0, "", ErrInvalidCredentials
	}
	if err := s.Users.UpdateLastLogin(ctx, u.SnowflakeID, ip, ua); err != nil {
		return 0, "", err
	}
	tok, err := s.Sessions.Create(ctx, u.SnowflakeID, s.SessionTTL, ip, ua)
	if err != nil {
		return 0, "", err
	}
	if s.Lockout != nil {
		_ = s.Lockout.Reset(ctx, e)
	}
	return u.SnowflakeID, tok, nil
}

func (s *LoginService) recordFailure(ctx context.Context, email string) {
	if s.Lockout == nil {
		return
	}
	_, _ = s.Lockout.Record(ctx, email, s.lockThreshold(), s.lockWindow(), s.lockDuration())
}
