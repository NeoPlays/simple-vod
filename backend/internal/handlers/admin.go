package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/NeoPlays/simple-vod/backend/internal/db"
	"github.com/NeoPlays/simple-vod/backend/internal/middleware"
)

func (h *Handler) CreateRegistrationToken(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())

	token, err := db.CreateRegistrationToken(r.Context(), h.DB, u.ID)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (h *Handler) ListRegistrationTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := db.ListRegistrationTokens(r.Context(), h.DB)
	if err != nil {
		http.Error(w, "Failed to list tokens", http.StatusInternalServerError)
		return
	}
	if tokens == nil {
		tokens = []db.RegistrationToken{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) RevokeRegistrationToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	if err := db.ConsumeRegistrationToken(r.Context(), h.DB, token); err != nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
