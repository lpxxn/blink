package oauth

import "errors"

var (
	ErrNotFound     = errors.New("oauth: identity not found")
	ErrInvalidState = errors.New("oauth: invalid or expired state")
)
