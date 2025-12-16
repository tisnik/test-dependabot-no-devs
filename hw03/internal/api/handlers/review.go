package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	fieldsToValidate = 3
	dateLayout       = "2006-01-02"
)

func CreateReviewHandler(s *service.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reviews domain.RawReviewsRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		w.Header().Set("Content-Type", "application/json")
		err := decoder.Decode(&reviews)
		if err != nil {
			WriteFailResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		errors := validateJson(reviews)
		if len(errors) > 0 {
			WriteFailResponse(w, http.StatusBadRequest, errors)
			return
		}
		responseReviews, err := s.SaveReviews(reviews.Data.Reviews)
		if err != nil {
			WriteErrorResponse(w)
			return
		}
		response := domain.SuccessResponse[domain.ReviewsData]{
			Status: domain.StatusSuccess,
			Data: domain.ReviewsData{
				Reviews: responseReviews,
			},
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func DeleteReviewHandler(s *service.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "reviewId")
		w.Header().Set("Content-Type", "application/json")
		_, err := uuid.Parse(id)
		if id == "" || err != nil {
			WriteFailResponse(w, http.StatusBadRequest, map[string]string{
				"reviewId": "ID is not a valid UUID.",
			})
			return
		}
		errDel := s.DeleteReview(id)
		if errors.Is(errDel, service.ErrNotFound) {
			WriteFailResponse(w, http.StatusNotFound, map[string]string{"reviewId": "Review with such ID not found."})
			return
		}
		if errDel != nil {
			WriteErrorResponse(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := domain.SuccessResponse[any]{
			Status: "success",
			Data:   nil,
		}
		_ = json.NewEncoder(w).Encode(response)
	}
}

func validateJson(reviews domain.RawReviewsRequest) map[string]string {
	errors := make(map[string]string, fieldsToValidate)
	for _, review := range reviews.Data.Reviews {
		_, err1 := uuid.Parse(review.ContentID)
		if err1 != nil {
			errors["contentId"] = "ID is not a valid UUID."
		}
		_, err2 := uuid.Parse(review.UserID)
		if err2 != nil {
			errors["userId"] = "ID is not a valid UUID."
		}
		_, err3 := time.Parse(dateLayout, review.Released)
		if err3 != nil {
			errors["released"] = "Invalid date formats."
		}
	}
	return errors
}
