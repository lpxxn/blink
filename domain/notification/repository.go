package notification

import "context"

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	ListByUserID(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*Notification, error)
	MarkRead(ctx context.Context, userID, id int64) error
	MarkAllRead(ctx context.Context, userID int64) error
	CountUnread(ctx context.Context, userID int64) (int64, error)
}
