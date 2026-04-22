package messaging

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"

	"github.com/lpxxn/blink/application/eventing"
)

// PostSensitiveScanWatermillPublisher publishes scan requests on TopicPostSensitiveScan.
type PostSensitiveScanWatermillPublisher struct {
	inner message.Publisher
}

func NewPostSensitiveScanWatermillPublisher(pub message.Publisher) *PostSensitiveScanWatermillPublisher {
	return &PostSensitiveScanWatermillPublisher{inner: pub}
}

var _ eventing.PostSensitiveScanPublisher = (*PostSensitiveScanWatermillPublisher)(nil)

func (p *PostSensitiveScanWatermillPublisher) PublishPostSensitiveScan(ctx context.Context, postID, authorID int64, postUpdatedAtUnixNano int64, kind string) error {
	_ = ctx
	payload := struct {
		PostID    int64  `json:"post_id,string"`
		AuthorID  int64  `json:"author_id,string"`
		UpdatedAt int64  `json:"post_updated_at_unix_nano"`
		Kind      string `json:"kind"`
		EventVer  int    `json:"event_ver"`
	}{PostID: postID, AuthorID: authorID, UpdatedAt: postUpdatedAtUnixNano, Kind: kind, EventVer: 1}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := message.NewMessage(uuid.NewString(), b)
	return p.inner.Publish(TopicPostSensitiveScan, msg)
}
