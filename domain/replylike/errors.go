package replylike

import "errors"

var (
	ErrNotFound     = errors.New("replylike: not found")
	ErrAlreadyLiked = errors.New("replylike: already liked")
)
