package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/http/controllers/recommendations"
	"github.com/course-go/reelgoofy/internal/http/controllers/reviews"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(
	recomController recommendations.Controller,
	reviewsController reviews.Controller,
) error {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/recommendations", func(r chi.Router) {
		r.Get("/content/{contentId}/content", recomController.GetContentToContent)
		r.Get("/users/{userId}/content", recomController.GetContentToUser)
	})
	router.Route("/reviews", func(r chi.Router) {
		r.Post("/", reviewsController.PostReview)
		r.Delete("/{reviewId}", reviewsController.DeleteReview)
	})

	slog.Info("Server is running")

	const maxTime = 5 * time.Second
	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  maxTime,
		WriteTimeout: maxTime,
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("Error with server", "err", err)
		return fmt.Errorf("error starting http server: %w", err)
	}
	return nil
}
