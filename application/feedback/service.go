package feedback

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	appnotification "github.com/lpxxn/blink/application/notification"
	domainfeedback "github.com/lpxxn/blink/domain/feedback"
	domainuser "github.com/lpxxn/blink/domain/user"
)

const (
	MaxBodyLen       = 4000
	MaxUserFollowups = 2
)

var (
	ErrInvalidBody = errors.New("feedback: body is required")
	ErrForbidden   = errors.New("feedback: forbidden")
	ErrReplyLimit  = errors.New("feedback: user reply limit reached")
	ErrClosed      = errors.New("feedback: closed")
)

type Service struct {
	Repo          domainfeedback.Repository
	Users         domainuser.Repository
	Notifications *appnotification.Service
	NewID         func() int64
}

func normalizeBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ErrInvalidBody
	}
	if len(body) > MaxBodyLen {
		return "", ErrInvalidBody
	}
	return body, nil
}

func snippet(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 240 {
		return body[:240] + "..."
	}
	return body
}

func (s *Service) Create(ctx context.Context, userID int64, body string) (*domainfeedback.Thread, error) {
	body, err := normalizeBody(body)
	if err != nil {
		return nil, err
	}
	t := &domainfeedback.Thread{
		ID:     s.NewID(),
		UserID: userID,
		Status: domainfeedback.StatusOpen,
	}
	m := &domainfeedback.Message{
		ID:         s.NewID(),
		FeedbackID: t.ID,
		SenderID:   userID,
		SenderType: domainfeedback.SenderUser,
		Body:       body,
	}
	if err := s.Repo.CreateThreadWithMessage(ctx, t, m); err != nil {
		return nil, err
	}
	s.notifyAdmins(ctx, userID, t.ID, body)
	s.notifyUserSubmitted(ctx, userID, t.ID)
	return s.Repo.GetThreadByID(ctx, t.ID)
}

func (s *Service) ReplyByUser(ctx context.Context, userID, feedbackID int64, body string) (*domainfeedback.Message, error) {
	body, err := normalizeBody(body)
	if err != nil {
		return nil, err
	}
	t, err := s.Repo.GetThreadByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden
	}
	if t.Status == domainfeedback.StatusClosed {
		return nil, ErrClosed
	}
	if t.UserReplyCount >= MaxUserFollowups {
		return nil, ErrReplyLimit
	}
	m := &domainfeedback.Message{
		ID:         s.NewID(),
		FeedbackID: feedbackID,
		SenderID:   userID,
		SenderType: domainfeedback.SenderUser,
		Body:       body,
	}
	if err := s.Repo.AddMessageAndTouch(ctx, m, 1); err != nil {
		return nil, err
	}
	s.notifyAdmins(ctx, userID, feedbackID, body)
	return m, nil
}

func (s *Service) ReplyByAdmin(ctx context.Context, adminID, feedbackID int64, body string) (*domainfeedback.Message, error) {
	body, err := normalizeBody(body)
	if err != nil {
		return nil, err
	}
	t, err := s.Repo.GetThreadByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}
	if t.Status == domainfeedback.StatusClosed {
		return nil, ErrClosed
	}
	m := &domainfeedback.Message{
		ID:         s.NewID(),
		FeedbackID: feedbackID,
		SenderID:   adminID,
		SenderType: domainfeedback.SenderAdmin,
		Body:       body,
	}
	if err := s.Repo.AddMessageAndTouch(ctx, m, 0); err != nil {
		return nil, err
	}
	if s.Notifications != nil {
		if err := s.Notifications.OnFeedbackReply(ctx, t.UserID, feedbackID, body); err != nil {
			log.Printf("feedback: notify user %d for feedback %d: %v", t.UserID, feedbackID, err)
		}
	}
	return m, nil
}

func (s *Service) GetForUser(ctx context.Context, userID, feedbackID int64) (*domainfeedback.Thread, []*domainfeedback.Message, error) {
	t, err := s.Repo.GetThreadByID(ctx, feedbackID)
	if err != nil {
		return nil, nil, err
	}
	if t.UserID != userID {
		return nil, nil, ErrForbidden
	}
	msgs, err := s.Repo.ListMessages(ctx, feedbackID)
	return t, msgs, err
}

func (s *Service) GetForAdmin(ctx context.Context, feedbackID int64) (*domainfeedback.Thread, []*domainfeedback.Message, error) {
	t, err := s.Repo.GetThreadByID(ctx, feedbackID)
	if err != nil {
		return nil, nil, err
	}
	msgs, err := s.Repo.ListMessages(ctx, feedbackID)
	return t, msgs, err
}

func (s *Service) ListForUser(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*domainfeedback.Thread, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.Repo.ListByUserID(ctx, userID, beforeID, limit)
}

func (s *Service) ListForAdmin(ctx context.Context, offset, limit int) ([]*domainfeedback.Thread, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.Repo.ListPage(ctx, offset, limit)
}

func (s *Service) notifyAdmins(ctx context.Context, userID, feedbackID int64, body string) {
	if s.Notifications == nil || s.Users == nil {
		return
	}
	ids, err := s.adminRecipientIDs(ctx)
	if err != nil {
		log.Printf("feedback: list admin recipients: %v", err)
		return
	}
	title := "新的意见反馈"
	msg := "用户 " + strconv.FormatInt(userID, 10) + " 提交/补充了意见反馈 #" + strconv.FormatInt(feedbackID, 10) + "。\n内容摘要：" + snippet(body)
	msg += "\n请在管理后台「反馈」中回复。"
	for _, adminID := range ids {
		if adminID == userID {
			continue
		}
		if err := s.Notifications.OnFeedbackForAdmin(ctx, adminID, feedbackID, title, msg); err != nil {
			log.Printf("feedback: notify admin %d for feedback %d: %v", adminID, feedbackID, err)
		}
	}
}

func (s *Service) adminRecipientIDs(ctx context.Context) ([]int64, error) {
	superIDs, err := s.Users.ListSnowflakeIDsByRole(ctx, domainuser.RoleSuperAdmin)
	if err != nil {
		return nil, err
	}
	adminIDs, err := s.Users.ListSnowflakeIDsByRole(ctx, domainuser.RoleAdmin)
	if err != nil {
		return nil, err
	}
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(superIDs)+len(adminIDs))
	for _, id := range append(superIDs, adminIDs...) {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (s *Service) notifyUserSubmitted(ctx context.Context, userID, feedbackID int64) {
	if s.Notifications == nil {
		return
	}
	if err := s.Notifications.OnFeedbackSubmitted(ctx, userID, feedbackID); err != nil {
		log.Printf("feedback: notify submitter %d for feedback %d: %v", userID, feedbackID, err)
	}
}
