package httpadmin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appadmin "github.com/lpxxn/blink/application/admin"
)

type registerEmailVerificationJSON struct {
	Required bool `json:"required"`
}

func (s *Server) GetRegisterEmailVerificationRequired(c *gin.Context) {
	required, err := s.Admin.GetRegisterEmailVerificationRequired(c.Request.Context())
	if err != nil {
		if errors.Is(err, appadmin.ErrSettingsNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registerEmailVerificationJSON{Required: required})
}

func (s *Server) SetRegisterEmailVerificationRequired(c *gin.Context) {
	var body registerEmailVerificationJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.Admin.SetRegisterEmailVerificationRequired(c.Request.Context(), body.Required); err != nil {
		if errors.Is(err, appadmin.ErrSettingsNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registerEmailVerificationJSON{Required: body.Required})
}
