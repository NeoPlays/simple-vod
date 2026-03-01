package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NeoPlays/simple-vod/backend/internal/db"
	"github.com/NeoPlays/simple-vod/backend/internal/routes"
)


func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.InitDB("./db/sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	initCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.InitSchema(initCtx, database); err != nil {
		log.Fatal(err)
	}

	mux := routes.New(database)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Listening on :8080")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
