package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appsearch "github.com/lpxxn/blink/application/search"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) ListFollowingPosts(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var beforeID *int64
	if v := c.Query("cursor"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		beforeID = &id
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := s.Posts.ListFollowingFeed(c.Request.Context(), uid, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	viewer := uid
	out := s.postsToJSON(c.Request.Context(), list, &viewer)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, PostsPageJSON{Posts: out, NextCursor: next})
}

func (s *Server) SearchPosts(c *gin.Context) {
	if s.Search == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search unavailable"})
		return
	}
	q := c.Query("q")
	var beforeID *int64
	if v := c.Query("cursor"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		beforeID = &id
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := s.Search.SearchPosts(c.Request.Context(), q, beforeID, limit)
	if err != nil {
		if errors.Is(err, appsearch.ErrInvalidQuery) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	out := s.postsToJSON(c.Request.Context(), list, viewer)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, PostsPageJSON{Posts: out, NextCursor: next})
}

func (s *Server) SearchUsers(c *gin.Context) {
	if s.Search == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search unavailable"})
		return
	}
	q := c.Query("q")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := s.Search.SearchUsers(c.Request.Context(), q, offset, limit)
	if err != nil {
		if errors.Is(err, appsearch.ErrInvalidQuery) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]PublicUserJSON, 0, len(list))
	for _, u := range list {
		out = append(out, PublicUserJSON{UserID: u.SnowflakeID, Name: u.Name})
	}
	resp := UsersSearchPageJSON{Users: out}
	if len(list) == clampSearchLimit(limit) {
		next := offset + len(list)
		resp.NextOffset = &next
	}
	c.JSON(http.StatusOK, resp)
}

func clampSearchLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 20
	}
	return limit
}
