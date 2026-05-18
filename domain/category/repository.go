package category

import "context"

type Repository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id int64) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	ListActive(ctx context.Context) ([]*Category, error)
	ListAll(ctx context.Context) ([]*Category, error)
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}
