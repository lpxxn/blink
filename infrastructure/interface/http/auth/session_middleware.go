package httpauth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domainsession "github.com/lpxxn/blink/domain/session"
)

const (
	// ContextUserIDKey is the gin.Context key set by RequireSession middleware.
	ContextUserIDKey = "blink_user_id"
)

// SessionTokenFromRequest extracts a Blink session token from an HTTP request.
//
// Priority:
//  1) Authorization: Bearer <token>
//  2) Cookie: blink_session=<token>
func SessionTokenFromRequest(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	if tok, ok := bearerToken(r.Header.Get("Authorization")); ok {
		return tok, true
	}
	if c, err := r.Cookie("blink_session"); err == nil && c != nil && c.Value != "" {
		return c.Value, true
	}
	return "", false
}

func bearerToken(h string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(h))
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	if parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

// RequireSession authenticates requests using either Authorization Bearer token or blink_session cookie.
// On success it stores the user id in gin context under ContextUserIDKey.
func RequireSession(store domainsession.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		tok, ok := SessionTokenFromRequest(c.Request)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		sess, err := store.Get(c.Request.Context(), tok)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(ContextUserIDKey, sess.UserID)
		c.Next()
	}
}

// UserIDFromContext fetches user id set by RequireSession.
func UserIDFromContext(c *gin.Context) (int64, bool) {
	if c == nil {
		return 0, false
	}
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
