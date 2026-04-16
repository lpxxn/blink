package sensitiveword

import "time"

// Word is a stored moderation term (normalized word text is unique).
type Word struct {
	ID        int64
	Word      string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
