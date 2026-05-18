package admin

import (
	"context"
	"errors"
	"regexp"
	"strings"

	domainadminaudit "github.com/lpxxn/blink/domain/adminaudit"
	domaincategory "github.com/lpxxn/blink/domain/category"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

var ErrInvalidCategory = errors.New("admin: invalid category")

func normalizeSlug(slug string) (string, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if !slugPattern.MatchString(slug) {
		return "", ErrInvalidCategory
	}
	return slug, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]*domaincategory.Category, error) {
	if s.Categories == nil {
		return nil, ErrInvalidCategory
	}
	return s.Categories.ListAll(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, actorID int64, slug, name string, sortOrder int) (*domaincategory.Category, error) {
	if s.Categories == nil || s.NewID == nil {
		return nil, ErrInvalidCategory
	}
	slug, err := normalizeSlug(slug)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidCategory
	}
	if _, err := s.Categories.GetBySlug(ctx, slug); err == nil {
		return nil, domaincategory.ErrDuplicate
	} else if !errors.Is(err, domaincategory.ErrNotFound) {
		return nil, err
	}
	c := &domaincategory.Category{
		ID:        s.NewID(),
		Slug:      slug,
		Name:      name,
		SortOrder: sortOrder,
	}
	if err := s.Categories.Create(ctx, c); err != nil {
		return nil, err
	}
	cid := c.ID
	s.logAudit(ctx, actorID, AuditCategoryCreate, "category", &cid, slug+": "+name)
	return s.Categories.GetByID(ctx, c.ID)
}

func (s *Service) UpdateCategory(ctx context.Context, actorID, id int64, slug, name *string, sortOrder *int) (*domaincategory.Category, error) {
	if s.Categories == nil {
		return nil, ErrInvalidCategory
	}
	c, err := s.Categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if slug != nil {
		norm, err := normalizeSlug(*slug)
		if err != nil {
			return nil, err
		}
		if other, err := s.Categories.GetBySlug(ctx, norm); err == nil && other.ID != id {
			return nil, domaincategory.ErrDuplicate
		} else if err != nil && !errors.Is(err, domaincategory.ErrNotFound) {
			return nil, err
		}
		c.Slug = norm
	}
	if name != nil {
		n := strings.TrimSpace(*name)
		if n == "" {
			return nil, ErrInvalidCategory
		}
		c.Name = n
	}
	if sortOrder != nil {
		c.SortOrder = *sortOrder
	}
	if err := s.Categories.Update(ctx, c); err != nil {
		return nil, err
	}
	cid := id
	s.logAudit(ctx, actorID, AuditCategoryUpdate, "category", &cid, c.Slug+": "+c.Name)
	return s.Categories.GetByID(ctx, id)
}

func (s *Service) DeleteCategory(ctx context.Context, actorID, id int64) error {
	if s.Categories == nil {
		return ErrInvalidCategory
	}
	if _, err := s.Categories.GetByID(ctx, id); err != nil {
		return err
	}
	if err := s.Categories.SoftDelete(ctx, id); err != nil {
		return err
	}
	cid := id
	s.logAudit(ctx, actorID, AuditCategoryDelete, "category", &cid, "")
	return nil
}

func (s *Service) ListAuditLogs(ctx context.Context, offset, limit int) ([]*domainadminaudit.Entry, int64, error) {
	if s.Audit == nil {
		return []*domainadminaudit.Entry{}, 0, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	list, err := s.Audit.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.Audit.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
