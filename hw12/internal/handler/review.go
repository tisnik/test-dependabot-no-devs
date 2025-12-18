package handler

import (
	"encoding/json"
	"net/http"

	"github.com/course-go/reelgoofy/internal/model"
)

// CreateReview handles the ingestion of a new review.
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var review model.Review
	err := json.NewDecoder(r.Body).Decode(&review)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.reviewService.Create(r.Context(), review)
	if err != nil {
		http.Error(w, "failed to create review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
