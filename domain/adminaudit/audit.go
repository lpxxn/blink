package adminaudit

import "time"

type Entry struct {
	ID         int64
	ActorID    int64
	Action     string
	TargetType string
	TargetID   *int64
	Detail     string
	CreatedAt  time.Time
}
