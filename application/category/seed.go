package category

import (
	"context"
	"errors"

	domaincategory "github.com/lpxxn/blink/domain/category"
)

var defaultCategories = []struct {
	slug  string
	name  string
	order int
}{
	{"general", "综合", 0},
	{"tech", "科技", 10},
	{"life", "生活", 20},
	{"fun", "娱乐", 30},
}

// SeedDefaults inserts built-in categories when the table is empty.
func SeedDefaults(ctx context.Context, repo domaincategory.Repository, newID func() int64) error {
	n, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, d := range defaultCategories {
		c := &domaincategory.Category{
			ID:        newID(),
			Slug:      d.slug,
			Name:      d.name,
			SortOrder: d.order,
		}
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// ErrInvalidCategory is returned when a category id does not resolve.
var ErrInvalidCategory = errors.New("category: invalid or missing")
