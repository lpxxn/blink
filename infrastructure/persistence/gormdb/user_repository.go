package gormdb

import (
	"context"
	"errors"
	"time"

	domainuser "github.com/lpxxn/blink/domain/user"
	"gorm.io/gorm"
)

// UserRepository implements domain/user.Repository with GORM.
type UserRepository struct {
	DB *gorm.DB
}

func domainToUserModel(u *domainuser.User) *UserModel {
	return &UserModel{
		SnowflakeID:  u.SnowflakeID,
		Email:        u.Email,
		Name:         u.Name,
		WechatID:     u.WechatID,
		Phone:        u.Phone,
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		Status:       u.Status,
		Role:         u.Role,
	}
}

func userModelToDomain(m *UserModel) *domainuser.User {
	return &domainuser.User{
		SnowflakeID:  m.SnowflakeID,
		Email:        m.Email,
		Name:         m.Name,
		WechatID:     m.WechatID,
		Phone:        m.Phone,
		PasswordHash: m.PasswordHash,
		PasswordSalt: m.PasswordSalt,
		Status:       m.Status,
		Role:         m.Role,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	now := time.Now().UTC()
	m := domainToUserModel(u)
	m.CreatedAt = now
	m.UpdatedAt = now
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var m UserModel
	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return userModelToDomain(&m), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	var m UserModel
	err := r.DB.WithContext(ctx).Where("snowflake_id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return userModelToDomain(&m), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, ip, device string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"last_login_at":     now,
		"last_login_ip":     nullStringPtr(ip),
		"last_login_device": nullStringPtr(device),
		"updated_at":        now,
	}
	return r.DB.WithContext(ctx).Model(&UserModel{}).Where("snowflake_id = ?", id).Updates(updates).Error
}

func nullStringPtr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (r *UserRepository) ListSnowflakeIDsByRole(ctx context.Context, role string) ([]int64, error) {
	var ids []int64
	err := r.DB.WithContext(ctx).Model(&UserModel{}).Where("role = ?", role).Order("snowflake_id ASC").Pluck("snowflake_id", &ids).Error
	return ids, err
}

func (r *UserRepository) ListForAdmin(ctx context.Context, offset, limit int) ([]domainuser.AdminListEntry, error) {
	var rows []UserModel
	err := r.DB.WithContext(ctx).Model(&UserModel{}).Order("snowflake_id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainuser.AdminListEntry, 0, len(rows))
	for i := range rows {
		out = append(out, domainuser.AdminListEntry{
			SnowflakeID:     rows[i].SnowflakeID,
			Email:           rows[i].Email,
			Name:            rows[i].Name,
			Status:          rows[i].Status,
			Role:            rows[i].Role,
			LastLoginAt:     rows[i].LastLoginAt,
			LastLoginIP:     derefString(rows[i].LastLoginIP),
			LastLoginDevice: derefString(rows[i].LastLoginDevice),
			CreatedAt:       rows[i].CreatedAt,
		})
	}
	return out, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&UserModel{}).Count(&n).Error
	return n, err
}

func (r *UserRepository) UpdateStatusRole(ctx context.Context, id int64, status *int, role *string) error {
	updates := map[string]interface{}{
		"updated_at": time.Now().UTC(),
	}
	if status != nil {
		updates["status"] = *status
	}
	if role != nil {
		updates["role"] = *role
	}
	if len(updates) == 1 {
		return nil
	}
	return r.DB.WithContext(ctx).Model(&UserModel{}).Where("snowflake_id = ?", id).Updates(updates).Error
}

func (r *UserRepository) UpdateName(ctx context.Context, id int64, name string) error {
	now := time.Now().UTC()
	return r.DB.WithContext(ctx).Model(&UserModel{}).Where("snowflake_id = ?", id).Updates(map[string]interface{}{
		"name":       name,
		"updated_at": now,
	}).Error
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	now := time.Now().UTC()
	res := r.DB.WithContext(ctx).Model(&UserModel{}).Where("snowflake_id = ?", id).Updates(map[string]interface{}{
		"password_hash": passwordHash,
		"updated_at":    now,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainuser.ErrNotFound
	}
	return nil
}
