package gormdb

import (
	"context"
	"errors"
	"time"

	domainreplylike "github.com/lpxxn/blink/domain/replylike"
	"gorm.io/gorm"
)

type ReplyLikeRepository struct {
	DB *gorm.DB
}

func (r *ReplyLikeRepository) Like(ctx context.Context, userID, replyID int64) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m ReplyLikeModel
		err := tx.Unscoped().Where("user_id = ? AND reply_id = ?", userID, replyID).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&ReplyLikeModel{
				UserID:    userID,
				ReplyID:   replyID,
				CreatedAt: now,
				UpdatedAt: now,
			}).Error
		}
		if err != nil {
			return err
		}
		if m.DeletedAt.Valid {
			return tx.Unscoped().Model(&ReplyLikeModel{}).
				Where("user_id = ? AND reply_id = ?", userID, replyID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"updated_at": now,
				}).Error
		}
		return domainreplylike.ErrAlreadyLiked
	})
}

func (r *ReplyLikeRepository) Unlike(ctx context.Context, userID, replyID int64) error {
	res := r.DB.WithContext(ctx).
		Where("user_id = ? AND reply_id = ? AND deleted_at IS NULL", userID, replyID).
		Delete(&ReplyLikeModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainreplylike.ErrNotFound
	}
	return nil
}

func (r *ReplyLikeRepository) IsLiked(ctx context.Context, userID, replyID int64) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&ReplyLikeModel{}).
		Where("user_id = ? AND reply_id = ? AND deleted_at IS NULL", userID, replyID).
		Count(&n).Error
	return n > 0, err
}

func (r *ReplyLikeRepository) CountByReplyID(ctx context.Context, replyID int64) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&ReplyLikeModel{}).
		Where("reply_id = ? AND deleted_at IS NULL", replyID).
		Count(&n).Error
	return n, err
}

func (r *ReplyLikeRepository) CountByReplyIDs(ctx context.Context, replyIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	if len(replyIDs) == 0 {
		return out, nil
	}
	type row struct {
		ReplyID int64
		Cnt     int64
	}
	var rows []row
	err := r.DB.WithContext(ctx).Model(&ReplyLikeModel{}).
		Select("reply_id, COUNT(*) AS cnt").
		Where("reply_id IN ? AND deleted_at IS NULL", replyIDs).
		Group("reply_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.ReplyID] = r.Cnt
	}
	return out, nil
}

func (r *ReplyLikeRepository) LikedReplyIDs(ctx context.Context, userID int64, replyIDs []int64) (map[int64]bool, error) {
	out := make(map[int64]bool)
	if len(replyIDs) == 0 || userID == 0 {
		return out, nil
	}
	var ids []int64
	err := r.DB.WithContext(ctx).Model(&ReplyLikeModel{}).
		Where("user_id = ? AND reply_id IN ? AND deleted_at IS NULL", userID, replyIDs).
		Pluck("reply_id", &ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}
