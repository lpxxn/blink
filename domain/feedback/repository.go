package feedback

import "context"

type ListFilters struct {
	Status *string
	UserID *int64
}

type Repository interface {
	CreateThreadWithMessage(ctx context.Context, thread *Thread, message *Message) error
	GetThreadByID(ctx context.Context, id int64) (*Thread, error)
	ListByUserID(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*Thread, error)
	ListPage(ctx context.Context, f ListFilters, offset, limit int) ([]*Thread, int64, error)
	Count(ctx context.Context, f ListFilters) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	ListMessages(ctx context.Context, feedbackID int64) ([]*Message, error)
	AddMessageAndTouch(ctx context.Context, message *Message, userReplyDelta int) error
}
