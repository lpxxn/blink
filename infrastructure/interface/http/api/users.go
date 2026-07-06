package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	domainuser "github.com/lpxxn/blink/domain/user"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func parsePathUserID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return 0, false
	}
	return id, true
}

func publicUserVisible(u *domainuser.User) bool {
	return u != nil && u.Status == domainuser.StatusActive
}

func (s *Server) loadPublicUser(c *gin.Context, userID int64) (*domainuser.User, bool) {
	if s.Users == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users unavailable"})
		return nil, false
	}
	u, err := s.Users.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return nil, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	if !publicUserVisible(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return nil, false
	}
	return u, true
}

func (s *Server) GetUser(c *gin.Context) {
	userID, ok := parsePathUserID(c)
	if !ok {
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	if viewer != nil && s.isEitherBlocked(c, *viewer, userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	u, ok := s.loadPublicUser(c, userID)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, PublicUserJSON{
		UserID: u.SnowflakeID,
		Name:   u.Name,
	})
}

func (s *Server) ListUserPosts(c *gin.Context) {
	userID, ok := parsePathUserID(c)
	if !ok {
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	if viewer != nil && s.isEitherBlocked(c, *viewer, userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if _, ok := s.loadPublicUser(c, userID); !ok {
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
	list, err := s.Posts.ListPublicByUser(c.Request.Context(), userID, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := s.postsToJSON(c.Request.Context(), list, viewer)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, PostsPageJSON{Posts: out, NextCursor: next})
}
