package sensitiveword

import "context"

type Repository interface {
	Create(ctx context.Context, w *Word) error
	GetByID(ctx context.Context, id int64) (*Word, error)
	// ListEnabled returns enabled words for the in-memory matcher (unordered).
	ListEnabledWords(ctx context.Context) ([]string, error)
	ListPage(ctx context.Context, offset, limit int) ([]*Word, int64, error)
	UpdateEnabled(ctx context.Context, id int64, enabled bool) error
	Delete(ctx context.Context, id int64) error
}
