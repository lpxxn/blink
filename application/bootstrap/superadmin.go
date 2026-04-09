package bootstrap

import (
	"context"
	"errors"
	"strings"

	domainuser "github.com/lpxxn/blink/domain/user"
)

// PromoteSuperAdminFromEnv sets role to super_admin for the given email when it is still "user".
func PromoteSuperAdminFromEnv(ctx context.Context, users domainuser.Repository, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}
	u, err := users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return nil
		}
		return err
	}
	if u.Role != domainuser.RoleUser {
		return nil
	}
	r := domainuser.RoleSuperAdmin
	return users.UpdateStatusRole(ctx, u.SnowflakeID, nil, &r)
}
