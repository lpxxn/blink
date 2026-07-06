package block

import "errors"

var (
	ErrNotFound    = errors.New("block: not found")
	ErrSelfBlock   = errors.New("block: cannot block yourself")
	ErrAlreadyBlocked = errors.New("block: already blocked")
)
