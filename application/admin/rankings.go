package admin

import (
	"context"
	"strconv"
	"time"
)

// RankingPeriod defines today / this-month / this-year windows.
type RankingPeriod struct {
	Label string
	Since time.Time
	Until time.Time
}

func rankingPeriods(now time.Time) []RankingPeriod {
	now = now.UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture := now.Add(24 * time.Hour)
	return []RankingPeriod{
		{Label: "today", Since: todayStart, Until: farFuture},
		{Label: "month", Since: monthStart, Until: farFuture},
		{Label: "year", Since: yearStart, Until: farFuture},
	}
}

// PostRanking is the top posters for one period.
type PostRanking struct {
	Period string                  `json:"period"`
	Items  []PostRankingItem       `json:"items"`
}

type PostRankingItem struct {
	UserID    int64  `json:"user_id,string"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	PostCount int64  `json:"post_count,string"`
}

// UserActivityRanking is the active users for one period.
type UserActivityRanking struct {
	Period string                      `json:"period"`
	Items  []UserActivityRankingItem   `json:"items"`
}

type UserActivityRankingItem struct {
	UserID     int64  `json:"user_id,string"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	PostCount  int64  `json:"post_count,string"`
	ReplyCount int64  `json:"reply_count,string"`
	LikeCount  int64  `json:"like_count,string"`
	Total      int64  `json:"total,string"`
}

type Rankings struct {
	PostRankings     []PostRanking         `json:"post_rankings"`
	ActivityRankings []UserActivityRanking `json:"activity_rankings"`
}

func (s *Service) Rankings(ctx context.Context, limit int) (*Rankings, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	periods := rankingPeriods(time.Now())
	result := &Rankings{}

	for _, p := range periods {
		pr, err := s.postRanking(ctx, p, limit)
		if err != nil {
			return nil, err
		}
		result.PostRankings = append(result.PostRankings, *pr)

		ar, err := s.activityRanking(ctx, p, limit)
		if err != nil {
			return nil, err
		}
		result.ActivityRankings = append(result.ActivityRankings, *ar)
	}
	return result, nil
}

func (s *Service) postRanking(ctx context.Context, p RankingPeriod, limit int) (*PostRanking, error) {
	rows, err := s.Posts.TopPosters(ctx, p.Since, p.Until, limit)
	if err != nil {
		return nil, err
	}
	items := make([]PostRankingItem, 0, len(rows))
	for _, r := range rows {
		name, email := s.resolveUser(ctx, r.UserID)
		items = append(items, PostRankingItem{
			UserID:    r.UserID,
			UserName:  name,
			UserEmail: email,
			PostCount: r.PostCount,
		})
	}
	return &PostRanking{Period: p.Label, Items: items}, nil
}

func (s *Service) activityRanking(ctx context.Context, p RankingPeriod, limit int) (*UserActivityRanking, error) {
	rows, err := s.Users.TopActiveUsers(ctx, p.Since, p.Until, limit)
	if err != nil {
		return nil, err
	}
	items := make([]UserActivityRankingItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, UserActivityRankingItem{
			UserID:     r.UserID,
			UserName:   displayName(r.Name, r.UserID),
			UserEmail:  r.Email,
			PostCount:  r.PostCount,
			ReplyCount: r.ReplyCount,
			LikeCount:  r.LikeCount,
			Total:      r.Total,
		})
	}
	return &UserActivityRanking{Period: p.Label, Items: items}, nil
}

func (s *Service) resolveUser(ctx context.Context, id int64) (name, email string) {
	u, err := s.Users.GetByID(ctx, id)
	if err != nil || u == nil {
		return displayName("", id), ""
	}
	return displayName(u.Name, id), u.Email
}

func displayName(name string, id int64) string {
	if name != "" {
		return name
	}
	return "用户 " + strconv.FormatInt(id, 10)
}
