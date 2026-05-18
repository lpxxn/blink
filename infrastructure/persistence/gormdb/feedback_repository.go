package gormdb

import (
	"context"
	"errors"
	"strings"
	"time"

	domainfeedback "github.com/lpxxn/blink/domain/feedback"
	"gorm.io/gorm"
)

type FeedbackRepository struct {
	DB *gorm.DB
}

func feedbackThreadToDomain(m *FeedbackThreadModel) *domainfeedback.Thread {
	return &domainfeedback.Thread{
		ID:             m.ID,
		UserID:         m.UserID,
		Status:         m.Status,
		UserReplyCount: m.UserReplyCount,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		LastMessageAt:  m.LastMessageAt,
	}
}

func feedbackMessageToDomain(m *FeedbackMessageModel) *domainfeedback.Message {
	return &domainfeedback.Message{
		ID:         m.ID,
		FeedbackID: m.FeedbackID,
		SenderID:   m.SenderID,
		SenderType: m.SenderType,
		Body:       m.Body,
		CreatedAt:  m.CreatedAt,
	}
}

func feedbackThreadModel(t *domainfeedback.Thread, now time.Time) *FeedbackThreadModel {
	created := t.CreatedAt
	if created.IsZero() {
		created = now
	}
	updated := t.UpdatedAt
	if updated.IsZero() {
		updated = now
	}
	last := t.LastMessageAt
	if last.IsZero() {
		last = now
	}
	status := t.Status
	if status == "" {
		status = domainfeedback.StatusOpen
	}
	return &FeedbackThreadModel{
		ID:             t.ID,
		UserID:         t.UserID,
		Status:         status,
		UserReplyCount: t.UserReplyCount,
		CreatedAt:      created,
		UpdatedAt:      updated,
		LastMessageAt:  last,
	}
}

func feedbackMessageModel(m *domainfeedback.Message, now time.Time) *FeedbackMessageModel {
	created := m.CreatedAt
	if created.IsZero() {
		created = now
	}
	return &FeedbackMessageModel{
		ID:         m.ID,
		FeedbackID: m.FeedbackID,
		SenderID:   m.SenderID,
		SenderType: m.SenderType,
		Body:       m.Body,
		CreatedAt:  created,
	}
}

func (r *FeedbackRepository) CreateThreadWithMessage(ctx context.Context, thread *domainfeedback.Thread, message *domainfeedback.Message) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(feedbackThreadModel(thread, now)).Error; err != nil {
			return err
		}
		return tx.Create(feedbackMessageModel(message, now)).Error
	})
}

func (r *FeedbackRepository) GetThreadByID(ctx context.Context, id int64) (*domainfeedback.Thread, error) {
	var m FeedbackThreadModel
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainfeedback.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return feedbackThreadToDomain(&m), nil
}

func (r *FeedbackRepository) ListByUserID(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*domainfeedback.Thread, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := r.DB.WithContext(ctx).Model(&FeedbackThreadModel{}).Where("user_id = ?", userID)
	if beforeID != nil {
		q = q.Where("id < ?", *beforeID)
	}
	var rows []FeedbackThreadModel
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainfeedback.Thread, 0, len(rows))
	for i := range rows {
		out = append(out, feedbackThreadToDomain(&rows[i]))
	}
	return out, nil
}

func feedbackListQuery(db *gorm.DB, f domainfeedback.ListFilters) *gorm.DB {
	q := db.Model(&FeedbackThreadModel{})
	if f.Status != nil && strings.TrimSpace(*f.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(*f.Status))
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	return q
}

func (r *FeedbackRepository) ListPage(ctx context.Context, f domainfeedback.ListFilters, offset, limit int) ([]*domainfeedback.Thread, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int64
	if err := feedbackListQuery(r.DB.WithContext(ctx), f).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []FeedbackThreadModel
	err := feedbackListQuery(r.DB.WithContext(ctx), f).
		Order("last_message_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domainfeedback.Thread, 0, len(rows))
	for i := range rows {
		out = append(out, feedbackThreadToDomain(&rows[i]))
	}
	return out, total, nil
}

func (r *FeedbackRepository) Count(ctx context.Context, f domainfeedback.ListFilters) (int64, error) {
	var n int64
	err := feedbackListQuery(r.DB.WithContext(ctx), f).Count(&n).Error
	return n, err
}

func (r *FeedbackRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	now := time.Now().UTC()
	res := r.DB.WithContext(ctx).Model(&FeedbackThreadModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainfeedback.ErrNotFound
	}
	return nil
}

func (r *FeedbackRepository) ListMessages(ctx context.Context, feedbackID int64) ([]*domainfeedback.Message, error) {
	var rows []FeedbackMessageModel
	err := r.DB.WithContext(ctx).Model(&FeedbackMessageModel{}).
		Where("feedback_id = ?", feedbackID).Order("id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domainfeedback.Message, 0, len(rows))
	for i := range rows {
		out = append(out, feedbackMessageToDomain(&rows[i]))
	}
	return out, nil
}

func (r *FeedbackRepository) AddMessageAndTouch(ctx context.Context, message *domainfeedback.Message, userReplyDelta int) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(feedbackMessageModel(message, now)).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"updated_at":      now,
			"last_message_at": now,
		}
		if userReplyDelta != 0 {
			updates["user_reply_count"] = gorm.Expr("user_reply_count + ?", userReplyDelta)
		}
		return tx.Model(&FeedbackThreadModel{}).Where("id = ?", message.FeedbackID).Updates(updates).Error
	})
}

var _ domainfeedback.Repository = (*FeedbackRepository)(nil)
