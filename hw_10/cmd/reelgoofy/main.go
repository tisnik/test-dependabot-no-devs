package main

import (
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/handler"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const readHeaderTimeout = 3 * time.Second

func main() {
	reviewRepository := repository.NewReviewRepository()
	reviewService := service.NewReviewService(reviewRepository)
	recommendationService := service.NewRecommendationService(reviewRepository)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))

	r.Route("/api/v1", func(r chi.Router) {
		reviewHandler := handler.NewReviewHandler(reviewService)
		recommendationsHandler := handler.NewRecommendationsHandler(recommendationService)

		r.Route("/reviews", func(r chi.Router) {
			r.Post("/", reviewHandler.CreateReview)
			r.Delete("/{id}", reviewHandler.DeleteReview)
		})

		r.Route("/recommendations", func(r chi.Router) {
			r.Get("/users/{userId}/content", recommendationsHandler.GetRecommendationsByUser)
			r.Get("/content/{contentId}/content", recommendationsHandler.GetRecommendationsByContent)
		})
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	_ = server.ListenAndServe()
}
