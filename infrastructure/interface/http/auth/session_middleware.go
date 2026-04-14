package httpauth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
)

const (
	// ContextUserIDKey is the gin.Context key set by RequireSession middleware.
	ContextUserIDKey = "blink_user_id"
)

// SessionTokenFromRequest extracts a Blink session token from an HTTP request.
//
// Priority:
//  1. Authorization: Bearer <token>
//  2. Cookie: blink_session=<token>
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

// OptionalSession sets ContextUserIDKey when a valid session is present and the user is active; otherwise continues without user.
// Banned or inactive accounts are treated as logged out; the session token is removed when users repository is provided.
func OptionalSession(store domainsession.Store, users domainuser.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}
		tok, ok := SessionTokenFromRequest(c.Request)
		if !ok {
			c.Next()
			return
		}
		sess, err := store.Get(c.Request.Context(), tok)
		if err != nil {
			c.Next()
			return
		}
		if users != nil {
			u, err := users.GetByID(c.Request.Context(), sess.UserID)
			if err != nil {
				if !errors.Is(err, domainuser.ErrNotFound) {
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
				c.Next()
				return
			}
			if u.Status != domainuser.StatusActive {
				_ = store.Delete(c.Request.Context(), tok)
				c.Next()
				return
			}
		}
		c.Set(ContextUserIDKey, sess.UserID)
		c.Next()
	}
}
