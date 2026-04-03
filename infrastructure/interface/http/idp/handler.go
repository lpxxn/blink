package httpidp

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/lpxxn/blink/application/idp"
)

// Handler exposes the embedded OAuth2 authorization server (builtin IdP).
type Handler struct {
	Svc        *idp.Service
	FormAction string // e.g. /auth/idp/authorize
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/authorize", h.Authorize)
	r.Post("/authorize", h.Authorize)
	r.Post("/token", h.Token)
	r.Get("/userinfo", h.UserInfo)
	return r
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.authorizeGet(w, r)
	case http.MethodPost:
		h.authorizePost(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
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
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	loc, err := h.Svc.LoginWithPassword(r.Context(), clientID, redirectURI, state, email, password)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, loc, http.StatusFound)
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	grant := r.Form.Get("grant_type")
	code := r.Form.Get("code")
	redir := r.Form.Get("redirect_uri")
	cid := r.Form.Get("client_id")
	sec := r.Form.Get("client_secret")
	tok, exp, err := h.Svc.Exchange(r.Context(), grant, code, redir, cid, sec)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": tok,
		"token_type":   "Bearer",
		"expires_in":   exp,
	})
}

func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	tok := strings.TrimSpace(parts[1])
	sub, email, name, err := h.Svc.UserInfo(r.Context(), tok)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
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
<form method="POST" action="{{.FormAction}}">
  <input type="hidden" name="client_id" value="{{.ClientID}}">
  <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
  <input type="hidden" name="state" value="{{.State}}">
  <input type="hidden" name="response_type" value="{{.ResponseType}}">
  <p><label>邮箱 <input type="email" name="email" required autocomplete="username"></label></p>
  <p><label>密码 <input type="password" name="password" required autocomplete="current-password"></label></p>
  <p><button type="submit">继续</button></p>
</form>
</body>
</html>
`))
