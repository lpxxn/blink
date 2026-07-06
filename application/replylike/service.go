package replylike

import (
	"context"
	"errors"

	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	domainreplylike "github.com/lpxxn/blink/domain/replylike"
)

var ErrReplyNotFound = errors.New("replylike: reply not found")

type Service struct {
	Likes   domainreplylike.Repository
	Replies *apppostreply.Service
}

func (s *Service) Like(ctx context.Context, userID, replyID int64) error {
	if err := s.ensureLikeable(ctx, replyID); err != nil {
		return err
	}
	return s.Likes.Like(ctx, userID, replyID)
}

func (s *Service) Unlike(ctx context.Context, userID, replyID int64) error {
	return s.Likes.Unlike(ctx, userID, replyID)
}

func (s *Service) ensureLikeable(ctx context.Context, replyID int64) error {
	rep, err := s.Replies.GetByID(ctx, replyID)
	if err != nil {
		if errors.Is(err, domainpostreply.ErrNotFound) {
			return ErrReplyNotFound
		}
		return err
	}
	if rep.DeletedAt != nil || rep.Status != domainpostreply.StatusVisible {
		return ErrReplyNotFound
	}
	if _, err := s.Replies.Posts.GetPublic(ctx, rep.PostID); err != nil {
		if errors.Is(err, domainpost.ErrNotFound) || errors.Is(err, apppost.ErrNotVisible) {
			return ErrReplyNotFound
		}
		return err
	}
	return nil
}

type ReplyLikeMeta struct {
	LikeCount int64
	Liked     bool
}

func (s *Service) VerifyPublic(ctx context.Context, replyID int64) error {
	return s.ensureLikeable(ctx, replyID)
}

func (s *Service) MetaForReply(ctx context.Context, replyID int64, viewerID *int64) (*ReplyLikeMeta, error) {
	cnt, err := s.Likes.CountByReplyID(ctx, replyID)
	if err != nil {
		return nil, err
	}
	meta := &ReplyLikeMeta{LikeCount: cnt}
	if viewerID != nil && *viewerID != 0 {
		ok, err := s.Likes.IsLiked(ctx, *viewerID, replyID)
		if err != nil {
			return nil, err
		}
		meta.Liked = ok
	}
	return meta, nil
}

func (s *Service) MetaForReplies(ctx context.Context, replyIDs []int64, viewerID *int64) (map[int64]ReplyLikeMeta, error) {
	counts, err := s.Likes.CountByReplyIDs(ctx, replyIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]ReplyLikeMeta, len(replyIDs))
	for _, id := range replyIDs {
		out[id] = ReplyLikeMeta{LikeCount: counts[id]}
	}
	if viewerID != nil && *viewerID != 0 {
		liked, err := s.Likes.LikedReplyIDs(ctx, *viewerID, replyIDs)
		if err != nil {
			return nil, err
		}
		for id, ok := range liked {
			if ok {
				m := out[id]
				m.Liked = true
				out[id] = m
			}
		}
	}
	return out, nil
}
