package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/service"
)

type ReviewHandler struct {
	logger        *slog.Logger
	reviewService *service.ReviewService
}

func NewReviewHandler(
	logger *slog.Logger,
	reviewService *service.ReviewService,
) (h *ReviewHandler) {
	return &ReviewHandler{
		logger:        logger,
		reviewService: reviewService,
	}
}

func (h *ReviewHandler) IngestReviews(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			h.logger.Error("failed to close request body", "err", err)
		}
	}()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "err", err)

		w.WriteHeader(http.StatusInternalServerError)
		responseBytes := getReviewErrorResponseBytes("Failed to read request body")
		_, _ = w.Write(responseBytes)

		return
	}

	var request model.RawReviewsRequest

	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		h.logger.Error("failed to parse request body", "err", err)

		data := map[string]any{
			"format": err.Error(),
		}

		w.WriteHeader(http.StatusBadRequest)
		responseBytes := getReviewFailResponseBytes(data)
		_, _ = w.Write(responseBytes)

		return
	}

	if request.Data == nil || request.Data.Reviews == nil || len(*request.Data.Reviews) == 0 {
		h.logger.Info("No reviews given")

		reviews := make([]model.Review, 0)

		w.WriteHeader(http.StatusCreated)
		responseBytes := getReviewSuccessResponseBytes(&reviews)
		_, _ = w.Write(responseBytes)

		return
	}

	reviews, err := h.reviewService.IngestReviews(request.Data.Reviews)
	if err != nil {
		h.logger.Error("supplied data format or review values are invalid", "err", err)
		data := validationErrorToMap(err)

		w.WriteHeader(http.StatusBadRequest)
		responseBytes := getReviewFailResponseBytes(data)
		_, _ = w.Write(responseBytes)

		return
	}

	response := model.ReviewsSuccessResponse{
		Status: model.Success,
		Data: model.Reviews{
			Reviews: &reviews,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to construct response", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		responseBytes := getReviewErrorResponseBytes("Failed to construct response")
		_, _ = w.Write(responseBytes)

		return
	}

	h.logger.Info("Successfully created reviews")

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(responseBytes)
}

func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request, reviewId string) {
	err := h.reviewService.DeleteReview(reviewId)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			h.logger.Error("No such review found", "error", err)

			data := map[string]any{
				"reviewId": "Review with such ID not found.",
			}

			w.WriteHeader(http.StatusNotFound)
			responseBytes := getReviewFailResponseBytes(data)
			_, _ = w.Write(responseBytes)

			return
		}

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			h.logger.Error("Supplied data is invalid", "error", err)

			data := map[string]any{
				"reviewId": "ID is not a valid UUID.",
			}

			w.WriteHeader(http.StatusBadRequest)
			responseBytes := getReviewFailResponseBytes(data)
			_, _ = w.Write(responseBytes)
		}

		return
	}

	response := model.ReviewsSuccessResponse{
		Status: model.Success,
		Data:   model.Reviews{},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to construct response", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		responseBytes := getReviewErrorResponseBytes("Failed to construct response")
		_, _ = w.Write(responseBytes)

		return
	}

	h.logger.Info("Successfully deleted review")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
}

func validationErrorToMap(err error) map[string]any {
	data := map[string]any{}
	var fieldsErr validator.ValidationErrors
	if errors.As(err, &fieldsErr) {
		for _, field := range fieldsErr {
			switch field.Tag() {
			case "required":
				data[strings.ToLower(field.Field()[:1])+field.Field()[1:]] = "The field is required."
			case "uuid":
				data[strings.ToLower(field.Field()[:1])+field.Field()[1:]] = "Invalid UUID."
			case "datetime":
				data[strings.ToLower(field.Field()[:1])+field.Field()[1:]] = "Invalid date format."
			default:
				data[strings.ToLower(field.Field()[:1])+field.Field()[1:]] = "Invalid value."
			}
		}
	}
	return data
}

func getReviewErrorResponseBytes(message string) []byte {
	code := http.StatusInternalServerError
	response := model.ReviewsErrorResponse{
		Code:    &code,
		Status:  model.Error,
		Message: message,
		Data:    nil,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return responseBytes
}

func getReviewFailResponseBytes(data map[string]any) []byte {
	response := model.ReviewsFailResponse{
		Status: model.Fail,
		Data:   data,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return responseBytes
}

func getReviewSuccessResponseBytes(reviews *[]model.Review) []byte {
	response := model.ReviewsSuccessResponse{
		Status: model.Success,
		Data: model.Reviews{
			Reviews: reviews,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return responseBytes
}
