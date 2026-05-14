package feedback

import "time"

const (
	StatusOpen   = "open"
	StatusClosed = "closed"

	SenderUser  = "user"
	SenderAdmin = "admin"
)

type Thread struct {
	ID             int64
	UserID         int64
	Status         string
	UserReplyCount int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastMessageAt  time.Time
}

type Message struct {
	ID         int64
	FeedbackID int64
	SenderID   int64
	SenderType string
	Body       string
	CreatedAt  time.Time
}
