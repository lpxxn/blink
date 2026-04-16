package postreply

import "context"

type Repository interface {
	Create(ctx context.Context, r *Reply) error
	GetByID(ctx context.Context, id int64) (*Reply, error)
	// ListByPostID returns replies oldest-first; when afterID is set, only rows with id > afterID.
	ListByPostID(ctx context.Context, postID int64, afterID *int64, limit int) ([]*Reply, error)
	// ListByPostIDAllStatuses is like ListByPostID but includes hidden replies (still excludes soft-deleted rows).
	ListByPostIDAllStatuses(ctx context.Context, postID int64, afterID *int64, limit int) ([]*Reply, error)
	SoftDelete(ctx context.Context, id int64) error
	// HideSubtree sets status hidden for rootID and all descendants (non-deleted rows).
	HideSubtree(ctx context.Context, rootID int64) error
	// UnhideSubtree sets status visible for rootID and all descendants (non-deleted rows).
	UnhideSubtree(ctx context.Context, rootID int64) error
}
