package block

import (
	"context"

	domainuser "github.com/lpxxn/blink/domain/user"
)

// Repository persists user block edges (user_blocks table).
type Repository interface {
	Block(ctx context.Context, blockerID, blockedID int64) error
	Unblock(ctx context.Context, blockerID, blockedID int64) error
	IsBlocked(ctx context.Context, blockerID, blockedID int64) (bool, error)
	// IsEitherBlocked is true when blocker blocked blocked or the reverse.
	IsEitherBlocked(ctx context.Context, a, b int64) (bool, error)
	ListBlocked(ctx context.Context, blockerID int64, offset, limit int) ([]domainuser.PublicProfile, error)
}
