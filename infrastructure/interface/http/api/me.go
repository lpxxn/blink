package httpapi

import (
	"errors"
	"net/http"
	"strings"

	domainuser "github.com/lpxxn/blink/domain/user"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/gin-gonic/gin"
)

const maxProfileNameLen = 80

// GetMe returns the authenticated user's id and public profile fields.
func (s *Server) GetMe(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Users == nil {
		c.JSON(http.StatusOK, MeJSON{UserID: uid})
		return
	}
	u, err := s.Users.GetByID(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, MeJSON{
		UserID: u.SnowflakeID,
		Email:  u.Email,
		Name:   u.Name,
		Role:   u.Role,
		Status: u.Status,
	})
}

type patchMeBody struct {
	Name *string `json:"name"`
}

// PatchMe updates the current user's profile (display name).
func (s *Server) PatchMe(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Users == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var body patchMeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing name"})
		return
	}
	name := strings.TrimSpace(*body.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昵称不能为空"})
		return
	}
	if len(name) > maxProfileNameLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昵称过长"})
		return
	}
	if err := s.Users.UpdateName(c.Request.Context(), uid, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u, err := s.Users.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusOK, MeJSON{UserID: uid, Name: name})
		return
	}
	c.JSON(http.StatusOK, MeJSON{
		UserID: u.SnowflakeID,
		Email:  u.Email,
		Name:   u.Name,
		Role:   u.Role,
		Status: u.Status,
	})
}

// Logout deletes the current session (if any) and clears the session cookie.
func (s *Server) Logout(c *gin.Context) {
	if s.Sessions != nil {
		if tok, ok := httpauth.SessionTokenFromRequest(c.Request); ok && tok != "" {
			_ = s.Sessions.Delete(c.Request.Context(), tok)
		}
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "blink_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.AbortWithStatus(http.StatusNoContent)
}
