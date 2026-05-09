// Package cors provides Gin middleware for browser cross-origin access.
package cors

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware allows cross-origin requests. When the browser sends an Origin
// header, it is echoed back with Access-Control-Allow-Credentials so cookies
// (e.g. session) work from Flutter web / SPA dev servers. Preflight OPTIONS
// requests get 204 and standard CORS headers.
//
// Note: reflecting arbitrary Origins with credentials is appropriate for local
// development and trusted frontends; lock this down in production if the API
// is exposed on the public internet.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
