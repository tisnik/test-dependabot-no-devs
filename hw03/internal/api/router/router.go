package router

import (
	"net/http"

	"github.com/course-go/reelgoofy/internal/api/handlers"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewRouter(s *service.ReviewService) http.Handler {
	r := chi.NewRouter()

	r.Post("/reviews", handlers.CreateReviewHandler(s))
	r.Delete("/reviews/{reviewId}", handlers.DeleteReviewHandler(s))

	r.Get("/recommendations/content/{contentId}/content", handlers.RecommendContentToContentHandler(s))
	r.Get("/recommendations/users/{userId}/content", handlers.RecommendContentToUserHandler(s))

	return r
}
