package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/course-go/reelgoofy/internal/api"
	"github.com/course-go/reelgoofy/internal/db"
	"github.com/go-chi/chi/v5"
)

const (
	defaultServerReadHeaderTimeout  = 2 * time.Second
	defaultServerIdleTimeoutTimeout = 30 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	database := db.CreateDb()
	r := chi.NewRouter()

	// Create your controller
	myAPI := &api.API{
		Database: database,
		Logger:   logger,
	}

	// Wrap it in the StrictHandler
	strictHandler := api.NewStrictHandler(myAPI, nil)

	// Register it with Chi using the Handler wrapper
	api.HandlerFromMuxWithBaseURL(strictHandler, r, "/api/v1")

	s := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: defaultServerReadHeaderTimeout,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       defaultServerIdleTimeoutTimeout,
	}

	err := s.ListenAndServe()
	if err != nil {
		myAPI.Logger.Error("Error with listening and serving port")
	}
}
