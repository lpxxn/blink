package gormdb

import (
	"context"
	"errors"
	"time"

	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	"gorm.io/gorm"
)

type PostReplyRepository struct {
	DB *gorm.DB
}

func replyModelToDomain(m *PostReplyModel) *domainpostreply.Reply {
	var deleted *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deleted = &t
	}
	return &domainpostreply.Reply{
		ID:            m.ID,
		PostID:        m.PostID,
		UserID:        m.UserID,
		ParentReplyID: m.ParentReplyID,
		Body:          m.Body,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     deleted,
	}
}

func (r *PostReplyRepository) Create(ctx context.Context, rep *domainpostreply.Reply) error {
	now := time.Now().UTC()
	m := &PostReplyModel{
		ID:            rep.ID,
		PostID:        rep.PostID,
		UserID:        rep.UserID,
		ParentReplyID: rep.ParentReplyID,
		Body:          rep.Body,
		Status:        rep.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if !rep.CreatedAt.IsZero() {
		m.CreatedAt = rep.CreatedAt
	}
	if !rep.UpdatedAt.IsZero() {
		m.UpdatedAt = rep.UpdatedAt
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *PostReplyRepository) GetByID(ctx context.Context, id int64) (*domainpostreply.Reply, error) {
	var m PostReplyModel
	err := r.DB.WithContext(ctx).Unscoped().Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainpostreply.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return replyModelToDomain(&m), nil
}

func (r *PostReplyRepository) ListByPostID(ctx context.Context, postID int64, afterID *int64, limit int) ([]*domainpostreply.Reply, error) {
	q := r.DB.WithContext(ctx).Model(&PostReplyModel{}).
		Where("post_id = ? AND deleted_at IS NULL AND status = ?", postID, domainpostreply.StatusVisible)
	if afterID != nil {
		q = q.Where("id > ?", *afterID)
	}
	var rows []PostReplyModel
	if err := q.Order("id ASC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpostreply.Reply, 0, len(rows))
	for i := range rows {
		out = append(out, replyModelToDomain(&rows[i]))
	}
	return out, nil
}

func (r *PostReplyRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&PostReplyModel{}).Error
}
