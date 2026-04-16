package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appnotification "github.com/lpxxn/blink/application/notification"
	domainevent "github.com/lpxxn/blink/domain/event"
	domainsession "github.com/lpxxn/blink/domain/session"
)

// RunNotificationWatermillRouter consumes notification domain events and writes in-app notifications.
// sess may be nil; when set, user_banned events trigger idempotent session invalidation.
// reloadSensitiveWords, when non-nil, is invoked when another instance signals a sensitive-word list change.
func RunNotificationWatermillRouter(ctx context.Context, sub message.Subscriber, notif *appnotification.Service, sess domainsession.Store, reloadSensitiveWords func(context.Context) error, wmLog watermill.LoggerAdapter) (*message.Router, error) {
	if notif == nil {
		return nil, errors.New("messaging: notification service is nil")
	}
	router, err := message.NewRouter(message.RouterConfig{}, wmLog)
	if err != nil {
		return nil, err
	}

	router.AddConsumerHandler(
		"blink_notification_events",
		TopicNotificationEvents,
		sub,
		func(msg *message.Message) error {
			if err := dispatchNotificationEvent(context.Background(), notif, sess, msg.Payload); err != nil {
				log.Printf("notification event handler: %v", err)
				return err
			}
			return nil
		},
	)

	go func() {
		if err := router.Run(ctx); err != nil {
			log.Printf("watermill router stopped: %v", err)
		}
	}()

	<-router.Running()
	return router, nil
}

func dispatchNotificationEvent(ctx context.Context, notif *appnotification.Service, sess domainsession.Store, payload []byte) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &head); err != nil {
		return nil // malformed: ack to avoid poison-loop; logged upstream if needed
	}
	switch head.Type {
	case domainevent.NotificationReplyToPost:
		var body struct {
			PostAuthorID int64  `json:"post_author_id,string"`
			PostID       int64  `json:"post_id,string"`
			ReplyID      int64  `json:"reply_id,string"`
			Snippet      string `json:"snippet"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnNewReply(ctx, body.PostAuthorID, body.PostID, body.ReplyID, body.Snippet)
	case domainevent.NotificationReplyToComment:
		var body struct {
			ParentAuthorID int64  `json:"parent_author_id,string"`
			PostID         int64  `json:"post_id,string"`
			ReplyID        int64  `json:"reply_id,string"`
			Snippet        string `json:"snippet"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnReplyToYourComment(ctx, body.ParentAuthorID, body.PostID, body.ReplyID, body.Snippet)
	case domainevent.NotificationPostRemoved:
		var body struct {
			AuthorID int64  `json:"author_id,string"`
			PostID   int64  `json:"post_id,string"`
			Reason   string `json:"reason"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnPostRemoved(ctx, body.AuthorID, body.PostID, body.Reason)
	case domainevent.NotificationPostFlagged:
		var body struct {
			AuthorID int64  `json:"author_id,string"`
			PostID   int64  `json:"post_id,string"`
			Note     string `json:"note"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnPostFlagged(ctx, body.AuthorID, body.PostID, body.Note)
	case domainevent.NotificationSensitiveHit:
		var body struct {
			AuthorID int64    `json:"author_id,string"`
			PostID   int64    `json:"post_id,string"`
			Hits     []string `json:"hits"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnSensitiveHitForAdmins(ctx, body.AuthorID, body.PostID, body.Hits)
	case domainevent.NotificationAppealSubmitted:
		var body struct {
			AuthorID int64  `json:"author_id,string"`
			PostID   int64  `json:"post_id,string"`
			Kind     string `json:"kind"`
			Message  string `json:"message"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnAppealSubmittedForAdmins(ctx, body.AuthorID, body.PostID, body.Kind, body.Message)
	case domainevent.NotificationAppealResolved:
		var body struct {
			AuthorID  int64  `json:"author_id,string"`
			PostID    int64  `json:"post_id,string"`
			Approved  bool   `json:"approved"`
			AdminNote string `json:"admin_note"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		return notif.OnAppealResolved(ctx, body.AuthorID, body.PostID, body.Approved, body.AdminNote)
	case domainevent.UserBanned:
		var body struct {
			UserID int64 `json:"user_id,string"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil
		}
		if sess == nil {
			return nil
		}
		return sess.DeleteAllForUser(ctx, body.UserID)
	default:
		return nil
	}
}
