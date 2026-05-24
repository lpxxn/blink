package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apppost "github.com/lpxxn/blink/application/post"
	apppostlike "github.com/lpxxn/blink/application/postlike"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostlike "github.com/lpxxn/blink/domain/postlike"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) LikePost(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Likes.Like(c.Request.Context(), uid, postID)
	if err != nil {
		if errors.Is(err, apppostlike.ErrPostNotFound) || errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, domainpostlike.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{"error": "already liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s.NotifyEvents != nil && s.Posts != nil {
		post, perr := s.Posts.GetByID(c.Request.Context(), postID)
		if perr == nil && post != nil && post.UserID != uid {
			_ = s.NotifyEvents.PublishPostLiked(c.Request.Context(), post.UserID, postID, uid)
		}
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) UnlikePost(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Likes.Unlike(c.Request.Context(), uid, postID)
	if err != nil {
		if errors.Is(err, domainpostlike.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) GetPostLikes(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if _, err := s.Posts.GetPublic(c.Request.Context(), postID); err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppost.ErrNotVisible) {
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
	meta, err := s.Likes.MetaForPost(c.Request.Context(), postID, viewer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := PostLikeJSON{PostID: postID, LikeCount: meta.LikeCount}
	if viewer != nil {
		out.Liked = meta.Liked
	}
	c.JSON(http.StatusOK, out)
}
