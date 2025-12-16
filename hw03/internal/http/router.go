package http

import (
	"github.com/course-go/reelgoofy/internal/http/handlers"
	"github.com/course-go/reelgoofy/internal/recommendations"
	"github.com/course-go/reelgoofy/internal/reviews"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	reviewService *reviews.ReviewService,
	recommendationService *recommendations.RecommendationService,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	reviewsHandler := handlers.NewReviewsHandler(reviewService)
	recommendationsHandler := handlers.NewRecommendationsHandler(recommendationService)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/reviews", ReviewRoutes(reviewsHandler))
		r.Mount("/recommendations", RecommendationRoutes(recommendationsHandler))
	})

	return r
}

func ReviewRoutes(reviewsHandler *handlers.ReviewsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", reviewsHandler.IngestReviews)
	r.Delete("/{reviewId}", reviewsHandler.DeleteReview)
	return r
}

func RecommendationRoutes(recommendationsHandler *handlers.RecommendationsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/content/{contentId}/content", recommendationsHandler.RecommendContentToContent)
	r.Get("/users/{userId}/content", recommendationsHandler.RecommendContentToUser)
	return r
}
