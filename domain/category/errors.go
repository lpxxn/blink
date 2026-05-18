package category

import "errors"

var (
	ErrNotFound  = errors.New("category: not found")
	ErrDuplicate = errors.New("category: duplicate slug")
)
