package post

import "time"

type Post struct {
	ID               int64
	UserID           int64
	PostType         int
	ReplyToPostID    *int64
	ReferencedPostID *int64
	Visibility       int
	AudienceListID   *int64
	CategoryID       *int64
	Body             string
	Images           []string
	Status           int
	ModerationFlag   int
	ModerationNote   string
	AppealBody       string
	AppealStatus     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time // non-nil when soft-deleted (Unscoped loads)
}
