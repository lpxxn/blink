package httpidp

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lpxxn/blink/application/idp"
)

// Handler exposes the embedded OAuth2 authorization server (builtin IdP).
type Handler struct {
	Svc        *idp.Service
	FormAction string // e.g. /auth/idp/authorize
}

// Mount registers routes under the given group (e.g. r.Group("/auth/idp")).
func (h *Handler) Mount(rg *gin.RouterGroup) {
	rg.GET("/authorize", h.Authorize)
	rg.POST("/authorize", h.Authorize)
	rg.POST("/token", h.Token)
	rg.GET("/userinfo", h.UserInfo)
}

func (h *Handler) Authorize(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		h.authorizeGet(c.Writer, c.Request)
	case http.MethodPost:
		h.authorizePost(c.Writer, c.Request)
	default:
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func (h *Handler) authorizeGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	state := q.Get("state")
	if err := h.Svc.ValidateAuthorizeQuery(clientID, redirectURI, responseType); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = authorizeTmpl.Execute(w, map[string]string{
		"FormAction":   h.FormAction,
		"ClientID":     clientID,
		"RedirectURI":  redirectURI,
		"State":        state,
		"ResponseType": responseType,
	})
}

func (h *Handler) authorizePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	clientID := r.Form.Get("client_id")
	redirectURI := r.Form.Get("redirect_uri")
	state := r.Form.Get("state")
	responseType := r.Form.Get("response_type")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	loc, err := h.Svc.LoginWithPassword(r.Context(), clientID, redirectURI, state, email, password)
	if err != nil {
		// Re-render the login page with an error message instead of leaving the user on a blank 401 page.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_ = authorizeTmpl.Execute(w, map[string]string{
			"FormAction":   h.FormAction,
			"ClientID":     clientID,
			"RedirectURI":  redirectURI,
			"State":        state,
			"ResponseType": responseType,
			"Email":        email,
			"Error":        "登录失败：请检查邮箱/密码，或账号状态是否正常。",
		})
		return
	}
	http.Redirect(w, r, loc, http.StatusFound)
}

func (h *Handler) Token(c *gin.Context) {
	r := c.Request
	if r.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
	if err := r.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	grant := r.Form.Get("grant_type")
	code := r.Form.Get("code")
	redir := r.Form.Get("redirect_uri")
	cid := r.Form.Get("client_id")
	sec := r.Form.Get("client_secret")
	tok, exp, err := h.Svc.Exchange(r.Context(), grant, code, redir, cid, sec)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"access_token": tok,
		"token_type":   "Bearer",
		"expires_in":   exp,
	})
}

func (h *Handler) UserInfo(c *gin.Context) {
	r := c.Request
	auth := r.Header.Get("Authorization")
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		c.String(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	tok := strings.TrimSpace(parts[1])
	sub, email, name, err := h.Svc.UserInfo(r.Context(), tok)
	if err != nil {
		c.String(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"sub":   sub,
		"email": email,
		"name":  name,
	})
}

var authorizeTmpl = template.Must(template.New("authorize").Parse(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>登录</title></head>
<body>
<h1>登录</h1>
{{if .Error}}<p style="color:#b91c1c">{{.Error}}</p>{{end}}
<form method="POST" action="{{.FormAction}}">
  <input type="hidden" name="client_id" value="{{.ClientID}}">
  <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
  <input type="hidden" name="state" value="{{.State}}">
  <input type="hidden" name="response_type" value="{{.ResponseType}}">
  <p><label>邮箱 <input type="email" name="email" required autocomplete="username" value="{{.Email}}"></label></p>
  <p><label>密码 <input type="password" name="password" required autocomplete="current-password"></label></p>
  <p><button type="submit">继续</button></p>
</form>
</body>
</html>
`))
