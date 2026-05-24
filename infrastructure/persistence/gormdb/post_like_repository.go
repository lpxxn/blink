package gormdb

import (
	"context"
	"errors"
	"time"

	domainpostlike "github.com/lpxxn/blink/domain/postlike"
	"gorm.io/gorm"
)

type PostLikeRepository struct {
	DB *gorm.DB
}

func (r *PostLikeRepository) Like(ctx context.Context, userID, postID int64) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m PostLikeModel
		err := tx.Unscoped().Where("user_id = ? AND post_id = ?", userID, postID).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&PostLikeModel{
				UserID:    userID,
				PostID:    postID,
				CreatedAt: now,
				UpdatedAt: now,
			}).Error
		}
		if err != nil {
			return err
		}
		if m.DeletedAt.Valid {
			return tx.Unscoped().Model(&PostLikeModel{}).
				Where("user_id = ? AND post_id = ?", userID, postID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"updated_at": now,
				}).Error
		}
		return domainpostlike.ErrAlreadyLiked
	})
}

func (r *PostLikeRepository) Unlike(ctx context.Context, userID, postID int64) error {
	res := r.DB.WithContext(ctx).
		Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", userID, postID).
		Delete(&PostLikeModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainpostlike.ErrNotFound
	}
	return nil
}

func (r *PostLikeRepository) IsLiked(ctx context.Context, userID, postID int64) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&PostLikeModel{}).
		Where("user_id = ? AND post_id = ? AND deleted_at IS NULL", userID, postID).
		Count(&n).Error
	return n > 0, err
}

func (r *PostLikeRepository) CountByPostID(ctx context.Context, postID int64) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&PostLikeModel{}).
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Count(&n).Error
	return n, err
}

func (r *PostLikeRepository) CountByPostIDs(ctx context.Context, postIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	if len(postIDs) == 0 {
		return out, nil
	}
	type row struct {
		PostID int64
		Cnt    int64
	}
	var rows []row
	err := r.DB.WithContext(ctx).Model(&PostLikeModel{}).
		Select("post_id, COUNT(*) AS cnt").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.PostID] = r.Cnt
	}
	return out, nil
}

func (r *PostLikeRepository) LikedPostIDs(ctx context.Context, userID int64, postIDs []int64) (map[int64]bool, error) {
	out := make(map[int64]bool)
	if len(postIDs) == 0 || userID == 0 {
		return out, nil
	}
	var ids []int64
	err := r.DB.WithContext(ctx).Model(&PostLikeModel{}).
		Where("user_id = ? AND post_id IN ? AND deleted_at IS NULL", userID, postIDs).
		Pluck("post_id", &ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}
