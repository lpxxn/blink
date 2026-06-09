package follow

import "time"

// ListEntry is one user in a following/followers list.
type ListEntry struct {
	UserID    int64
	CreatedAt time.Time
}

// PageCursor continues list queries ordered by created_at desc, user id desc.
type PageCursor struct {
	CreatedAt time.Time
	UserID    int64
}
