package admin

import (
	"context"
	"errors"
	"time"

	domainpost "github.com/lpxxn/blink/domain/post"
	domainuser "github.com/lpxxn/blink/domain/user"
)

var (
	ErrCannotDemoteSelf   = errors.New("admin: cannot remove own super_admin role")
	ErrInvalidRole        = errors.New("admin: invalid role")
	ErrInvalidModeration  = errors.New("admin: invalid moderation flag")
)

type Service struct {
	Users domainuser.Repository
	Posts domainpost.Repository
}

type Overview struct {
	UserCount      int64
	PostCount      int64
	PostsToday     int64
	CategoryCount  int64 // optional, set from handler if needed
}

func (s *Service) Overview(ctx context.Context) (*Overview, error) {
	uc, err := s.Users.Count(ctx)
	if err != nil {
		return nil, err
	}
	pc, err := s.Posts.Count(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Now().UTC().Truncate(24 * time.Hour)
	pt, err := s.Posts.CountCreatedSince(ctx, start)
	if err != nil {
		return nil, err
	}
	return &Overview{UserCount: uc, PostCount: pc, PostsToday: pt}, nil
}

func validRole(r string) bool {
	switch r {
	case domainuser.RoleUser, domainuser.RoleAdmin, domainuser.RoleSuperAdmin:
		return true
	default:
		return false
	}
}

func (s *Service) PatchUser(ctx context.Context, actorID, targetID int64, status *int, role *string) error {
	if role != nil && !validRole(*role) {
		return ErrInvalidRole
	}
	if actorID == targetID && role != nil && *role != domainuser.RoleSuperAdmin {
		u, err := s.Users.GetByID(ctx, actorID)
		if err != nil {
			return err
		}
		if u.Role == domainuser.RoleSuperAdmin {
			return ErrCannotDemoteSelf
		}
	}
	return s.Users.UpdateStatusRole(ctx, targetID, status, role)
}

func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]domainuser.AdminListEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.Users.ListForAdmin(ctx, offset, limit)
}

func (s *Service) ListPosts(ctx context.Context, f domainpost.AdminListFilters, offset, limit int) ([]*domainpost.Post, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.Posts.AdminList(ctx, f, offset, limit)
}

func (s *Service) PatchPost(ctx context.Context, postID int64, moderationFlag *int, moderationNote *string, status *int) (*domainpost.Post, error) {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if moderationFlag != nil {
		switch *moderationFlag {
		case domainpost.ModerationNormal, domainpost.ModerationFlagged, domainpost.ModerationRemoved:
			p.ModerationFlag = *moderationFlag
		default:
			return nil, ErrInvalidModeration
		}
	}
	if moderationNote != nil {
		p.ModerationNote = *moderationNote
	}
	if status != nil {
		switch *status {
		case domainpost.StatusDraft, domainpost.StatusPublished, domainpost.StatusHidden:
			p.Status = *status
		default:
			return nil, errors.New("admin: invalid post status")
		}
	}
	if err := s.Posts.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
