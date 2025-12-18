package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/course-go/reelgoofy/internal/domain"
	apierrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/handler/dto"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		service: service,
	}
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateReviewRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response := dto.FailResponse{
			Status: "fail",
			Data: map[string]string{
				"reviews": "Invalid JSON format.",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	err = validator.New().Struct(req)
	if err != nil {
		validationErrors := formatValidationErrors(err)

		response := dto.FailResponse{
			Status: "fail",
			Data:   validationErrors,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	reviews := make([]domain.Review, 0, len(req.Data.Reviews))

	reviewDTOs := req.Data.Reviews
	for _, reviewDTO := range reviewDTOs {
		review, err := domain.FromReviewDTO(reviewDTO)
		if err != nil {
			response := dto.ErrorResponse{
				Status:  "error",
				Message: "Error parsing json: " + err.Error(),
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
			return
		}
		reviews = append(reviews, review)
	}

	savedReviews, err := h.service.CreateReviews(reviews)
	if err != nil {
		response := dto.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	savedReviewDTOs := make([]dto.ReviewDTO, 0, len(savedReviews))

	for _, review := range savedReviews {
		savedReviewDTO := domain.ToReviewDTO(review)
		savedReviewDTOs = append(savedReviewDTOs, savedReviewDTO)
	}

	var response dto.CreateReviewResponse
	response.Status = "success"
	response.Data.Reviews = savedReviewDTOs

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response := dto.FailResponse{
			Status: "fail",
			Data: map[string]string{
				"reviewId": "Invalid review ID format.",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	err = h.service.DeleteReview(id)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			response := dto.FailResponse{
				Status: "fail",
				Data: map[string]string{
					"reviewId": "Review with such ID not found.",
				},
			}
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		} else {
			response := dto.ErrorResponse{
				Status:  "error",
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		}

		return
	}

	response := map[string]any{
		"status": "success",
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func formatValidationErrors(err error) map[string]string {
	errorsMap := make(map[string]string)

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fe := range validationErrs {
			field := fe.Field()

			switch fe.Tag() {
			case "uuid4", "uuid":
				errorsMap[field] = "ID is not a valid UUID."
			case "datetime":
				errorsMap[field] = "Invalid date formats."
			case "required":
				errorsMap[field] = "This field is required."
			default:
				errorsMap[field] = fmt.Sprintf("Validation failed on '%s' rule.", fe.Tag())
			}
		}
	}

	return errorsMap
}
