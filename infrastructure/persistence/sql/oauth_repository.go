package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
)

type OAuthRepository struct {
	DB *sql.DB
}

func (r *OAuthRepository) FindByProviderSubject(ctx context.Context, provider, subject string) (*domainoauth.Identity, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT snowflake_id, provider, provider_subject, user_id
		FROM oauth_identities
		WHERE provider = ? AND provider_subject = ? AND deleted_at IS NULL
	`, provider, subject)
	var o domainoauth.Identity
	err := row.Scan(&o.SnowflakeID, &o.Provider, &o.ProviderSubject, &o.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainoauth.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OAuthRepository) Create(ctx context.Context, id *domainoauth.Identity) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO oauth_identities (snowflake_id, provider, provider_subject, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id.SnowflakeID, id.Provider, id.ProviderSubject, id.UserID, time.Now().UTC(), time.Now().UTC())
	return err
}
