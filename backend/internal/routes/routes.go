package routes

import (
	"database/sql"
	"net/http"

	"github.com/NeoPlays/simple-vod/backend/internal/handlers"
	"github.com/NeoPlays/simple-vod/backend/internal/middleware"
)

func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	h := &handlers.Handler{
		DB: database,
	}

	authed := func(next http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(database, next)
	}

	// Public
	mux.HandleFunc("POST /register", h.RegisterUser)
	mux.HandleFunc("POST /login", h.LoginUser)
	mux.HandleFunc("GET /health", h.Health)

	// Authenticated
	mux.Handle("POST /logout", authed(h.LogoutUser))
	mux.Handle("GET /user", authed(h.GetUser))
	mux.Handle("DELETE /user", authed(h.DeleteUser))

	mux.Handle("GET /videos", authed(h.ListVideos))
	mux.Handle("GET /videos/{id}", authed(h.StreamVideo))
	mux.Handle("POST /videos/sync", authed(h.SyncVideos))

	return middleware.Logging(mux)
}
