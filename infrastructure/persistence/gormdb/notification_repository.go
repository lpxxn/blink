package gormdb

import (
	"context"
	"time"

	domainnotification "github.com/lpxxn/blink/domain/notification"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	DB *gorm.DB
}

func notifModelToDomain(m *NotificationModel) *domainnotification.Notification {
	return &domainnotification.Notification{
		ID:         m.ID,
		UserID:     m.UserID,
		Type:       m.Type,
		Title:      m.Title,
		Body:       m.Body,
		RefPostID:  m.RefPostID,
		RefReplyID: m.RefReplyID,
		ReadAt:     m.ReadAt,
		CreatedAt:  m.CreatedAt,
	}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domainnotification.Notification) error {
	now := time.Now().UTC()
	m := &NotificationModel{
		ID:         n.ID,
		UserID:     n.UserID,
		Type:       n.Type,
		Title:      n.Title,
		Body:       n.Body,
		RefPostID:  n.RefPostID,
		RefReplyID: n.RefReplyID,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*domainnotification.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Model(&NotificationModel{}).Where("user_id = ?", userID)
	if beforeID != nil {
		q = q.Where("id < ?", *beforeID)
	}
	var rows []NotificationModel
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainnotification.Notification, 0, len(rows))
	for i := range rows {
		out = append(out, notifModelToDomain(&rows[i]))
	}
	return out, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, id int64) error {
	now := time.Now().UTC()
	res := r.DB.WithContext(ctx).Model(&NotificationModel{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).
		Update("read_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID int64) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Model(&NotificationModel{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&NotificationModel{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}

var _ domainnotification.Repository = (*NotificationRepository)(nil)
