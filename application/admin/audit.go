package admin

import (
	"context"
	"strings"

	domainadminaudit "github.com/lpxxn/blink/domain/adminaudit"
)

const (
	AuditUserStatus   = "user.status"
	AuditUserRole     = "user.role"
	AuditUserPassword = "user.password_reset"
	AuditPostModerate = "post.moderate"
	AuditPostAppeal   = "post.appeal_resolve"
	AuditReplyHide    = "reply.hide"
	AuditFeedbackClose = "feedback.close"
	AuditCategoryCreate = "category.create"
	AuditCategoryUpdate = "category.update"
	AuditCategoryDelete = "category.delete"
	AuditSettingChange  = "setting.change"
)

func (s *Service) logAudit(ctx context.Context, actorID int64, action, targetType string, targetID *int64, detail string) {
	if s.Audit == nil || s.NewID == nil {
		return
	}
	detail = strings.TrimSpace(detail)
	if len(detail) > 4000 {
		detail = detail[:4000]
	}
	_ = s.Audit.Create(ctx, &domainadminaudit.Entry{
		ID:         s.NewID(),
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
