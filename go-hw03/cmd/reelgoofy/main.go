package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/course-go/reelgoofy/internal"
	"github.com/go-chi/chi/v5"
)

const (
	ReadTime       = 5 * time.Second
	WriteTime      = 10 * time.Second
	IdleTime       = 120 * time.Second
	ReadHeaderTime = 2 * time.Second
)

func main() {
	r := chi.NewRouter()
	app := internal.NewApp()
	app.Routing(r)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       ReadTime,
		WriteTimeout:      WriteTime,
		IdleTimeout:       IdleTime,
		ReadHeaderTimeout: ReadHeaderTime,
	}

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
