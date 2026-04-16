package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appadmin "github.com/lpxxn/blink/application/admin"
	appmoderation "github.com/lpxxn/blink/application/moderation"
	domainpost "github.com/lpxxn/blink/domain/post"
)

// RunPostSensitiveScanWatermillRouter consumes TopicPostSensitiveScan and applies moderation actions.
func RunPostSensitiveScanWatermillRouter(ctx context.Context, sub message.Subscriber, admin *appadmin.Service, wmLog watermill.LoggerAdapter) (*message.Router, error) {
	if admin == nil || admin.Posts == nil {
		return nil, errors.New("messaging: admin service not configured")
	}
	router, err := message.NewRouter(message.RouterConfig{}, wmLog)
	if err != nil {
		return nil, err
	}
	router.AddConsumerHandler(
		"blink_post_sensitive_scan",
		TopicPostSensitiveScan,
		sub,
		func(msg *message.Message) error {
			if err := handlePostSensitiveScan(context.Background(), admin, msg.Payload); err != nil {
				log.Printf("post sensitive scan handler: %v", err)
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

func handlePostSensitiveScan(ctx context.Context, admin *appadmin.Service, payload []byte) error {
	var body struct {
		PostID    int64  `json:"post_id,string"`
		AuthorID  int64  `json:"author_id,string"`
		UpdatedAt int64  `json:"post_updated_at_unix_nano"`
		Kind      string `json:"kind"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil // ack malformed
	}
	if body.PostID == 0 {
		return nil
	}

	p, err := admin.Posts.GetByID(ctx, body.PostID)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			return nil
		}
		return err
	}
	if p == nil || p.DeletedAt != nil {
		return nil
	}
	// Published, not admin-removed (scan normal posts and re-scan violation-flagged posts).
	if p.Status != domainpost.StatusPublished || p.ModerationFlag == domainpost.ModerationRemoved {
		return nil
	}

	// Skip stale events (user edited/re-published after this event).
	if body.UpdatedAt != 0 && p.UpdatedAt.UTC().UnixNano() != body.UpdatedAt {
		return nil
	}

	words := appmoderation.SensitiveWords()
	hits := appmoderation.FindSensitiveHits(p.Body, words)
	if len(hits) == 0 {
		// Do not auto-clear ModerationFlagged; authors use moderation_request + admin resolve.
		return nil
	}
	if admin.NotifyEvents != nil {
		_ = admin.NotifyEvents.PublishSensitiveHitForAdmins(ctx, p.UserID, p.ID, hits)
	}

	mode := appadmin.SensitivePostModeReview
	if m, err := admin.GetSensitivePostMode(ctx); err == nil && strings.TrimSpace(m) != "" {
		mode = m
	}
	note := appmoderation.ModerationNoteForSensitiveHits(hits)
	switch mode {
	case appadmin.SensitivePostModeAutoRemove:
		flag := domainpost.ModerationRemoved
		_, err := admin.PatchPost(ctx, p.ID, &flag, &note, nil)
		return err
	default:
		flag := domainpost.ModerationFlagged
		_, err := admin.PatchPost(ctx, p.ID, &flag, &note, nil)
		return err
	}
}

