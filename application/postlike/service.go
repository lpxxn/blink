package postlike

import (
	"context"
	"errors"

	apppost "github.com/lpxxn/blink/application/post"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostlike "github.com/lpxxn/blink/domain/postlike"
)

var ErrPostNotFound = errors.New("postlike: post not found")

type Service struct {
	Likes domainpostlike.Repository
	Posts *apppost.Service
}

func (s *Service) Like(ctx context.Context, userID, postID int64) error {
	if _, err := s.Posts.GetPublic(ctx, postID); err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			return ErrPostNotFound
		}
		return err
	}
	return s.Likes.Like(ctx, userID, postID)
}

func (s *Service) Unlike(ctx context.Context, userID, postID int64) error {
	return s.Likes.Unlike(ctx, userID, postID)
}

type PostLikeMeta struct {
	LikeCount int64
	Liked     bool
}

func (s *Service) MetaForPost(ctx context.Context, postID int64, viewerID *int64) (*PostLikeMeta, error) {
	cnt, err := s.Likes.CountByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}
	meta := &PostLikeMeta{LikeCount: cnt}
	if viewerID != nil && *viewerID != 0 {
		ok, err := s.Likes.IsLiked(ctx, *viewerID, postID)
		if err != nil {
			return nil, err
		}
		meta.Liked = ok
	}
	return meta, nil
}

func (s *Service) MetaForPosts(ctx context.Context, postIDs []int64, viewerID *int64) (map[int64]PostLikeMeta, error) {
	counts, err := s.Likes.CountByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]PostLikeMeta, len(postIDs))
	for _, id := range postIDs {
		out[id] = PostLikeMeta{LikeCount: counts[id]}
	}
	if viewerID != nil && *viewerID != 0 {
		liked, err := s.Likes.LikedPostIDs(ctx, *viewerID, postIDs)
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
