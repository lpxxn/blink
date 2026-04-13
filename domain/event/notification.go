package event

// 站内通知相关的领域事件类型（经 Watermill + Redis Stream 投递，由 notification 应用服务消费落库）。

const (
	NotificationReplyToPost    = "reply_to_post"
	NotificationReplyToComment = "reply_to_comment"
	NotificationPostRemoved    = "post_removed"
	NotificationPostFlagged    = "post_flagged"
	NotificationAppealResolved = "appeal_resolved"
)
