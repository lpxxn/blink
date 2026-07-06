package search

import (
	"context"
	"errors"
	"strings"

	domainpost "github.com/lpxxn/blink/domain/post"
	domainuser "github.com/lpxxn/blink/domain/user"
)

const (
	minQueryLen = 1
	maxQueryLen = 100
)

var ErrInvalidQuery = errors.New("search: invalid query")

type Service struct {
	Posts domainpost.Repository
	Users domainuser.Repository
}

func normalizeQuery(q string) (string, error) {
	q = strings.TrimSpace(q)
	if len(q) < minQueryLen {
		return "", ErrInvalidQuery
	}
	if len(q) > maxQueryLen {
		q = q[:maxQueryLen]
	}
	return q, nil
}

func clampLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 20
	}
	return limit
}

func (s *Service) SearchPosts(ctx context.Context, rawQuery string, beforeID *int64, limit int, viewerID *int64) ([]*domainpost.Post, error) {
	query, err := normalizeQuery(rawQuery)
	if err != nil {
		return nil, err
	}
	if s.Posts == nil {
		return nil, errors.New("search: posts unavailable")
	}
	return s.Posts.SearchPublic(ctx, query, beforeID, clampLimit(limit), viewerID)
}

func (s *Service) SearchUsers(ctx context.Context, rawQuery string, offset, limit int) ([]domainuser.PublicProfile, error) {
	query, err := normalizeQuery(rawQuery)
	if err != nil {
		return nil, err
	}
	if s.Users == nil {
		return nil, errors.New("search: users unavailable")
	}
	if offset < 0 {
		offset = 0
	}
	return s.Users.SearchPublic(ctx, query, offset, clampLimit(limit))
}
