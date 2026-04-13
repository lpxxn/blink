package event

// 领域事件名称（写入 Redis Stream 的 envelope.name，供消费者路由）。
const (
	ReplyPosted            = "reply.posted"
	PostRemovedByModerator = "post.removed_by_moderator"
	PostFlaggedByModerator = "post.flagged_by_moderator"
	AppealResolved         = "post.appeal_resolved"
)

// ReplyPostedPayload 评论已创建（含楼中楼语义所需字段，由发布端组装）。
type ReplyPostedPayload struct {
	PostID              int64  `json:"post_id,string"`
	ReplyID             int64  `json:"reply_id,string"`
	ReplierUserID       int64  `json:"replier_user_id,string"`
	PostAuthorID        int64  `json:"post_author_id,string"`
	ParentReplyAuthorID *int64 `json:"parent_reply_author_id,string,omitempty"`
	BodySnippet         string `json:"body_snippet"`
}

// PostRemovedByModeratorPayload 帖子被管理员下架。
type PostRemovedByModeratorPayload struct {
	PostID   int64  `json:"post_id,string"`
	AuthorID int64  `json:"author_id,string"`
	Reason   string `json:"reason"`
}

// PostFlaggedByModeratorPayload 帖子被管理员标记违规。
type PostFlaggedByModeratorPayload struct {
	PostID   int64  `json:"post_id,string"`
	AuthorID int64  `json:"author_id,string"`
	Note     string `json:"note"`
}

// AppealResolvedPayload 申诉/复核已裁决。
type AppealResolvedPayload struct {
	PostID    int64  `json:"post_id,string"`
	AuthorID  int64  `json:"author_id,string"`
	Approved  bool   `json:"approved"`
	AdminNote string `json:"admin_note"`
}
