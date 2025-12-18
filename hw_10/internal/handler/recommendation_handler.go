package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	apierrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/handler/dto"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RecommendationsHandler struct {
	service *service.RecommendationService
}

func NewRecommendationsHandler(service *service.RecommendationService) *RecommendationsHandler {
	return &RecommendationsHandler{
		service: service,
	}
}

func (h *RecommendationsHandler) GetRecommendationsByUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	inputErrors := make(map[string]string)
	userId, err := uuid.Parse(idStr)
	if err != nil {
		inputErrors["userId"] = "userId must be a valid UUID"
	}

	limit, err := resolveLimit(r)
	if err != nil {
		inputErrors["limit"] = "limit must be a valid integer"
	}

	offset, err := resolveOffset(r)
	if err != nil {
		inputErrors["offset"] = "offset must be a valid integer"
	}

	if len(inputErrors) > 0 {
		response := dto.FailResponse{
			Status: "fail",
			Data:   inputErrors,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	recommendations, err := h.service.GetRecommendationsByUser(userId, limit, offset)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			response := dto.FailResponse{
				Status: "fail",
				Data: map[string]string{
					"userId": "User with such ID not found.",
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

	data := dto.RecommendationData{
		Recommendations: recommendations,
	}

	response := dto.RecommendationResponse{
		Status: "success",
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func (h *RecommendationsHandler) GetRecommendationsByContent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "contentId")
	inputErrors := make(map[string]string)
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		inputErrors["contentId"] = "contentId must be a valid UUID"
	}

	limit, err := resolveLimit(r)
	if err != nil {
		inputErrors["limit"] = "limit must be a valid integer"
	}

	offset, err := resolveOffset(r)
	if err != nil {
		inputErrors["offset"] = "offset must be a valid integer"
	}

	if len(inputErrors) > 0 {
		response := dto.FailResponse{
			Status: "fail",
			Data:   inputErrors,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
		return
	}

	recommendations, err := h.service.GetRecommendationsByContent(contentId, limit, offset)
	if err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			response := dto.FailResponse{
				Status: "fail",
				Data: map[string]string{
					"contentId": "Content with such ID not found.",
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

	data := dto.RecommendationData{
		Recommendations: recommendations,
	}

	response := dto.RecommendationResponse{
		Status: "success",
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func resolveLimit(r *http.Request) (int, error) {
	limit := 100
	limitStr := r.URL.Query().Get("limit")

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return 0, errors.New("invalid limit")
		}
	}

	return limit, nil
}

func resolveOffset(r *http.Request) (int, error) {
	var offset int
	offsetStr := r.URL.Query().Get("offset")

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, errors.New("invalid offset")
		}
	}

	return offset, nil
}
