package gormdb

import (
	"context"
	"errors"
	"strings"
	"time"

	domainsensitiveword "github.com/lpxxn/blink/domain/sensitiveword"
	"gorm.io/gorm"
)

type SensitiveWordRepository struct {
	DB *gorm.DB
}

func sensitiveWordModelToDomain(m *SensitiveWordModel) *domainsensitiveword.Word {
	return &domainsensitiveword.Word{
		ID:        m.ID,
		Word:      m.Word,
		Enabled:   m.Enabled != 0,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (r *SensitiveWordRepository) Create(ctx context.Context, w *domainsensitiveword.Word) error {
	now := time.Now().UTC()
	m := &SensitiveWordModel{
		ID:        w.ID,
		Word:      strings.TrimSpace(w.Word),
		Enabled:   boolToInt(w.Enabled),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if !w.CreatedAt.IsZero() {
		m.CreatedAt = w.CreatedAt
	}
	if !w.UpdatedAt.IsZero() {
		m.UpdatedAt = w.UpdatedAt
	}
	var existing int64
	if err := r.DB.WithContext(ctx).Model(&SensitiveWordModel{}).Where("word = ?", m.Word).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return domainsensitiveword.ErrDuplicateWord
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *SensitiveWordRepository) GetByID(ctx context.Context, id int64) (*domainsensitiveword.Word, error) {
	var m SensitiveWordModel
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainsensitiveword.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return sensitiveWordModelToDomain(&m), nil
}

func (r *SensitiveWordRepository) ListEnabledWords(ctx context.Context) ([]string, error) {
	var rows []string
	err := r.DB.WithContext(ctx).Model(&SensitiveWordModel{}).
		Where("enabled = ?", 1).
		Order("id ASC").
		Pluck("word", &rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SensitiveWordRepository) ListPage(ctx context.Context, offset, limit int) ([]*domainsensitiveword.Word, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int64
	if err := r.DB.WithContext(ctx).Model(&SensitiveWordModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ms []SensitiveWordModel
	if err := r.DB.WithContext(ctx).Order("id DESC").Offset(offset).Limit(limit).Find(&ms).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*domainsensitiveword.Word, 0, len(ms))
	for i := range ms {
		out = append(out, sensitiveWordModelToDomain(&ms[i]))
	}
	return out, total, nil
}

func (r *SensitiveWordRepository) UpdateEnabled(ctx context.Context, id int64, enabled bool) error {
	res := r.DB.WithContext(ctx).Model(&SensitiveWordModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    boolToInt(enabled),
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainsensitiveword.ErrNotFound
	}
	return nil
}

func (r *SensitiveWordRepository) Delete(ctx context.Context, id int64) error {
	res := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&SensitiveWordModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainsensitiveword.ErrNotFound
	}
	return nil
}
