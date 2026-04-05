package gormdb

import (
	"context"
	"errors"
	"time"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	"gorm.io/gorm"
)

// OAuthRepository implements domain/oauth.Repository with GORM.
type OAuthRepository struct {
	DB *gorm.DB
}

func identityToModel(id *domainoauth.Identity) *OAuthIdentityModel {
	return &OAuthIdentityModel{
		SnowflakeID:     id.SnowflakeID,
		Provider:        id.Provider,
		ProviderSubject: id.ProviderSubject,
		UserID:          id.UserID,
	}
}

func identityModelToDomain(m *OAuthIdentityModel) *domainoauth.Identity {
	return &domainoauth.Identity{
		SnowflakeID:     m.SnowflakeID,
		Provider:        m.Provider,
		ProviderSubject: m.ProviderSubject,
		UserID:          m.UserID,
	}
}

func (r *OAuthRepository) FindByProviderSubject(ctx context.Context, provider, subject string) (*domainoauth.Identity, error) {
	var m OAuthIdentityModel
	err := r.DB.WithContext(ctx).Where("provider = ? AND provider_subject = ?", provider, subject).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainoauth.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return identityModelToDomain(&m), nil
}

func (r *OAuthRepository) Create(ctx context.Context, id *domainoauth.Identity) error {
	now := time.Now().UTC()
	m := identityToModel(id)
	m.CreatedAt = now
	m.UpdatedAt = now
	return r.DB.WithContext(ctx).Create(m).Error
}
