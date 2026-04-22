package httpauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lpxxn/blink/application/auth"
)

type registerEmailVerificationSettings interface {
	GetRegisterEmailVerificationRequired(context.Context) (bool, error)
}

type RegisterHandler struct {
	Svc      *auth.RegisterService
	Settings registerEmailVerificationSettings
}

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Code     string `json:"code"`
}

func (h *RegisterHandler) canRegisterWithSession() bool {
	s := h.Svc
	return s != nil && s.Tx != nil && s.Sessions != nil && s.SessionTTL > 0
}

func (h *RegisterHandler) emailCodeRequired(ctx context.Context) (bool, error) {
	if h.Settings == nil {
		return true, nil
	}
	required, err := h.Settings.GetRegisterEmailVerificationRequired(ctx)
	if err != nil {
		return false, err
	}
	return required, nil
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
	required, err := h.emailCodeRequired(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	switch {
	case required && h.canRegisterWithSession():
		uid, sessionToken, err = h.Svc.RegisterWithSessionVerified(ctx, body.Email, body.Password, body.Name, body.Code, r.RemoteAddr, r.UserAgent())
	case required:
		uid, err = h.Svc.RegisterWithPasswordVerified(ctx, body.Email, body.Password, body.Name, body.Code)
	case h.canRegisterWithSession():
		uid, sessionToken, err = h.Svc.RegisterWithSession(ctx, body.Email, body.Password, body.Name, r.RemoteAddr, r.UserAgent())
	default:
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
		case errors.Is(err, auth.ErrInvalidCode):
			http.Error(w, "invalid or expired verification code", http.StatusBadRequest)
		case errors.Is(err, auth.ErrCodesNotConfigured):
			http.Error(w, "email verification required but not configured", http.StatusServiceUnavailable)
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
