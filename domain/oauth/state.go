package oauth

import (
	"context"
	"time"
)

// RedirectState is stored server-side while the user completes authorization at the IdP.
type RedirectState struct {
	Provider string
	NextURL  string
}

// StateStore holds short-lived CSRF state for the OAuth2 authorization round-trip.
type StateStore interface {
	Save(ctx context.Context, state string, p RedirectState, ttl time.Duration) error
	Consume(ctx context.Context, state string) (RedirectState, error)
}
