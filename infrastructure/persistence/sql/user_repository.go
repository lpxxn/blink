package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/jmoiron/sqlx"
)

// UserRepository implements domain/user.Repository using sqlx.
type UserRepository struct {
	DB *sqlx.DB
}

// userRow maps DB columns; kept in persistence to avoid leaking `db` tags into domain.
type userRow struct {
	SnowflakeID  int64  `db:"snowflake_id"`
	Email        string `db:"email"`
	Name         string `db:"name"`
	WechatID     string `db:"wechat_id"`
	Phone        string `db:"phone"`
	PasswordHash string `db:"password_hash"`
	PasswordSalt string `db:"password_salt"`
	Status       int    `db:"status"`
	Role         string `db:"role"`
}

func (r *userRow) toDomain() *domainuser.User {
	return &domainuser.User{
		SnowflakeID:  r.SnowflakeID,
		Email:        r.Email,
		Name:         r.Name,
		WechatID:     r.WechatID,
		Phone:        r.Phone,
		PasswordHash: r.PasswordHash,
		PasswordSalt: r.PasswordSalt,
		Status:       r.Status,
		Role:         r.Role,
	}
}

const userSelectCols = `snowflake_id, email, name, wechat_id, phone, password_hash, password_salt, status, role`

func (r *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO users (
			snowflake_id, email, name, wechat_id, phone,
			password_hash, password_salt, status, role,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		u.SnowflakeID, u.Email, u.Name, u.WechatID, u.Phone,
		u.PasswordHash, u.PasswordSalt, u.Status, u.Role,
		time.Now().UTC(), time.Now().UTC(),
	)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var row userRow
	err := r.DB.GetContext(ctx, &row, `
		SELECT `+userSelectCols+`
		FROM users WHERE email = ? AND deleted_at IS NULL
	`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	var row userRow
	err := r.DB.GetContext(ctx, &row, `
		SELECT `+userSelectCols+`
		FROM users WHERE snowflake_id = ? AND deleted_at IS NULL
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, ip, device string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE users SET
			last_login_at = ?,
			last_login_ip = ?,
			last_login_device = ?,
			updated_at = ?
		WHERE snowflake_id = ? AND deleted_at IS NULL
	`, time.Now().UTC(), nullString(ip), nullString(device), time.Now().UTC(), id)
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
