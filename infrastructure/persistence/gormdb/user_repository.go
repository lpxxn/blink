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
