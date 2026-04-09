package httpauth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	domainuser "github.com/lpxxn/blink/domain/user"
)

// RequireUserRole loads the user by session id and requires one of the allowed roles.
func RequireUserRole(users domainuser.Repository, allowed ...string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allow[a] = struct{}{}
	}
	return func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if users == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		u, err := users.GetByID(c.Request.Context(), uid)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if _, ok := allow[u.Role]; !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
