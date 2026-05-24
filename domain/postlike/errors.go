package postlike

import "errors"

var (
	ErrNotFound     = errors.New("postlike: not found")
	ErrAlreadyLiked = errors.New("postlike: already liked")
)
