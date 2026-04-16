package gormdb

import (
	"context"
	"errors"
	"time"

	domainsettings "github.com/lpxxn/blink/domain/settings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppSettingsRepository struct {
	DB *gorm.DB
}

func (r *AppSettingsRepository) GetString(ctx context.Context, key string) (string, error) {
	var m AppSettingModel
	err := r.DB.WithContext(ctx).Model(&AppSettingModel{}).Where("key = ?", key).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domainsettings.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return m.Value, nil
}

func (r *AppSettingsRepository) SetString(ctx context.Context, key, value string) error {
	now := time.Now().UTC()
	m := &AppSettingModel{
		Key:       key,
		Value:     value,
		UpdatedAt: now,
	}
	// Use upsert for portability: INSERT ... ON CONFLICT(key) DO UPDATE (SQLite/Postgres),
	// and for MySQL this becomes ON DUPLICATE KEY UPDATE under gorm.
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(m).Error
}

