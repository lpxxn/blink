package httpauth

import (
	"encoding/json"
	"net/http"
)

type RegisterConfigHandler struct {
	Settings registerEmailVerificationSettings
}

type registerConfigJSON struct {
	EmailVerificationRequired bool `json:"email_verification_required"`
}

func (h *RegisterConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if h.Settings == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	required, err := h.Settings.GetRegisterEmailVerificationRequired(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(registerConfigJSON{EmailVerificationRequired: required})
}
