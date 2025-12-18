package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/course-go/reelgoofy/internal/handler"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	readHeaderTimeout = 5 * time.Second
)

// Run starts the ReelGoofy service.
func Run(ctx context.Context) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("starting reelgoofy service")

	memStorage := storage.NewMemoryStorage()
	reviewService := service.NewReviewService(memStorage)
	recommendationService := service.NewRecommendationService(memStorage)
	apiHandler := handler.NewHandler(reviewService, recommendationService)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Post("/reviews", apiHandler.CreateReview)
	r.Get("/recommendations/{userID}", apiHandler.GetRecommendations)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server is listening on", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	err := <-serverErrors
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info("server stopped")

	return nil
}
