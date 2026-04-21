package httpauth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lpxxn/blink/application/auth"
	"github.com/lpxxn/blink/application/emailcode"
)

// PasswordResetHandler serves the unauthenticated forgot-password flow.
type PasswordResetHandler struct {
	Svc *auth.PasswordService
}

type pwResetSendBody struct {
	Email string `json:"email"`
}

// SendCode requests a reset-password code. Always returns 200 to avoid account
// enumeration; internally refuses to send when the email is unknown.
func (h *PasswordResetHandler) SendCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var body pwResetSendBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := h.Svc.SendResetCode(r.Context(), body.Email, clientIPKey(r)); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			http.Error(w, "invalid email", http.StatusBadRequest)
		case errors.Is(err, emailcode.ErrCoolingDown), errors.Is(err, emailcode.ErrTooMany):
			http.Error(w, err.Error(), http.StatusTooManyRequests)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	respondOK(w)
}

type pwResetBody struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// Reset verifies the code and overwrites the password hash.
func (h *PasswordResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var body pwResetBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := h.Svc.ResetPassword(r.Context(), body.Email, body.Code, body.NewPassword); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			http.Error(w, "invalid email", http.StatusBadRequest)
		case errors.Is(err, auth.ErrInvalidCode):
			http.Error(w, "invalid or expired verification code", http.StatusBadRequest)
		case errors.Is(err, auth.ErrWeakPassword):
			http.Error(w, "password too short", http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
