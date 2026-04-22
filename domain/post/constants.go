package post

const (
	TypeOriginal = 0

	StatusDraft     = 0
	StatusPublished = 1
	StatusHidden    = 2

	VisibilityPublic = 0

	ModerationNormal  = 0
	ModerationFlagged = 1
	ModerationRemoved = 2

	// AppealStatus — 管理员下架后作者的申诉 / 复核申请
	AppealNone     = 0
	AppealPending  = 1
	AppealRejected = 3
)
