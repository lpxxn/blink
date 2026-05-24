package follow

import (
	"context"
	"errors"
	"log"

	appnotification "github.com/lpxxn/blink/application/notification"
	domainfollow "github.com/lpxxn/blink/domain/follow"
	domainuser "github.com/lpxxn/blink/domain/user"
)

var ErrUserNotFound = errors.New("follow: user not found")

type Service struct {
	Follows       domainfollow.Repository
	Users         domainuser.Repository
	Notifications *appnotification.Service // optional; sync in-app notification on follow
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
	if err := s.Follows.Follow(ctx, followerID, followeeID); err != nil {
		return err
	}
	if s.Notifications != nil && followerID != followeeID {
		if err := s.Notifications.OnUserFollowed(ctx, followeeID, followerID); err != nil {
			log.Printf("follow: notification: %v", err)
		}
	}
	return nil
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return domainfollow.ErrSelfFollow
	}
	return s.Follows.Unfollow(ctx, followerID, followeeID)
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
