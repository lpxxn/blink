package user

import "time"

// AdminListEntry is a row for super-admin user listing (extra columns vs login User).
type AdminListEntry struct {
	SnowflakeID     int64
	Email           string
	Name            string
	Status          int
	Role            string
	LastLoginAt     *time.Time
	LastLoginIP     string
	LastLoginDevice string
	CreatedAt       time.Time
}
