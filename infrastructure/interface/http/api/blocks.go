package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appblock "github.com/lpxxn/blink/application/block"
	domainblock "github.com/lpxxn/blink/domain/block"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) BlockUser(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	blockedID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.Blocks == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "blocks unavailable"})
		return
	}
	err = s.Blocks.Block(c.Request.Context(), uid, blockedID)
	if err != nil {
		if errors.Is(err, domainblock.ErrSelfBlock) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
			return
		}
		if errors.Is(err, domainblock.ErrAlreadyBlocked) {
			c.JSON(http.StatusConflict, gin.H{"error": "already blocked"})
			return
		}
		if errors.Is(err, appblock.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) UnblockUser(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	blockedID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if s.Blocks == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "blocks unavailable"})
		return
	}
	err = s.Blocks.Unblock(c.Request.Context(), uid, blockedID)
	if err != nil {
		if errors.Is(err, domainblock.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not blocked"})
			return
		}
		if errors.Is(err, domainblock.ErrSelfBlock) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) ListMyBlocked(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Blocks == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "blocks unavailable"})
		return
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := s.Blocks.ListBlocked(c.Request.Context(), uid, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]PublicUserJSON, 0, len(list))
	for _, u := range list {
		out = append(out, PublicUserJSON{UserID: u.SnowflakeID, Name: u.Name})
	}
	resp := BlockedUsersPageJSON{Users: out}
	if len(list) == clampBlockListLimit(limit) {
		next := offset + len(list)
		resp.NextOffset = &next
	}
	c.JSON(http.StatusOK, resp)
}

func clampBlockListLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

func (s *Server) isEitherBlocked(c *gin.Context, viewerID, targetID int64) bool {
	if s.Blocks == nil || viewerID == 0 || targetID == 0 || viewerID == targetID {
		return false
	}
	blocked, err := s.Blocks.IsEitherBlocked(c.Request.Context(), viewerID, targetID)
	return err == nil && blocked
}
