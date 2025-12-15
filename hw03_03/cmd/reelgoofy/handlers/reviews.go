package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/services"
	"github.com/course-go/reelgoofy/cmd/reelgoofy/structs"
)

type ReviewHandler struct {
	service *services.ReviewService
}

func NewReviewHandler(service *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) IngestReviews(w http.ResponseWriter, r *http.Request) {
	var request structs.RawReviewsRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"body": "Invalid JSON format",
		})
		return
	}

	reviews, err := h.service.IngestReviews(request.Data.Reviews)
	if err != nil {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"reviews": err.Error(),
		})
		return
	}

	RespondSuccess(w, http.StatusCreated, reviews)
}

func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/reviews/")
	reviewID := strings.TrimSpace(path)

	if reviewID == "" {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"reviewId": "Review ID is required",
		})
		return
	}

	err := h.service.DeleteReview(reviewID)
	if err != nil {
		if err.Error() == "invalid UUID format" {
			RespondFail(w, http.StatusBadRequest, map[string]any{
				"reviewId": "ID is not a valid UUID.",
			})
			return
		}

		if err.Error() == "review not found" {
			RespondFail(w, http.StatusNotFound, map[string]any{
				"reviewId": "Review with such ID not found.",
			})
			return
		}

		RespondError(w, err.Error())
		return
	}

	RespondSuccess(w, http.StatusOK, nil)
}
