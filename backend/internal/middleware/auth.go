package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/NeoPlays/simple-vod/backend/internal/db"
)

type contextKey string

const userContextKey contextKey = "user"

func RequireAuth(dbConn *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := db.GetSessionByToken(r.Context(), dbConn, cookie.Value)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user, err := db.GetUserByID(r.Context(), dbConn, session.UserID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) *db.User {
	u, _ := ctx.Value(userContextKey).(*db.User)
	return u
}
