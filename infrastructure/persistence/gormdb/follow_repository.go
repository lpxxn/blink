package gormdb

import (
	"context"
	"errors"
	"time"

	domainfollow "github.com/lpxxn/blink/domain/follow"
	"gorm.io/gorm"
)

type FollowRepository struct {
	DB *gorm.DB
}

func (r *FollowRepository) Follow(ctx context.Context, followerID, followeeID int64) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m UserFollowModel
		err := tx.Unscoped().Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&UserFollowModel{
				FollowerID: followerID,
				FolloweeID: followeeID,
				CreatedAt:  now,
				UpdatedAt:  now,
			}).Error
		}
		if err != nil {
			return err
		}
		if m.DeletedAt.Valid {
			return tx.Unscoped().Model(&UserFollowModel{}).
				Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"updated_at": now,
				}).Error
		}
		return domainfollow.ErrAlreadyExists
	})
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	res := r.DB.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", followerID, followeeID).
		Delete(&UserFollowModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainfollow.ErrNotFound
	}
	return nil
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserFollowModel{}).
		Where("follower_id = ? AND followee_id = ? AND deleted_at IS NULL", followerID, followeeID).
		Count(&n).Error
	return n > 0, err
}

func (r *FollowRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserFollowModel{}).
		Where("followee_id = ? AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *FollowRepository) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserFollowModel{}).
		Where("follower_id = ? AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}
