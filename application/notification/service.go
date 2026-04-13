package notification

import (
	"context"
	"strconv"
	"strings"

	domainnotification "github.com/lpxxn/blink/domain/notification"
	domainuser "github.com/lpxxn/blink/domain/user"
)

const (
	maxTitle = 500
	maxBody  = 8000
)

type Service struct {
	Repo  domainnotification.Repository
	NewID func() int64
	// Users is optional; when set, used to notify super admins (e.g. new appeal).
	Users domainuser.Repository
}

func (s *Service) send(ctx context.Context, userID int64, typ, title, body string, postID, replyID *int64) error {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		title = "通知"
	}
	if len(title) > maxTitle {
		title = title[:maxTitle]
	}
	if len(body) > maxBody {
		body = body[:maxBody]
	}
	n := &domainnotification.Notification{
		ID:         s.NewID(),
		UserID:     userID,
		Type:       typ,
		Title:      title,
		Body:       body,
		RefPostID:  postID,
		RefReplyID: replyID,
	}
	return s.Repo.Create(ctx, n)
}

func (s *Service) OnNewReply(ctx context.Context, postAuthorID, postID, replyID int64, replySnippet string) error {
	if postAuthorID == 0 {
		return nil
	}
	snip := strings.TrimSpace(replySnippet)
	if len(snip) > 200 {
		snip = snip[:200] + "…"
	}
	pid := postID
	rid := replyID
	return s.send(ctx, postAuthorID, domainnotification.TypeReply, "你的帖子有新评论",
		"内容摘要："+snip, &pid, &rid)
}

// OnReplyToYourComment notifies the author of the parent comment when someone replies under their comment.
func (s *Service) OnReplyToYourComment(ctx context.Context, parentAuthorID, postID, newReplyID int64, replySnippet string) error {
	if parentAuthorID == 0 {
		return nil
	}
	snip := strings.TrimSpace(replySnippet)
	if len(snip) > 200 {
		snip = snip[:200] + "…"
	}
	pid := postID
	rid := newReplyID
	return s.send(ctx, parentAuthorID, domainnotification.TypeReplyToComment, "有人回复了你的评论",
		"内容摘要："+snip, &pid, &rid)
}

func (s *Service) OnPostRemoved(ctx context.Context, authorID, postID int64, reason string) error {
	pid := postID
	body := "你的帖子已被管理员下架。"
	if strings.TrimSpace(reason) != "" {
		body += "\n原因：" + strings.TrimSpace(reason)
	}
	body += "\n你可修改正文后「申请复核」，或提交申诉说明。"
	return s.send(ctx, authorID, domainnotification.TypePostRemoved, "帖子已下架", body, &pid, nil)
}

func (s *Service) OnPostFlagged(ctx context.Context, authorID, postID int64, note string) error {
	pid := postID
	body := "管理员已将你的帖子标记为违规，公开流中将不再展示该帖。"
	if strings.TrimSpace(note) != "" {
		body += "\n说明：" + strings.TrimSpace(note)
	}
	return s.send(ctx, authorID, domainnotification.TypePostFlagged, "帖子被标记违规", body, &pid, nil)
}

// OnAppealSubmittedForAdmins notifies every super_admin (except the author) that a moderation request was filed.
func (s *Service) OnAppealSubmittedForAdmins(ctx context.Context, authorID, postID int64, kind, message string) error {
	if s.Users == nil {
		return nil
	}
	ids, err := s.Users.ListSnowflakeIDsByRole(ctx, domainuser.RoleSuperAdmin)
	if err != nil {
		return err
	}
	kind = strings.TrimSpace(strings.ToLower(kind))
	kindLabel := "复核申请"
	if kind == "appeal" {
		kindLabel = "申诉"
	}
	msg := strings.TrimSpace(message)
	if len(msg) > 500 {
		msg = msg[:500] + "…"
	}
	body := "用户 " + strconv.FormatInt(authorID, 10) + " 对帖子 " + strconv.FormatInt(postID, 10) + " 提交了「" + kindLabel + "」。"
	if msg != "" {
		body += "\n说明：" + msg
	}
	body += "\n请在管理后台「待处理申诉/复核」中处理。"
	pid := postID
	title := "待处理申诉/复核"
	for _, uid := range ids {
		if uid == authorID {
			continue
		}
		if err := s.send(ctx, uid, domainnotification.TypeAppealSubmittedAdmin, title, body, &pid, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) OnAppealResolved(ctx context.Context, authorID, postID int64, approved bool, adminNote string) error {
	pid := postID
	title := "申诉/复核结果"
	body := ""
	if approved {
		body = "管理员已通过，帖子已恢复展示。"
	} else {
		body = "管理员驳回了你的申诉或复核申请。"
	}
	if strings.TrimSpace(adminNote) != "" {
		body += "\n说明：" + strings.TrimSpace(adminNote)
	}
	return s.send(ctx, authorID, domainnotification.TypeAppealResult, title, body, &pid, nil)
}

func (s *Service) List(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*domainnotification.Notification, error) {
	return s.Repo.ListByUserID(ctx, userID, beforeID, limit)
}

func (s *Service) MarkRead(ctx context.Context, userID, id int64) error {
	return s.Repo.MarkRead(ctx, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) error {
	return s.Repo.MarkAllRead(ctx, userID)
}

func (s *Service) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	return s.Repo.CountUnread(ctx, userID)
}
