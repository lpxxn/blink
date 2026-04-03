package oauth

import "errors"

var (
	ErrUnknownProvider = errors.New("oauth: unknown provider")
	ErrUserSuspended   = errors.New("oauth: user inactive or banned")
)
