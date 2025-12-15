package reviews

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/request"
	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/response"
	reviewsService "github.com/course-go/reelgoofy/internal/http/services/reviews"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Controller struct {
	service   reviewsService.Service
	validator *validator.Validate
}

func NewController(service reviewsService.Service, validator *validator.Validate) Controller {
	return Controller{
		service:   service,
		validator: validator,
	}
}

func (c Controller) PostReview(w http.ResponseWriter, r *http.Request) {
	slog.Info("Create review called")
	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading body", "error", err)
		code := http.StatusInternalServerError
		w.WriteHeader(code)
		_, _ = w.Write(response.NewReviewsErrorResponse(code))

		return
	}

	var newReviews request.RawReviewsRequest
	err = json.Unmarshal(body, &newReviews)
	if err != nil {
		slog.Error("Error unmarshaling body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		data := make(map[string]string)
		data["error"] = err.Error()
		_, _ = w.Write(response.NewReviewsFailResponse(data))

		return
	}

	err = c.validator.Struct(newReviews)
	if err != nil {
		slog.Error("Error validating body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			data := make(map[string]string)
			for _, err := range validationErrors {
				data[err.Field()] = err.ActualTag()
			}
			_, _ = w.Write(response.NewReviewsFailResponse(data))
		} else {
			_, _ = w.Write(response.NewReviewsFailResponse(map[string]string{"error": "invalid request"}))
		}
		return
	}

	savedReviews := c.service.CreateReviews(newReviews)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(response.NewReviewsSuccessResponsePost(savedReviews))
}

func (c Controller) DeleteReview(w http.ResponseWriter, r *http.Request) {
	slog.Info("Delete review called")

	id := chi.URLParam(r, "reviewId")

	err := c.validator.Struct(request.DeleteReviewRequest{ID: id})
	if err != nil {
		slog.Error("Error validating body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			data := make(map[string]string)
			for _, err := range validationErrors {
				data[err.Field()] = err.ActualTag()
			}
			_, _ = w.Write(response.NewReviewsFailResponse(data))
		} else {
			_, _ = w.Write(response.NewReviewsFailResponse(map[string]string{"error": "invalid request"}))
		}
		return
	}

	err = c.service.DeleteReview(id)
	if err != nil {
		slog.Error("Error deleting review", "error", err)
		w.WriteHeader(http.StatusNotFound)
		data := make(map[string]string)
		data["error"] = err.Error()
		_, _ = w.Write(response.NewReviewsFailResponse(data))

		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.NewReviewsSuccessResponseDelete())
}
