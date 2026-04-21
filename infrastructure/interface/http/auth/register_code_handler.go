package httpauth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lpxxn/blink/application/emailcode"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// RegisterCodeHandler serves POST /auth/register/send_code. It asks the email
// code service to send a register-purpose verification code, but refuses to
// actually send (while still returning 200) when the address is already
// registered, so the endpoint cannot be used to enumerate accounts.
type RegisterCodeHandler struct {
	Codes *emailcode.Service
	Users domainuser.Repository
}

type registerCodeBody struct {
	Email string `json:"email"`
}

func (h *RegisterCodeHandler) Send(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var body registerCodeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	email := emailcode.NormalizeEmail(body.Email)
	if email == "" {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if h.Users != nil {
		if _, err := h.Users.FindByEmail(ctx, email); err == nil {
			respondOK(w)
			return
		} else if !errors.Is(err, domainuser.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	if err := h.Codes.Send(ctx, emailcode.PurposeRegister, email, clientIPKey(r)); err != nil {
		switch {
		case errors.Is(err, emailcode.ErrInvalidEmail):
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

func respondOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// clientIPKey extracts an ip-ish key for rate limiting. Uses X-Forwarded-For
// first (trusting the reverse proxy, consistent with the rest of the app) and
// falls back to RemoteAddr.
func clientIPKey(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return v
	}
	return r.RemoteAddr
}
