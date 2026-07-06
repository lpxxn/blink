package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appreplylike "github.com/lpxxn/blink/application/replylike"
	domainreplylike "github.com/lpxxn/blink/domain/replylike"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) LikeReply(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	replyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.ReplyLikes == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reply likes unavailable"})
		return
	}
	err = s.ReplyLikes.Like(c.Request.Context(), uid, replyID)
	if err != nil {
		if errors.Is(err, appreplylike.ErrReplyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, domainreplylike.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{"error": "already liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.notifyReplyLiked(replyID, uid)
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) UnlikeReply(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	replyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.ReplyLikes == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reply likes unavailable"})
		return
	}
	err = s.ReplyLikes.Unlike(c.Request.Context(), uid, replyID)
	if err != nil {
		if errors.Is(err, domainreplylike.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) GetReplyLikes(c *gin.Context) {
	replyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.ReplyLikes == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reply likes unavailable"})
		return
	}
	if err := s.ReplyLikes.VerifyPublic(c.Request.Context(), replyID); err != nil {
		if errors.Is(err, appreplylike.ErrReplyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	meta, err := s.ReplyLikes.MetaForReply(c.Request.Context(), replyID, viewer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := ReplyLikeJSON{ReplyID: replyID, LikeCount: meta.LikeCount}
	if viewer != nil {
		out.Liked = meta.Liked
	}
	c.JSON(http.StatusOK, out)
}
