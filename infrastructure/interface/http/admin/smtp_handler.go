package httpadmin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appadmin "github.com/lpxxn/blink/application/admin"
)

func (s *Server) GetSMTPSettings(c *gin.Context) {
	if s.SMTP == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	v, err := s.SMTP.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (s *Server) PutSMTPSettings(c *gin.Context) {
	if s.SMTP == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var body appadmin.SMTPConfigUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.SMTP.Set(c.Request.Context(), body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := s.SMTP.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

type testSMTPBody struct {
	To string `json:"to"`
}

func (s *Server) TestSMTP(c *gin.Context) {
	if s.SMTP == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var body testSMTPBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.SMTP.Test(c.Request.Context(), body.To); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
