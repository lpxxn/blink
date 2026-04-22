package oauth

import "context"

type Identity struct {
	SnowflakeID     int64
	Provider        string
	ProviderSubject string
	UserID          int64
}

type Repository interface {
	FindByProviderSubject(ctx context.Context, provider, subject string) (*Identity, error)
	Create(ctx context.Context, id *Identity) error
}
