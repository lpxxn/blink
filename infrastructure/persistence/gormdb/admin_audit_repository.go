package gormdb

import (
	"context"
	"time"

	domainadminaudit "github.com/lpxxn/blink/domain/adminaudit"
	"gorm.io/gorm"
)

type AdminAuditRepository struct {
	DB *gorm.DB
}

func auditModelToDomain(m *AdminAuditLogModel) *domainadminaudit.Entry {
	return &domainadminaudit.Entry{
		ID:         m.ID,
		ActorID:    m.ActorID,
		Action:     m.Action,
		TargetType: m.TargetType,
		TargetID:   m.TargetID,
		Detail:     m.Detail,
		CreatedAt:  m.CreatedAt,
	}
}

func (r *AdminAuditRepository) Create(ctx context.Context, e *domainadminaudit.Entry) error {
	now := time.Now().UTC()
	m := &AdminAuditLogModel{
		ID:         e.ID,
		ActorID:    e.ActorID,
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Detail:     e.Detail,
		CreatedAt:  e.CreatedAt,
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *AdminAuditRepository) List(ctx context.Context, offset, limit int) ([]*domainadminaudit.Entry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows []AdminAuditLogModel
	err := r.DB.WithContext(ctx).Model(&AdminAuditLogModel{}).
		Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domainadminaudit.Entry, 0, len(rows))
	for i := range rows {
		out = append(out, auditModelToDomain(&rows[i]))
	}
	return out, nil
}

func (r *AdminAuditRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&AdminAuditLogModel{}).Count(&n).Error
	return n, err
}

var _ domainadminaudit.Repository = (*AdminAuditRepository)(nil)
