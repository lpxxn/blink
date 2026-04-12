package notification

import "time"

// Types for notifications.type (VARCHAR).
const (
	TypeReply          = "reply"
	TypeReplyToComment = "reply_to_comment"
	TypePostRemoved    = "post_removed"
	TypeAppealResult = "appeal_result"
	TypeSystem       = "system"
)

type Notification struct {
	ID         int64
	UserID     int64
	Type       string
	Title      string
	Body       string
	RefPostID  *int64
	RefReplyID *int64
	ReadAt     *time.Time
	CreatedAt  time.Time
}
