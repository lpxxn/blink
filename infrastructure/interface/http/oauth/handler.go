package httpoauth

import (
	"net/http"

	appoauth "github.com/lpxxn/blink/application/oauth"
	"github.com/go-chi/chi/v5"
)

// Handler wires OAuth2 login/callback to LoginService.
type Handler struct {
	Svc *appoauth.LoginService
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{provider}/login", h.Login)
	r.Get("/{provider}/callback", h.Callback)
	return r
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	next := r.URL.Query().Get("next")
	url, _, err := h.Svc.LoginRedirectURL(r.Context(), provider, next)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}
	ip := r.RemoteAddr
	ua := r.UserAgent()
	token, next, err := h.Svc.CompleteLogin(r.Context(), provider, code, state, ip, ua)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	maxAge := int(h.Svc.SessionTTL.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "blink_session",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, next, http.StatusFound)
}
