package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apppostlike "github.com/lpxxn/blink/application/postlike"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) ListLikeRankings(c *gin.Context) {
	if s.Likes == nil {
		c.JSON(http.StatusOK, LikeRankingPageJSON{Period: c.Query("period"), Posts: []LikeRankingPostJSON{}})
		return
	}
	period := c.DefaultQuery("period", apppostlike.PeriodWeek)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := s.Likes.LikeRankings(c.Request.Context(), period, limit)
	if err != nil {
		if errors.Is(err, apppostlike.ErrInvalidPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period; use day, week, or month"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}

	postIDs := make([]int64, len(list))
	for i, e := range list {
		postIDs[i] = e.Post.ID
	}
	metaMap, err := s.Likes.MetaForPosts(c.Request.Context(), postIDs, viewer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userIDs := make([]int64, 0, len(list))
	for _, e := range list {
		userIDs = append(userIDs, e.Post.UserID)
	}
	names := ResolveUserNames(c.Request.Context(), s.Users, userIDs)

	out := make([]LikeRankingPostJSON, 0, len(list))
	for i, e := range list {
		j := LikeRankingPostJSON{
			PostJSON:        PostToJSON(e.Post),
			PeriodLikeCount: e.PeriodLikeCount,
			Rank:            i + 1,
		}
		j.UserName = names[e.Post.UserID]
		if m, ok := metaMap[e.Post.ID]; ok {
			j.LikeCount = m.LikeCount
			if viewer != nil {
				ok := m.Liked
				j.Liked = &ok
			}
		}
		out = append(out, j)
	}
	c.JSON(http.StatusOK, LikeRankingPageJSON{Period: period, Posts: out})
}
