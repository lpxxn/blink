package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) ListNotifications(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Notifications == nil {
		c.JSON(http.StatusOK, NotificationsPageJSON{Notifications: []NotificationJSON{}})
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	list, err := s.Notifications.List(c.Request.Context(), uid, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]NotificationJSON, 0, len(list))
	for _, n := range list {
		out = append(out, NotificationToJSON(n))
	}
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, NotificationsPageJSON{Notifications: out, NextCursor: next})
}

func (s *Server) UnreadNotificationCount(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Notifications == nil {
		c.JSON(http.StatusOK, gin.H{"unread_count": "0"})
		return
	}
	n, err := s.Notifications.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": strconv.FormatInt(n, 10)})
}

func (s *Server) MarkNotificationRead(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.Notifications == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	err = s.Notifications.MarkRead(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) MarkAllNotificationsRead(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Notifications == nil {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	if err := s.Notifications.MarkAllRead(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
