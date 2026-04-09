package gormdb

import (
	"context"
	"errors"
	"time"

	domaincategory "github.com/lpxxn/blink/domain/category"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	DB *gorm.DB
}

func categoryModelToDomain(m *CategoryModel) *domaincategory.Category {
	return &domaincategory.Category{
		ID:        m.ID,
		Slug:      m.Slug,
		Name:      m.Name,
		SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (r *CategoryRepository) Create(ctx context.Context, c *domaincategory.Category) error {
	now := time.Now().UTC()
	m := &CategoryModel{
		ID:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		SortOrder: c.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if !c.CreatedAt.IsZero() {
		m.CreatedAt = c.CreatedAt
	}
	if !c.UpdatedAt.IsZero() {
		m.UpdatedAt = c.UpdatedAt
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domaincategory.Category, error) {
	var m CategoryModel
	err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domaincategory.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return categoryModelToDomain(&m), nil
}

func (r *CategoryRepository) ListActive(ctx context.Context) ([]*domaincategory.Category, error) {
	var rows []CategoryModel
	err := r.DB.WithContext(ctx).Where("deleted_at IS NULL").Order("sort_order ASC, id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domaincategory.Category, 0, len(rows))
	for i := range rows {
		out = append(out, categoryModelToDomain(&rows[i]))
	}
	return out, nil
}

func (r *CategoryRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&CategoryModel{}).Where("deleted_at IS NULL").Count(&n).Error
	return n, err
}
