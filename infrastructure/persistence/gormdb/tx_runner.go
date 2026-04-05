package gormdb

import (
	"context"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"gorm.io/gorm"
)

// TxRunner implements application/auth.TxRunner using one GORM transaction.
type TxRunner struct {
	DB *gorm.DB
}

func (r *TxRunner) Run(ctx context.Context, fn func(ctx context.Context, users domainuser.Repository, identities domainoauth.Repository) error) error {
	return WithTransaction(r.DB, func(tx *gorm.DB) error {
		return fn(ctx,
			&UserRepository{DB: tx},
			&OAuthRepository{DB: tx},
		)
	})
}
