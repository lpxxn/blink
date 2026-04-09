package category

import "context"

type Repository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id int64) (*Category, error)
	ListActive(ctx context.Context) ([]*Category, error)
	Count(ctx context.Context) (int64, error)
}
