package auth

import (
	"context"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// TxRunner runs a callback with transactional user + oauth repositories.
// Application 层用其包裹「必须同事务提交」的注册逻辑；实现放在 infrastructure（如 gormdb）。
type TxRunner interface {
	Run(ctx context.Context, fn func(ctx context.Context, users domainuser.Repository, identities domainoauth.Repository) error) error
}
