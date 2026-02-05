package main

import (
	"log/slog"
	"net/http"
	"time"

	controller "github.com/course-go/reelgoofy/internal/controller"
	reviews "github.com/course-go/reelgoofy/internal/repository"
)

const (
	defaultPort               = "8081"
	defaultServerReadTimeout  = 3 * time.Second
	defaultServerWriteTimeout = 3 * time.Second
	defaultServerIdleTimeout  = 60 * time.Second
)

func main() {
	repository := reviews.NewRepository()
	router := controller.NewRouter(repository)

	hostname := ":" + defaultPort
	slog.Info("starting server", "hostname", hostname)
	server := &http.Server{
		Addr:         hostname,
		Handler:      router,
		ReadTimeout:  defaultServerReadTimeout,
		WriteTimeout: defaultServerWriteTimeout,
		IdleTimeout:  defaultServerIdleTimeout,
	}
	err := server.ListenAndServe()
	if err != nil {
		slog.Error("failed running server", "error", err)
	}
}
