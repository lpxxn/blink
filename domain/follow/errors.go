package follow

import "errors"

var (
	ErrNotFound      = errors.New("follow: not found")
	ErrAlreadyExists = errors.New("follow: already following")
	ErrSelfFollow    = errors.New("follow: cannot follow yourself")
)
