package admin

import (
	"context"
	"strings"

	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	domainsensitiveword "github.com/lpxxn/blink/domain/sensitiveword"
)

func (s *Service) afterSensitiveWordsChange(ctx context.Context) {
	if s.ReloadSensitiveWords != nil {
		_ = s.ReloadSensitiveWords(ctx)
	}
	if s.SensitiveWordsPublisher != nil {
		_ = s.SensitiveWordsPublisher.PublishSensitiveWordsChanged(ctx)
	}
}

func normalizeSensitiveWord(w string) string {
	return strings.ToLower(strings.TrimSpace(w))
}

// CreateSensitiveWord adds an enabled word (normalized to lower case).
func (s *Service) CreateSensitiveWord(ctx context.Context, word string) (*domainsensitiveword.Word, error) {
	if s.SensitiveWords == nil || s.NewID == nil {
		return nil, ErrInvalidSensitiveWord
	}
	n := normalizeSensitiveWord(word)
	if n == "" {
		return nil, ErrInvalidSensitiveWord
	}
	dw := &domainsensitiveword.Word{
		ID:      s.NewID(),
		Word:    n,
		Enabled: true,
	}
	if err := s.SensitiveWords.Create(ctx, dw); err != nil {
		return nil, err
	}
	s.afterSensitiveWordsChange(ctx)
	return s.SensitiveWords.GetByID(ctx, dw.ID)
}

// ListSensitiveWords returns a page of stored words (newest id first).
func (s *Service) ListSensitiveWords(ctx context.Context, offset, limit int) ([]*domainsensitiveword.Word, int64, error) {
	if s.SensitiveWords == nil {
		return nil, 0, ErrInvalidSensitiveWord
	}
	return s.SensitiveWords.ListPage(ctx, offset, limit)
}

// PatchSensitiveWord updates enabled flag when set.
func (s *Service) PatchSensitiveWord(ctx context.Context, id int64, enabled *bool) (*domainsensitiveword.Word, error) {
	if s.SensitiveWords == nil {
		return nil, ErrInvalidSensitiveWord
	}
	if enabled != nil {
		if err := s.SensitiveWords.UpdateEnabled(ctx, id, *enabled); err != nil {
			return nil, err
		}
		s.afterSensitiveWordsChange(ctx)
	}
	return s.SensitiveWords.GetByID(ctx, id)
}

// DeleteSensitiveWord removes a row by id.
func (s *Service) DeleteSensitiveWord(ctx context.Context, id int64) error {
	if s.SensitiveWords == nil {
		return ErrInvalidSensitiveWord
	}
	if err := s.SensitiveWords.Delete(ctx, id); err != nil {
		return err
	}
	s.afterSensitiveWordsChange(ctx)
	return nil
}

// HidePostReply hides a reply and all descendants.
func (s *Service) HidePostReply(ctx context.Context, replyID int64) error {
	if s.Replies == nil {
		return ErrRepliesNotConfigured
	}
	r, err := s.Replies.GetByID(ctx, replyID)
	if err != nil {
		return err
	}
	if r.DeletedAt != nil {
		return domainpostreply.ErrNotFound
	}
	return s.Replies.HideSubtree(ctx, replyID)
}

func (s *Service) UnhidePostReply(ctx context.Context, replyID int64) error {
	if s.Replies == nil {
		return ErrRepliesNotConfigured
	}
	r, err := s.Replies.GetByID(ctx, replyID)
	if err != nil {
		return err
	}
	if r.DeletedAt != nil {
		return domainpostreply.ErrNotFound
	}
	return s.Replies.UnhideSubtree(ctx, replyID)
}

func (s *Service) ListPostReplies(ctx context.Context, postID int64, afterID *int64, limit int) ([]*domainpostreply.Reply, error) {
	if s.Replies == nil || s.Posts == nil {
		return nil, ErrRepliesNotConfigured
	}
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.DeletedAt != nil {
		return nil, domainpost.ErrNotFound
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.Replies.ListByPostIDAllStatuses(ctx, postID, afterID, limit)
}
