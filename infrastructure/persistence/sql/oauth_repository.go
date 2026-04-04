package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	"github.com/jmoiron/sqlx"
)

// OAuthRepository implements domain/oauth.Repository using sqlx.
type OAuthRepository struct {
	DB *sqlx.DB
}

type oauthRow struct {
	SnowflakeID       int64  `db:"snowflake_id"`
	Provider          string `db:"provider"`
	ProviderSubject   string `db:"provider_subject"`
	UserID            int64  `db:"user_id"`
}

func (r *oauthRow) toDomain() *domainoauth.Identity {
	return &domainoauth.Identity{
		SnowflakeID:     r.SnowflakeID,
		Provider:        r.Provider,
		ProviderSubject: r.ProviderSubject,
		UserID:          r.UserID,
	}
}

func (r *OAuthRepository) FindByProviderSubject(ctx context.Context, provider, subject string) (*domainoauth.Identity, error) {
	var row oauthRow
	err := r.DB.GetContext(ctx, &row, `
		SELECT snowflake_id, provider, provider_subject, user_id
		FROM oauth_identities
		WHERE provider = ? AND provider_subject = ? AND deleted_at IS NULL
	`, provider, subject)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainoauth.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *OAuthRepository) Create(ctx context.Context, id *domainoauth.Identity) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO oauth_identities (snowflake_id, provider, provider_subject, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id.SnowflakeID, id.Provider, id.ProviderSubject, id.UserID, time.Now().UTC(), time.Now().UTC())
	return err
}
