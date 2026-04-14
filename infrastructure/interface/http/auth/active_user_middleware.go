package httpauth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// RequireActiveUser rejects the request if the authenticated user is not StatusActive.
// It runs after RequireSession. Non-active users get401; the current session token is deleted when possible.
func RequireActiveUser(store domainsession.Store, users domainuser.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		u, err := users.GetByID(c.Request.Context(), uid)
		if err != nil {
			if errors.Is(err, domainuser.ErrNotFound) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if u.Status != domainuser.StatusActive {
			if store != nil {
				if tok, ok := SessionTokenFromRequest(c.Request); ok {
					_ = store.Delete(c.Request.Context(), tok)
				}
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
