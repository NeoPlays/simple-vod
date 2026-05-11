package routes

import (
	"database/sql"
	"net/http"

	"github.com/NeoPlays/simple-vod/backend/internal/config"
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

	admin := func(next http.HandlerFunc) http.Handler {
		return middleware.RequireAdmin(database, next)
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
	mux.Handle("GET /videos/{id}/thumb", authed(h.ServeThumb))
	mux.Handle("POST /videos/sync", authed(h.SyncVideos))

	// Admin
	mux.Handle("POST /admin/tokens", admin(h.CreateRegistrationToken))
	mux.Handle("GET /admin/tokens", admin(h.ListRegistrationTokens))
	mux.Handle("DELETE /admin/tokens/{token}", admin(h.RevokeRegistrationToken))

	// Serve the frontend as static files, catches anything not matched above.
	fs := http.FileServer(http.Dir(config.GetFrontendDirectory()))
	mux.Handle("/", fs)

	return middleware.Logging(mux)
}
