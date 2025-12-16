package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Markek1/reelgoofy/internal/api"
	"github.com/Markek1/reelgoofy/internal/repository"
	"github.com/Markek1/reelgoofy/internal/service"
)

const serverTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	repo := repository.NewMemoryRepository()
	svc := service.NewService(repo, logger)
	router := api.NewRouter(svc, logger)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
	}

	logger.Info("Starting server", "addr", ":8080")
	err := server.ListenAndServe()
	if err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
