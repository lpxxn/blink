package gormdb

import (
	"time"

	"gorm.io/gorm"
)

// UserModel maps the users table (see platform/db/0001_init.sql).
type UserModel struct {
	SnowflakeID       int64          `gorm:"column:snowflake_id;primaryKey"`
	Email             string         `gorm:"column:email;size:255;not null;uniqueIndex:idx_users_email"`
	Name              string         `gorm:"column:name;size:128;not null"`
	WechatID          string         `gorm:"column:wechat_id;size:128;not null"`
	Phone             string         `gorm:"column:phone;size:32;not null"`
	PasswordHash      string         `gorm:"column:password_hash;size:255;not null"`
	PasswordSalt      string         `gorm:"column:password_salt;size:255;not null"`
	LastLoginAt       *time.Time     `gorm:"column:last_login_at"`
	LastLoginIP       *string        `gorm:"column:last_login_ip;size:45"`
	LastLoginDevice   *string        `gorm:"column:last_login_device;size:512"`
	Status            int            `gorm:"column:status;not null"`
	Role              string         `gorm:"column:role;size:32;not null"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserModel) TableName() string {
	return "users"
}

// OAuthIdentityModel maps oauth_identities (platform/db/0004_oauth_identities.sql).
type OAuthIdentityModel struct {
	SnowflakeID     int64          `gorm:"column:snowflake_id;primaryKey"`
	Provider        string         `gorm:"column:provider;size:32;not null;uniqueIndex:idx_oauth_provider_subject"`
	ProviderSubject string         `gorm:"column:provider_subject;size:255;not null;uniqueIndex:idx_oauth_provider_subject"`
	UserID          int64          `gorm:"column:user_id;not null;index:idx_oauth_user_id"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (OAuthIdentityModel) TableName() string {
	return "oauth_identities"
}
