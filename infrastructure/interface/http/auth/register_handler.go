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
	id, err := h.Svc.RegisterWithPassword(r.Context(), body.Email, body.Password, body.Name)
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
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int64{"user_id": id})
}
