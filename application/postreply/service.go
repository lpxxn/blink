package postreply

import (
	"context"
	"errors"
	"strings"

	appmoderation "github.com/lpxxn/blink/application/moderation"
	apppost "github.com/lpxxn/blink/application/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
)

var (
	ErrForbidden    = errors.New("postreply: forbidden")
	ErrInvalidInput = errors.New("postreply: invalid input")
)

const maxReplyBody = 8000

type Service struct {
	Posts   *apppost.Service
	Replies domainpostreply.Repository
	NewID   func() int64
}

func (s *Service) List(ctx context.Context, postID int64, afterID *int64, limit int, viewerID *int64, viewerIsSuperAdmin bool) ([]*domainpostreply.Reply, error) {
	if _, err := s.Posts.GetForViewer(ctx, postID, viewerID, viewerIsSuperAdmin); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.Replies.ListByPostID(ctx, postID, afterID, limit)
}

func (s *Service) Add(ctx context.Context, userID, postID int64, body string, parentReplyID *int64) (*domainpostreply.Reply, error) {
	if _, err := s.Posts.GetPublic(ctx, postID); err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" || len(body) > maxReplyBody {
		return nil, ErrInvalidInput
	}
	words := appmoderation.SensitiveWords()
	if hits := appmoderation.FindSensitiveHits(body, words); len(hits) > 0 {
		return nil, appmoderation.ErrSensitiveWithHits(hits)
	}
	if parentReplyID != nil {
		parent, err := s.Replies.GetByID(ctx, *parentReplyID)
		if err != nil {
			return nil, err
		}
		if parent.PostID != postID || parent.DeletedAt != nil || parent.Status != domainpostreply.StatusVisible {
			return nil, ErrInvalidInput
		}
	}
	rep := &domainpostreply.Reply{
		ID:            s.NewID(),
		PostID:        postID,
		UserID:        userID,
		ParentReplyID: parentReplyID,
		Body:          body,
		Status:        domainpostreply.StatusVisible,
	}
	if err := s.Replies.Create(ctx, rep); err != nil {
		return nil, err
	}
	return rep, nil
}

// GetByID loads a reply by id (for notifications / internal use).
func (s *Service) GetByID(ctx context.Context, replyID int64) (*domainpostreply.Reply, error) {
	return s.Replies.GetByID(ctx, replyID)
}

func (s *Service) DeleteOwn(ctx context.Context, userID, replyID int64) error {
	rep, err := s.Replies.GetByID(ctx, replyID)
	if err != nil {
		return err
	}
	if rep.UserID != userID {
		return ErrForbidden
	}
	if rep.DeletedAt != nil {
		return domainpostreply.ErrNotFound
	}
	return s.Replies.SoftDelete(ctx, replyID)
}
