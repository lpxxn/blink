package postlike

import (
	"context"
	"errors"
	"time"

	apppost "github.com/lpxxn/blink/application/post"
	domainpost "github.com/lpxxn/blink/domain/post"
)

var ErrInvalidPeriod = errors.New("postlike: invalid period")

const (
	PeriodDay   = "day"
	PeriodWeek  = "week"
	PeriodMonth = "month"
)

type LikeRankingEntry struct {
	Post            *domainpost.Post
	PeriodLikeCount int64
}

func likeRankingWindow(period string, now time.Time) (since, until time.Time, err error) {
	now = now.UTC()
	farFuture := now.Add(24 * time.Hour)
	switch period {
	case PeriodDay:
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return since, farFuture, nil
	case PeriodWeek:
		since = now.Add(-7 * 24 * time.Hour)
		return since, farFuture, nil
	case PeriodMonth:
		since = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return since, farFuture, nil
	default:
		return time.Time{}, time.Time{}, ErrInvalidPeriod
	}
}

func (s *Service) LikeRankings(ctx context.Context, period string, limit int) ([]LikeRankingEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	since, until, err := likeRankingWindow(period, time.Now())
	if err != nil {
		return nil, err
	}
	ranks, err := s.Likes.TopLikedPosts(ctx, since, until, limit)
	if err != nil {
		return nil, err
	}
	out := make([]LikeRankingEntry, 0, len(ranks))
	for _, r := range ranks {
		p, err := s.Posts.GetPublic(ctx, r.PostID)
		if err != nil {
			if errors.Is(err, domainpost.ErrNotFound) || errors.Is(err, apppost.ErrNotVisible) {
				continue
			}
			return nil, err
		}
		out = append(out, LikeRankingEntry{Post: p, PeriodLikeCount: r.LikeCount})
	}
	return out, nil
}
