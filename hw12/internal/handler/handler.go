package handler

import (
	"github.com/course-go/reelgoofy/internal/service"
)

// Handler holds the services that handlers will use.
type Handler struct {
	reviewService         service.Reviewer
	recommendationService service.Recommender
}

func NewHandler(reviewService service.Reviewer, recommendationService service.Recommender) *Handler {
	return &Handler{
		reviewService:         reviewService,
		recommendationService: recommendationService,
	}
}
