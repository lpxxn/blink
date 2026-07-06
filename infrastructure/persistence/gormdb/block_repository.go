package gormdb

import (
	"context"
	"errors"
	"time"

	domainblock "github.com/lpxxn/blink/domain/block"
	domainuser "github.com/lpxxn/blink/domain/user"
	"gorm.io/gorm"
)

type BlockRepository struct {
	DB *gorm.DB
}

func (r *BlockRepository) Block(ctx context.Context, blockerID, blockedID int64) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m UserBlockModel
		err := tx.Unscoped().Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&UserBlockModel{
				BlockerID: blockerID,
				BlockedID: blockedID,
				CreatedAt: now,
				UpdatedAt: now,
			}).Error
		}
		if err != nil {
			return err
		}
		if m.DeletedAt.Valid {
			return tx.Unscoped().Model(&UserBlockModel{}).
				Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"updated_at": now,
				}).Error
		}
		return domainblock.ErrAlreadyBlocked
	})
}

func (r *BlockRepository) Unblock(ctx context.Context, blockerID, blockedID int64) error {
	res := r.DB.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ? AND deleted_at IS NULL", blockerID, blockedID).
		Delete(&UserBlockModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainblock.ErrNotFound
	}
	return nil
}

func (r *BlockRepository) IsBlocked(ctx context.Context, blockerID, blockedID int64) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserBlockModel{}).
		Where("blocker_id = ? AND blocked_id = ? AND deleted_at IS NULL", blockerID, blockedID).
		Count(&n).Error
	return n > 0, err
}

func (r *BlockRepository) IsEitherBlocked(ctx context.Context, a, b int64) (bool, error) {
	if a == 0 || b == 0 || a == b {
		return false, nil
	}
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserBlockModel{}).
		Where("deleted_at IS NULL AND ((blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?))",
			a, b, b, a).
		Count(&n).Error
	return n > 0, err
}

func (r *BlockRepository) ListBlocked(ctx context.Context, blockerID int64, offset, limit int) ([]domainuser.PublicProfile, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	type row struct {
		SnowflakeID int64  `gorm:"column:snowflake_id"`
		Name        string `gorm:"column:name"`
	}
	var rows []row
	err := r.DB.WithContext(ctx).Model(&UserBlockModel{}).
		Select("users.snowflake_id, users.name").
		Joins("INNER JOIN users ON users.snowflake_id = user_blocks.blocked_id").
		Where("user_blocks.blocker_id = ? AND user_blocks.deleted_at IS NULL", blockerID).
		Where("users.status = ?", domainuser.StatusActive).
		Order("user_blocks.created_at DESC, user_blocks.blocked_id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainuser.PublicProfile, 0, len(rows))
	for i := range rows {
		out = append(out, domainuser.PublicProfile{
			SnowflakeID: rows[i].SnowflakeID,
			Name:        rows[i].Name,
		})
	}
	return out, nil
}
