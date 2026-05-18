package adminaudit

import "context"

type Repository interface {
	Create(ctx context.Context, e *Entry) error
	List(ctx context.Context, offset, limit int) ([]*Entry, error)
	Count(ctx context.Context) (int64, error)
}
