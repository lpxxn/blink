package httpapi

import (
	"net/http"

	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/gin-gonic/gin"
)

// GetMe returns the authenticated user's snowflake id.
func (s *Server) GetMe(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, MeJSON{UserID: uid})
}
