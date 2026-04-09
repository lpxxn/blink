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

// CategoryModel maps categories (platform/db/0005_post_categories_moderation.sql).
type CategoryModel struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	Slug      string         `gorm:"column:slug;size:64;not null;uniqueIndex:idx_categories_slug"`
	Name      string         `gorm:"column:name;size:128;not null"`
	SortOrder int            `gorm:"column:sort_order;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (CategoryModel) TableName() string {
	return "categories"
}

// PostModel maps posts (platform/db/0002 + 0005).
type PostModel struct {
	ID               int64          `gorm:"column:id;primaryKey"`
	UserID           int64          `gorm:"column:user_id;not null;index:idx_posts_user_id"`
	PostType         int            `gorm:"column:post_type;not null"`
	ReplyToPostID    *int64         `gorm:"column:reply_to_post_id"`
	ReferencedPostID *int64         `gorm:"column:referenced_post_id"`
	Visibility       int            `gorm:"column:visibility;not null"`
	AudienceListID   *int64         `gorm:"column:audience_list_id"`
	CategoryID       *int64         `gorm:"column:category_id;index:idx_posts_category_id"`
	Body             string         `gorm:"column:body;type:text;not null"`
	Images           string         `gorm:"column:images;type:text;not null"`
	Status           int            `gorm:"column:status;not null"`
	ModerationFlag   int            `gorm:"column:moderation_flag;not null"`
	ModerationNote   string         `gorm:"column:moderation_note;type:text;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (PostModel) TableName() string {
	return "posts"
}

// PostReplyModel maps post_replies (platform/db/0003_post_replies.sql).
type PostReplyModel struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	PostID        int64          `gorm:"column:post_id;not null;index:idx_post_replies_post_id_created_at"`
	UserID        int64          `gorm:"column:user_id;not null"`
	ParentReplyID *int64         `gorm:"column:parent_reply_id"`
	Body          string         `gorm:"column:body;type:text;not null"`
	Status        int            `gorm:"column:status;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (PostReplyModel) TableName() string {
	return "post_replies"
}
