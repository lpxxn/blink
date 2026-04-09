package postreply

import "time"

type Reply struct {
	ID             int64
	PostID         int64
	UserID         int64
	ParentReplyID  *int64
	Body           string
	Status         int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
