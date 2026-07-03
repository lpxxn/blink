package gormdb

import (
	"context"
	"errors"
	"strings"
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

func (r *UserRepository) adminListQuery(ctx context.Context, query string) *gorm.DB {
	q := r.DB.WithContext(ctx).Model(&UserModel{})
	query = strings.TrimSpace(query)
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("email LIKE ? OR name LIKE ? OR CAST(snowflake_id AS TEXT) LIKE ?", like, like, like)
	}
	return q
}

func (r *UserRepository) ListForAdmin(ctx context.Context, query string, offset, limit int) ([]domainuser.AdminListEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows []UserModel
	err := r.adminListQuery(ctx, query).Order("snowflake_id DESC").Offset(offset).Limit(limit).Find(&rows).Error
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

func (r *UserRepository) CountForAdmin(ctx context.Context, query string) (int64, error) {
	var n int64
	err := r.adminListQuery(ctx, query).Count(&n).Error
	return n, err
}

func (r *UserRepository) SearchPublic(ctx context.Context, query string, offset, limit int) ([]domainuser.PublicProfile, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	pattern := likeContainsPattern(query)
	var rows []UserModel
	err := r.DB.WithContext(ctx).Model(&UserModel{}).
		Where("status = ?", domainuser.StatusActive).
		Where("name LIKE ? ESCAPE '\\'", pattern).
		Order("snowflake_id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
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

func (r *UserRepository) TopActiveUsers(ctx context.Context, since, until time.Time, limit int) ([]domainuser.UserActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	// Three sub-queries for post/reply/like counts, then UNION ALL + GROUP BY.
	// Portable across SQLite/MySQL/PostgreSQL.
	sql := `
SELECT u.snowflake_id AS user_id, u.name, u.email,
       COALESCE(p.cnt, 0) AS post_count,
       COALESCE(r.cnt, 0) AS reply_count,
       COALESCE(l.cnt, 0) AS like_count,
       COALESCE(p.cnt, 0) + COALESCE(r.cnt, 0) + COALESCE(l.cnt, 0) AS total
FROM users u
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt FROM posts
    WHERE deleted_at IS NULL AND status = 1 AND created_at >= ? AND created_at < ?
    GROUP BY user_id
) p ON p.user_id = u.snowflake_id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt FROM post_replies
    WHERE deleted_at IS NULL AND created_at >= ? AND created_at < ?
    GROUP BY user_id
) r ON r.user_id = u.snowflake_id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt FROM post_likes
    WHERE deleted_at IS NULL AND created_at >= ? AND created_at < ?
    GROUP BY user_id
) l ON l.user_id = u.snowflake_id
WHERE (COALESCE(p.cnt, 0) + COALESCE(r.cnt, 0) + COALESCE(l.cnt, 0)) > 0
ORDER BY total DESC, u.snowflake_id ASC
LIMIT ?
`
	type row struct {
		UserID     int64  `gorm:"column:user_id"`
		Name       string `gorm:"column:name"`
		Email      string `gorm:"column:email"`
		PostCount  int64  `gorm:"column:post_count"`
		ReplyCount int64  `gorm:"column:reply_count"`
		LikeCount  int64  `gorm:"column:like_count"`
		Total      int64  `gorm:"column:total"`
	}
	var rows []row
	err := r.DB.WithContext(ctx).Raw(sql,
		since, until,
		since, until,
		since, until,
		limit,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainuser.UserActivity, len(rows))
	for i, r := range rows {
		out[i] = domainuser.UserActivity{
			UserID:     r.UserID,
			Name:       r.Name,
			Email:      r.Email,
			PostCount:  r.PostCount,
			ReplyCount: r.ReplyCount,
			LikeCount:  r.LikeCount,
			Total:      r.Total,
		}
	}
	return out, nil
}
