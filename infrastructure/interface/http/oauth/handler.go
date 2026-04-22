package httpoauth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appoauth "github.com/lpxxn/blink/application/oauth"
)

// Handler wires OAuth2 login/callback to LoginService.
type Handler struct {
	Svc *appoauth.LoginService
}

// Mount registers routes under the given group (e.g. r.Group("/auth/oauth")).
func (h *Handler) Mount(rg *gin.RouterGroup) {
	rg.GET("/:provider/login", h.Login)
	rg.GET("/:provider/callback", h.Callback)
}

func (h *Handler) Login(c *gin.Context) {
	provider := c.Param("provider")
	next := c.Query("next")
	url, _, err := h.Svc.LoginRedirectURL(c.Request.Context(), provider, next)
	if err != nil {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.String(http.StatusBadRequest, "missing code or state")
		return
	}
	req := c.Request
	token, next, err := h.Svc.CompleteLogin(c.Request.Context(), provider, code, state, req.RemoteAddr, req.UserAgent())
	if err != nil {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	maxAge := int(h.Svc.SessionTTL.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "blink_session",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.Redirect(http.StatusFound, next)
}
