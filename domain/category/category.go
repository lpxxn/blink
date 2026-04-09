package category

import "time"

type Category struct {
	ID        int64
	Slug      string
	Name      string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}
