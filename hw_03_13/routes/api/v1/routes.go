package v1

import (
	"net/http"

	recommendationsHttp "github.com/course-go/reelgoofy/internal/containers/recommendations/http"
	reviewsHttp "github.com/course-go/reelgoofy/internal/containers/reviews/http"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/go-chi/chi/v5"
)

// RegisterApiRoutes registers API routes with Request Handlers.
func RegisterApiRoutes(router chi.Router, repo repository.ReviewRepository) {
	// Register API home URL
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.JSONResponse(w, http.StatusOK, nil, "hello from api!")
	})

	registerReviews(router, repo)
	registerRecommendations(router, repo)
}

// registerReviews registers reviews routes.
func registerReviews(router chi.Router, repo repository.ReviewRepository) {
	handler := reviewsHttp.NewReviewRequestHandler(repo)

	router.Route("/reviews", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/", handler.GetCollection)
		r.Delete("/{id}", handler.Delete)
	})
}

// registerRecommendations registers recommendations routes.
func registerRecommendations(router chi.Router, repo repository.ReviewRepository) {
	handler := recommendationsHttp.NewRecommendationRequestHandler(repo)

	router.Route("/recommendations", func(r chi.Router) {
		r.Get("/content/{contentId}/content", handler.GetContentRecommendations)
		r.Get("/users/{userId}/content", handler.GetUserRecommendations)
	})
}
