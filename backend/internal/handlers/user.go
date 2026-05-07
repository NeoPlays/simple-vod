package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/NeoPlays/simple-vod/backend/internal/config"
	"github.com/NeoPlays/simple-vod/backend/internal/db"
	"github.com/NeoPlays/simple-vod/backend/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(config.GetPasswordPepper()+password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(config.GetPasswordPepper()+password))
	return err == nil
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username          string `json:"username"`
		Password          string `json:"password"`
		RegistrationToken string `json:"registration_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	count, err := db.UserCount(r.Context(), h.DB)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	role := db.RoleUser
	if count == 0 {
		role = db.RoleAdmin
	} else {
		if req.RegistrationToken == "" {
			http.Error(w, "Registration token required", http.StatusForbidden)
			return
		}
		if err := db.ConsumeRegistrationToken(r.Context(), h.DB, req.RegistrationToken); err != nil {
			http.Error(w, "Invalid registration token", http.StatusForbidden)
			return
		}
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	u := &db.User{
		Name:         req.Username,
		PasswordHash: hashedPassword,
		Role:         role,
	}
	if err := db.CreateUser(r.Context(), h.DB, u); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	u.PasswordHash = ""
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	u, err := db.GetUserByName(r.Context(), h.DB, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !checkPassword(u.PasswordHash, req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, err := db.CreateSession(r.Context(), h.DB, u.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	u.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	db.DeleteSession(r.Context(), h.DB, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	u.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := db.DeleteUser(r.Context(), h.DB, u.ID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	w.WriteHeader(http.StatusNoContent)
}
