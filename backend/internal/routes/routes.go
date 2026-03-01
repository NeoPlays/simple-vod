package routes

import (
	"net/http"

	"database/sql"
	"github.com/NeoPlays/simple-vod/backend/internal/handlers"
	"github.com/NeoPlays/simple-vod/backend/internal/middleware"
)

func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	h := &handlers.Handler{
		DB: database,
	}

	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/videos", h.ListVideos)
	mux.HandleFunc("/videos/", h.StreamVideo)
	mux.HandleFunc("/videos/sync", h.SyncVideos)

	return middleware.Logging(mux)
}