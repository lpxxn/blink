package httpapi

import (
	"errors"
	"strconv"
	"strings"
	"time"

	domainfollow "github.com/lpxxn/blink/domain/follow"
)

func parseFollowListCursor(raw string) (*domainfollow.PageCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty cursor")
	}
	parts := strings.SplitN(raw, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("bad cursor")
	}
	nano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, errors.New("bad cursor")
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, errors.New("bad cursor")
	}
	return &domainfollow.PageCursor{
		CreatedAt: time.Unix(0, nano).UTC(),
		UserID:    userID,
	}, nil
}

func formatFollowListCursor(e domainfollow.ListEntry) string {
	return strconv.FormatInt(e.CreatedAt.UTC().UnixNano(), 10) + "," + strconv.FormatInt(e.UserID, 10)
}
