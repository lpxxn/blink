package eventing

import "context"

// NotificationPublisher publishes domain notification events (e.g. to Redis via Watermill).
// Implementations must be non-blocking where possible; failures may be logged by callers.
type NotificationPublisher interface {
	PublishReplyToPost(ctx context.Context, postAuthorID, postID, replyID int64, snippet string) error
	PublishReplyToComment(ctx context.Context, parentAuthorID, postID, replyID int64, snippet string) error
	PublishPostRemoved(ctx context.Context, authorID, postID int64, reason string) error
	PublishAppealResolved(ctx context.Context, authorID, postID int64, approved bool, adminNote string) error
}
