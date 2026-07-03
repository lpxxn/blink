package user

// PublicProfile is a minimal user row for public search results (no email).
type PublicProfile struct {
	SnowflakeID int64
	Name        string
}
