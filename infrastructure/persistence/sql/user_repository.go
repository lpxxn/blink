package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domainuser "github.com/lpxxn/blink/domain/user"
)

type UserRepository struct {
	DB *sql.DB
}

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

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT snowflake_id, email, name, wechat_id, phone, password_hash, password_salt, status, role
		FROM users WHERE snowflake_id = ? AND deleted_at IS NULL
	`, id)
	var u domainuser.User
	err := row.Scan(
		&u.SnowflakeID, &u.Email, &u.Name, &u.WechatID, &u.Phone,
		&u.PasswordHash, &u.PasswordSalt, &u.Status, &u.Role,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
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
