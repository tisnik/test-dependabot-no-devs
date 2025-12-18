package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/service"
)

type RecommendationHandler struct {
	logger                *slog.Logger
	recommendationService *service.RecommendationService
}

func NewRecommendationHandler(
	logger *slog.Logger,
	recommendationService *service.RecommendationService,
) (h *RecommendationHandler) {
	return &RecommendationHandler{
		logger:                logger,
		recommendationService: recommendationService,
	}
}

func (h *RecommendationHandler) RecommendContentToContent(
	w http.ResponseWriter,
	r *http.Request,
	contentId string,
	params model.RecommendContentToContentParams,
) {
	data := h.checkParams(params.Limit, params.Offset)
	if len(*data) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		responseBytes := getRecommendFailResponseBytes(*data)
		_, _ = w.Write(responseBytes)

		return
	}

	recommendations, err := h.recommendationService.RecommendContentToContent(contentId, params)
	if err != nil {
		if errors.Is(err, service.ErrContentNotFound) {
			h.logger.Error("No such content found", "error", err)

			data := map[string]any{
				"contentId": "Content with such ID not found.",
			}

			w.WriteHeader(http.StatusNotFound)
			responseBytes := getRecommendFailResponseBytes(data)
			_, _ = w.Write(responseBytes)

			return
		}

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			h.logger.Error("Supplied data is invalid", "error", err)

			data := map[string]any{
				"contentId": "ID is not a valid UUID.",
			}

			w.WriteHeader(http.StatusBadRequest)
			responseBytes := getRecommendFailResponseBytes(data)
			_, _ = w.Write(responseBytes)
		}

		return
	}

	response := model.RecommendationsSuccessResponse{
		Status: model.Success,
		Data: model.Recommendations{
			Recommendations: &recommendations,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to construct response", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		responseBytes := getRecommendErrorResponseBytes("Failed to construct response")
		_, _ = w.Write(responseBytes)

		return
	}

	h.logger.Info("Successfully recommend content.")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
}

func (h *RecommendationHandler) RecommendContentToUser(
	w http.ResponseWriter,
	r *http.Request,
	userId string,
	params model.RecommendContentToUserParams,
) {
	data := h.checkParams(params.Limit, params.Offset)
	if len(*data) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		responseBytes := getRecommendFailResponseBytes(*data)
		_, _ = w.Write(responseBytes)

		return
	}

	recommendations, err := h.recommendationService.RecommendContentToUser(userId, params)
	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			h.logger.Error("Supplied data is invalid", "error", err)

			data := map[string]any{
				"userId": "ID is not a valid UUID.",
			}

			w.WriteHeader(http.StatusBadRequest)
			responseBytes := getRecommendFailResponseBytes(data)
			_, _ = w.Write(responseBytes)
		}

		return
	}

	response := model.RecommendationsSuccessResponse{
		Status: model.Success,
		Data: model.Recommendations{
			Recommendations: &recommendations,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to construct response", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		responseBytes := getRecommendErrorResponseBytes("Failed to construct response")
		_, _ = w.Write(responseBytes)

		return
	}

	h.logger.Info("Successfully recommend content.")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
}

func (h *RecommendationHandler) checkParams(limit *int, offset *int) *map[string]any {
	const numberOfParams = 2
	data := make(map[string]any, numberOfParams)
	if offset != nil && *offset < 0 {
		h.logger.Error("Offset is negative")
		data["offset"] = "Offset must be non-negative integer."
	}

	if limit != nil && *limit < 0 {
		h.logger.Error("Limit is negative")
		data["limit"] = "Limit must be non-negative integer."
	}

	return &data
}

func getRecommendErrorResponseBytes(message string) []byte {
	code := http.StatusInternalServerError
	response := model.RecommendationsErrorResponse{
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

func getRecommendFailResponseBytes(data map[string]any) []byte {
	response := model.RecommendationsFailResponse{
		Status: model.Fail,
		Data:   data,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return responseBytes
}
