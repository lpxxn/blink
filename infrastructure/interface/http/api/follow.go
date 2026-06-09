package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appfollow "github.com/lpxxn/blink/application/follow"
	domainfollow "github.com/lpxxn/blink/domain/follow"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

type FollowStatsJSON struct {
	UserID         int64 `json:"user_id,string"`
	FollowerCount  int64 `json:"follower_count,string"`
	FollowingCount int64 `json:"following_count,string"`
	IsFollowing    bool  `json:"is_following,omitempty"`
}

type FollowUserJSON struct {
	UserID      int64 `json:"user_id,string"`
	UserName    string `json:"user_name"`
	IsFollowing *bool  `json:"is_following,omitempty"`
}

type FollowUsersPageJSON struct {
	Users      []FollowUserJSON `json:"users"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

func (s *Server) GetUserFollowStats(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	st, err := s.Follows.Stats(c.Request.Context(), targetID, viewer)
	if err != nil {
		if errors.Is(err, appfollow.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, FollowStatsJSON{
		UserID:         targetID,
		FollowerCount:  st.FollowerCount,
		FollowingCount: st.FollowingCount,
		IsFollowing:    st.IsFollowing,
	})
}

func (s *Server) FollowUser(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	followeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Follows.Follow(c.Request.Context(), uid, followeeID)
	if err != nil {
		if errors.Is(err, domainfollow.ErrSelfFollow) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
			return
		}
		if errors.Is(err, domainfollow.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "already following"})
			return
		}
		if errors.Is(err, appfollow.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.notifyUserFollowed(followeeID, uid)
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) UnfollowUser(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	followeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Follows.Unfollow(c.Request.Context(), uid, followeeID)
	if err != nil {
		if errors.Is(err, domainfollow.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not following"})
			return
		}
		if errors.Is(err, domainfollow.ErrSelfFollow) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) ListMyFollowing(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	s.listFollowUsers(c, uid, true)
}

func (s *Server) ListMyFollowers(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	s.listFollowUsers(c, uid, false)
}

func (s *Server) ListUserFollowing(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	s.listFollowUsers(c, targetID, true)
}

func (s *Server) ListUserFollowers(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	s.listFollowUsers(c, targetID, false)
}

func (s *Server) listFollowUsers(c *gin.Context, targetUserID int64, following bool) {
	var cursor *domainfollow.PageCursor
	if v := c.Query("cursor"); v != "" {
		parsed, err := parseFollowListCursor(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		cursor = parsed
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	ctx := c.Request.Context()
	var entries []domainfollow.ListEntry
	var err error
	if following {
		entries, err = s.Follows.ListFollowing(ctx, targetUserID, cursor, limit)
	} else {
		entries, err = s.Follows.ListFollowers(ctx, targetUserID, cursor, limit)
	}
	if err != nil {
		if errors.Is(err, appfollow.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	userIDs := make([]int64, 0, len(entries))
	for _, e := range entries {
		userIDs = append(userIDs, e.UserID)
	}
	names := ResolveUserNames(ctx, s.Users, userIDs)
	out := make([]FollowUserJSON, 0, len(entries))
	for _, e := range entries {
		j := FollowUserJSON{UserID: e.UserID}
		if names != nil {
			j.UserName = names[e.UserID]
		}
		if viewer != nil && *viewer != e.UserID {
			ok, err := s.Follows.IsFollowing(ctx, *viewer, e.UserID)
			if err == nil {
				j.IsFollowing = &ok
			}
		}
		out = append(out, j)
	}
	var next *string
	if len(entries) == limit {
		s := formatFollowListCursor(entries[len(entries)-1])
		next = &s
	}
	c.JSON(http.StatusOK, FollowUsersPageJSON{Users: out, NextCursor: next})
}
