package postreply

import "context"

type Repository interface {
	Create(ctx context.Context, r *Reply) error
	GetByID(ctx context.Context, id int64) (*Reply, error)
	// ListByPostID returns replies oldest-first; when afterID is set, only rows with id > afterID.
	ListByPostID(ctx context.Context, postID int64, afterID *int64, limit int) ([]*Reply, error)
	SoftDelete(ctx context.Context, id int64) error
}
