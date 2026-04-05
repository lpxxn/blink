package httpauth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lpxxn/blink/application/auth"
)

type RegisterHandler struct {
	Svc *auth.RegisterService
}

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *RegisterHandler) canRegisterWithSession() bool {
	s := h.Svc
	return s != nil && s.Tx != nil && s.Sessions != nil && s.SessionTTL > 0
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var body registerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var uid int64
	var sessionToken string
	var err error
	if h.canRegisterWithSession() {
		uid, sessionToken, err = h.Svc.RegisterWithSession(ctx, body.Email, body.Password, body.Name, r.RemoteAddr, r.UserAgent())
	} else {
		uid, err = h.Svc.RegisterWithPassword(ctx, body.Email, body.Password, body.Name)
	}
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailTaken):
			http.Error(w, "email already registered", http.StatusConflict)
		case errors.Is(err, auth.ErrWeakPassword):
			http.Error(w, "password too short", http.StatusBadRequest)
		case errors.Is(err, auth.ErrInvalidEmail):
			http.Error(w, "invalid email", http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if sessionToken != "" {
		maxAge := int(h.Svc.SessionTTL.Seconds())
		if maxAge < 0 {
			maxAge = 0
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "blink_session",
			Value:    sessionToken,
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	w.WriteHeader(http.StatusCreated)
	if sessionToken != "" {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":        uid,
			"session_token":  sessionToken,
			"session_cookie": "blink_session",
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int64{"user_id": uid})
}
