package eventing

import "context"

// PostSensitiveScanPublisher publishes async scan requests for post bodies after publish.
type PostSensitiveScanPublisher interface {
	PublishPostSensitiveScan(ctx context.Context, postID, authorID int64, postUpdatedAtUnixNano int64, kind string) error
}

