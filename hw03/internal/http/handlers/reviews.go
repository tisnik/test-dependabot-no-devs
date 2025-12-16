package handlers

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/reviews"
	"github.com/go-chi/chi/v5"
)

type ReviewsHandler struct {
	service *reviews.ReviewService
}

func NewReviewsHandler(service *reviews.ReviewService) *ReviewsHandler {
	return &ReviewsHandler{
		service: service,
	}
}

func (h *ReviewsHandler) IngestReviews(w http.ResponseWriter, r *http.Request) {
	var request dto.RawReviewsRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondWithError(w, apperrors.NewValidationError("body", err.Error()))
		return
	}

	result, err := h.service.Ingest(request.Data.Reviews)
	if err != nil {
		respondWithError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, result)
}

func (h *ReviewsHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewId := chi.URLParam(r, "reviewId")
	err := validateUUID("reviewId", reviewId)
	if err != nil {
		respondWithError(w, err)
		return
	}

	err = h.service.Delete(reviewId)
	if err != nil {
		respondWithError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.ReviewsDTO{})
}
