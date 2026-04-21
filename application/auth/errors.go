package auth

import "errors"

var (
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrWeakPassword       = errors.New("auth: password too short")
	ErrInvalidEmail       = errors.New("auth: invalid email")
	ErrInvalidCode        = errors.New("auth: invalid or expired verification code")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrUserInactive       = errors.New("auth: user inactive")
	ErrCodesNotConfigured = errors.New("auth: email code verifier not configured")
)
