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
)

// RunNotificationWatermillRouter consumes notification domain events and writes in-app notifications.
func RunNotificationWatermillRouter(ctx context.Context, sub message.Subscriber, notif *appnotification.Service, wmLog watermill.LoggerAdapter) (*message.Router, error) {
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
			if err := dispatchNotificationEvent(context.Background(), notif, msg.Payload); err != nil {
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

func dispatchNotificationEvent(ctx context.Context, notif *appnotification.Service, payload []byte) error {
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
	default:
		return nil
	}
}
