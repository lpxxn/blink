package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lpxxn/blink/application/auth"
	"github.com/lpxxn/blink/application/emailcode"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

// SendChangePasswordCode POSTs a fresh verification code to the user's own email.
func (s *Server) SendChangePasswordCode(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Passwords == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if err := s.Passwords.SendChangeCode(c.Request.Context(), uid); err != nil {
		switch {
		case errors.Is(err, emailcode.ErrCoolingDown), errors.Is(err, emailcode.ErrTooMany):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type changePasswordBody struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// ChangePassword verifies the code, rotates the password hash, and revokes all
// active sessions for this user. The current request's cookie becomes invalid
// immediately; the client should navigate to /login.
func (s *Server) ChangePassword(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Passwords == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var body changePasswordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.Passwords.ChangePassword(c.Request.Context(), uid, body.Code, body.NewPassword); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification code"})
		case errors.Is(err, auth.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	// drop this request's cookie
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
