package auth

import "errors"

var (
	ErrEmailTaken       = errors.New("auth: email already registered")
	ErrWeakPassword     = errors.New("auth: password too short")
	ErrInvalidEmail     = errors.New("auth: invalid email")
)
