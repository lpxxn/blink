package messaging

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"

	"github.com/lpxxn/blink/application/eventing"
	domainevent "github.com/lpxxn/blink/domain/event"
)

// NotificationWatermillPublisher implements eventing.NotificationPublisher via Watermill + Redis Streams.
type NotificationWatermillPublisher struct {
	inner message.Publisher
}

func NewNotificationWatermillPublisher(pub message.Publisher) *NotificationWatermillPublisher {
	return &NotificationWatermillPublisher{inner: pub}
}

var _ eventing.NotificationPublisher = (*NotificationWatermillPublisher)(nil)

func (p *NotificationWatermillPublisher) publish(ctx context.Context, payload any) error {
	_ = ctx
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := message.NewMessage(uuid.NewString(), b)
	return p.inner.Publish(TopicNotificationEvents, msg)
}

func (p *NotificationWatermillPublisher) PublishReplyToPost(ctx context.Context, postAuthorID, postID, replyID int64, snippet string) error {
	return p.publish(ctx, struct {
		Type         string `json:"type"`
		PostAuthorID int64  `json:"post_author_id,string"`
		PostID       int64  `json:"post_id,string"`
		ReplyID      int64  `json:"reply_id,string"`
		Snippet      string `json:"snippet"`
	}{
		Type:         domainevent.NotificationReplyToPost,
		PostAuthorID: postAuthorID,
		PostID:       postID,
		ReplyID:      replyID,
		Snippet:      snippet,
	})
}

func (p *NotificationWatermillPublisher) PublishReplyToComment(ctx context.Context, parentAuthorID, postID, replyID int64, snippet string) error {
	return p.publish(ctx, struct {
		Type           string `json:"type"`
		ParentAuthorID int64  `json:"parent_author_id,string"`
		PostID         int64  `json:"post_id,string"`
		ReplyID        int64  `json:"reply_id,string"`
		Snippet        string `json:"snippet"`
	}{
		Type:           domainevent.NotificationReplyToComment,
		ParentAuthorID: parentAuthorID,
		PostID:         postID,
		ReplyID:        replyID,
		Snippet:        snippet,
	})
}

func (p *NotificationWatermillPublisher) PublishPostRemoved(ctx context.Context, authorID, postID int64, reason string) error {
	return p.publish(ctx, struct {
		Type     string `json:"type"`
		AuthorID int64  `json:"author_id,string"`
		PostID   int64  `json:"post_id,string"`
		Reason   string `json:"reason"`
	}{
		Type:     domainevent.NotificationPostRemoved,
		AuthorID: authorID,
		PostID:   postID,
		Reason:   reason,
	})
}

func (p *NotificationWatermillPublisher) PublishPostFlagged(ctx context.Context, authorID, postID int64, note string) error {
	return p.publish(ctx, struct {
		Type     string `json:"type"`
		AuthorID int64  `json:"author_id,string"`
		PostID   int64  `json:"post_id,string"`
		Note     string `json:"note"`
	}{
		Type:     domainevent.NotificationPostFlagged,
		AuthorID: authorID,
		PostID:   postID,
		Note:     note,
	})
}

func (p *NotificationWatermillPublisher) PublishAppealSubmitted(ctx context.Context, authorID, postID int64, kind, message string) error {
	return p.publish(ctx, struct {
		Type     string `json:"type"`
		AuthorID int64  `json:"author_id,string"`
		PostID   int64  `json:"post_id,string"`
		Kind     string `json:"kind"`
		Message  string `json:"message"`
	}{
		Type:     domainevent.NotificationAppealSubmitted,
		AuthorID: authorID,
		PostID:   postID,
		Kind:     kind,
		Message:  message,
	})
}

func (p *NotificationWatermillPublisher) PublishAppealResolved(ctx context.Context, authorID, postID int64, approved bool, adminNote string) error {
	return p.publish(ctx, struct {
		Type      string `json:"type"`
		AuthorID  int64  `json:"author_id,string"`
		PostID    int64  `json:"post_id,string"`
		Approved  bool   `json:"approved"`
		AdminNote string `json:"admin_note"`
	}{
		Type:      domainevent.NotificationAppealResolved,
		AuthorID:  authorID,
		PostID:    postID,
		Approved:  approved,
		AdminNote: adminNote,
	})
}

func (p *NotificationWatermillPublisher) PublishUserBanned(ctx context.Context, userID int64) error {
	return p.publish(ctx, struct {
		Type   string `json:"type"`
		UserID int64  `json:"user_id,string"`
	}{
		Type:   domainevent.UserBanned,
		UserID: userID,
	})
}
