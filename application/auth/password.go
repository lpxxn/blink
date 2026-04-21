package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// Verification-code purpose strings. Kept in sync with
// application/emailcode.Purpose* constants; duplicated here to break the
// (emailcode -> redisstore -> idp -> auth) package import cycle in tests.
const (
	purposeRegister       = "register"
	purposeChangePassword = "change_password"
	purposeResetPassword  = "reset_password"
)

// CodeSendVerifier is the subset of emailcode.Service required by PasswordService
// (send + verify). Defined here to keep the application-auth package free of
// any circular imports with application/emailcode (it only imports constants).
type CodeSendVerifier interface {
	Send(ctx context.Context, purpose, email, actor string) error
	Verify(ctx context.Context, purpose, email, code string) error
}

// PasswordService handles the two password-flow use cases:
//
//  1. logged-in "change password" (code is sent to the current account email)
//  2. logged-out "forgot password / reset" (code is sent to the supplied email)
//
// On successful change or reset all sessions for the user are invalidated so
// every device is forced to re-authenticate.
type PasswordService struct {
	Users    domainuser.Repository
	Sessions session.Store
	Codes    CodeSendVerifier
}

// SendChangeCode sends a change-password verification code to the user's current
// email. The email is looked up server-side to prevent the caller from sending
// codes to arbitrary addresses.
func (s *PasswordService) SendChangeCode(ctx context.Context, userID int64) error {
	if s.Users == nil || s.Codes == nil {
		return ErrCodesNotConfigured
	}
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	actor := userActorKey(userID)
	return s.Codes.Send(ctx, purposeChangePassword, u.Email, actor)
}

// ChangePassword validates the code, sets a new bcrypt hash and drops all of
// the user's sessions.
func (s *PasswordService) ChangePassword(ctx context.Context, userID int64, code, newPassword string) error {
	if s.Users == nil || s.Codes == nil {
		return ErrCodesNotConfigured
	}
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.Codes.Verify(ctx, purposeChangePassword, u.Email, code); err != nil {
		return ErrInvalidCode
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.Users.UpdatePasswordHash(ctx, u.SnowflakeID, string(hash)); err != nil {
		return err
	}
	if s.Sessions != nil {
		_ = s.Sessions.DeleteAllForUser(ctx, u.SnowflakeID)
	}
	return nil
}

// SendResetCode emits a reset-password code for the supplied email. When the
// email is not registered the method still returns nil (and does not send)
// so the endpoint can't be used to enumerate accounts. actor is typically an
// IP-derived key.
func (s *PasswordService) SendResetCode(ctx context.Context, email, actor string) error {
	if s.Users == nil || s.Codes == nil {
		return ErrCodesNotConfigured
	}
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" || !strings.Contains(e, "@") {
		return ErrInvalidEmail
	}
	if _, err := s.Users.FindByEmail(ctx, e); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.Codes.Send(ctx, purposeResetPassword, e, actor)
}

// ResetPassword validates the code for email and overwrites the stored hash.
// All of the affected user's sessions are dropped afterwards.
func (s *PasswordService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if s.Users == nil || s.Codes == nil {
		return ErrCodesNotConfigured
	}
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" || !strings.Contains(e, "@") {
		return ErrInvalidEmail
	}
	u, err := s.Users.FindByEmail(ctx, e)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return ErrInvalidCode
		}
		return err
	}
	if err := s.Codes.Verify(ctx, purposeResetPassword, e, code); err != nil {
		return ErrInvalidCode
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.Users.UpdatePasswordHash(ctx, u.SnowflakeID, string(hash)); err != nil {
		return err
	}
	if s.Sessions != nil {
		_ = s.Sessions.DeleteAllForUser(ctx, u.SnowflakeID)
	}
	return nil
}

func userActorKey(userID int64) string {
	return "uid:" + strconv.FormatInt(userID, 10)
}
