package block

import (
	"context"
	"errors"

	domainblock "github.com/lpxxn/blink/domain/block"
	domainfollow "github.com/lpxxn/blink/domain/follow"
	domainuser "github.com/lpxxn/blink/domain/user"
)

var ErrUserNotFound = errors.New("block: user not found")

type Service struct {
	Blocks  domainblock.Repository
	Follows domainfollow.Repository
	Users   domainuser.Repository
}

func (s *Service) Block(ctx context.Context, blockerID, blockedID int64) error {
	if blockerID == blockedID {
		return domainblock.ErrSelfBlock
	}
	if _, err := s.Users.GetByID(ctx, blockedID); err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := s.Blocks.Block(ctx, blockerID, blockedID); err != nil {
		return err
	}
	if s.Follows != nil {
		_ = s.Follows.Unfollow(ctx, blockerID, blockedID)
		_ = s.Follows.Unfollow(ctx, blockedID, blockerID)
	}
	return nil
}

func (s *Service) Unblock(ctx context.Context, blockerID, blockedID int64) error {
	if blockerID == blockedID {
		return domainblock.ErrSelfBlock
	}
	return s.Blocks.Unblock(ctx, blockerID, blockedID)
}

func (s *Service) IsEitherBlocked(ctx context.Context, a, b int64) (bool, error) {
	return s.Blocks.IsEitherBlocked(ctx, a, b)
}

func (s *Service) ListBlocked(ctx context.Context, blockerID int64, offset, limit int) ([]domainuser.PublicProfile, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.Blocks.ListBlocked(ctx, blockerID, offset, limit)
}
