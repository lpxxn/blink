package follow

import (
	"context"
	"errors"

	domainfollow "github.com/lpxxn/blink/domain/follow"
	domainuser "github.com/lpxxn/blink/domain/user"
)

var ErrUserNotFound = errors.New("follow: user not found")

type Service struct {
	Follows domainfollow.Repository
	Users   domainuser.Repository
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return domainfollow.ErrSelfFollow
	}
	if _, err := s.Users.GetByID(ctx, followeeID); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.Follows.Follow(ctx, followerID, followeeID)
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return domainfollow.ErrSelfFollow
	}
	return s.Follows.Unfollow(ctx, followerID, followeeID)
}

func (s *Service) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	return s.Follows.IsFollowing(ctx, followerID, followeeID)
}

type Stats struct {
	FollowerCount  int64
	FollowingCount int64
	IsFollowing    bool
}

func (s *Service) Stats(ctx context.Context, targetUserID int64, viewerID *int64) (*Stats, error) {
	if _, err := s.Users.GetByID(ctx, targetUserID); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	followers, err := s.Follows.CountFollowers(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	following, err := s.Follows.CountFollowing(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	st := &Stats{FollowerCount: followers, FollowingCount: following}
	if viewerID != nil && *viewerID != 0 {
		ok, err := s.Follows.IsFollowing(ctx, *viewerID, targetUserID)
		if err != nil {
			return nil, err
		}
		st.IsFollowing = ok
	}
	return st, nil
}

func (s *Service) ListFollowing(ctx context.Context, targetUserID int64, cursor *domainfollow.PageCursor, limit int) ([]domainfollow.ListEntry, error) {
	if _, err := s.Users.GetByID(ctx, targetUserID); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.Follows.ListFollowing(ctx, targetUserID, cursor, limit)
}

func (s *Service) ListFollowers(ctx context.Context, targetUserID int64, cursor *domainfollow.PageCursor, limit int) ([]domainfollow.ListEntry, error) {
	if _, err := s.Users.GetByID(ctx, targetUserID); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.Follows.ListFollowers(ctx, targetUserID, cursor, limit)
}
