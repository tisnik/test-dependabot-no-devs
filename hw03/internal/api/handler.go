package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Markek1/reelgoofy/internal/domain"
	"github.com/Markek1/reelgoofy/internal/repository"
	"github.com/Markek1/reelgoofy/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
	logger  *slog.Logger
}

func NewHandler(service *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) IngestReviews(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data struct {
			Reviews []domain.RawReview `json:"reviews"`
		} `json:"data"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Invalid JSON format", nil)
		return
	}

	reviews, err := h.service.IngestReviews(req.Data.Reviews)
	if err != nil {
		h.logger.Error("IngestReviews service error", "error", err)
		h.writeError(w, http.StatusInternalServerError, domain.StatusError, "Internal server error", nil)
		return
	}

	resp := domain.ReviewsSuccessResponse{
		Status: domain.StatusSuccess,
		Data: &struct {
			Reviews []domain.Review `json:"reviews,omitempty"`
		}{
			Reviews: reviews,
		},
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID := r.PathValue("reviewId")
	if reviewID == "" {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Missing review ID", nil)
		return
	}

	err := h.service.DeleteReview(reviewID)
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			h.writeError(
				w,
				http.StatusNotFound,
				domain.StatusFail,
				"Review not found",
				map[string]string{"reviewId": "Review with such ID not found"},
			)
			return
		}
		if uuid.Validate(reviewID) != nil {
			h.writeError(
				w,
				http.StatusBadRequest,
				domain.StatusFail,
				"Invalid review ID",
				map[string]string{"reviewId": "ID is not a valid UUID."},
			)
			return
		}
		h.logger.Error("DeleteReview service error", "error", err)
		h.writeError(w, http.StatusInternalServerError, domain.StatusError, "Internal server error", nil)
		return
	}

	resp := domain.ReviewsSuccessResponse{
		Status: domain.StatusSuccess,
		Data:   nil,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) RecommendContent(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("contentId")
	if contentID == "" {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Missing content ID", nil)
		return
	}

	err := uuid.Validate(contentID)
	if err != nil {
		h.writeError(
			w,
			http.StatusBadRequest,
			domain.StatusFail,
			"Invalid content ID",
			map[string]string{"contentId": "ID is not a valid UUID."},
		)
		return
	}

	limit, offset, errMap := h.parseLimitOffset(r)
	if errMap != nil {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Invalid pagination parameters", errMap)
		return
	}

	recommendations, err := h.service.RecommendContent(domain.ContentID(contentID), limit, offset)
	if err != nil {
		h.logger.Error("RecommendContent service error", "error", err)
		h.writeError(w, http.StatusInternalServerError, domain.StatusError, "Internal server error", nil)
		return
	}

	resp := domain.RecommendationsSuccessResponse{
		Status: domain.StatusSuccess,
	}
	resp.Data.Recommendations = recommendations

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) RecommendUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Missing user ID", nil)
		return
	}

	err := uuid.Validate(userID)
	if err != nil {
		h.writeError(
			w,
			http.StatusBadRequest,
			domain.StatusFail,
			"Invalid user ID",
			map[string]string{"userId": "ID is not a valid UUID."},
		)
		return
	}

	limit, offset, errMap := h.parseLimitOffset(r)
	if errMap != nil {
		h.writeError(w, http.StatusBadRequest, domain.StatusFail, "Invalid pagination parameters", errMap)
		return
	}

	recommendations, err := h.service.RecommendUser(domain.UserID(userID), limit, offset)
	if err != nil {
		h.logger.Error("RecommendUser service error", "error", err)
		h.writeError(w, http.StatusInternalServerError, domain.StatusError, "Internal server error", nil)
		return
	}

	resp := domain.RecommendationsSuccessResponse{
		Status: domain.StatusSuccess,
	}
	resp.Data.Recommendations = recommendations

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) parseLimitOffset(r *http.Request) (int, int, map[string]string) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	errors := make(map[string]string)

	limit := 0
	if limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val < 0 {
			errors["limit"] = "Limit must be non-negative integer."
		} else {
			limit = val
		}
	}

	offset := 0
	if offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			errors["offset"] = "Offset must be non-negative integer."
		} else {
			offset = val
		}
	}

	if len(errors) > 0 {
		return 0, 0, errors
	}
	return limit, offset, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *Handler) writeError(
	w http.ResponseWriter,
	status int,
	statusStr domain.StatusEnum,
	message string,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if statusStr == domain.StatusError {
		resp := domain.ErrorResponse{
			Status:  statusStr,
			Message: message,
			Code:    status,
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			h.logger.Error("Failed to encode JSON error response", "error", err)
		}
	} else {
		resp := domain.FailResponse{
			Status: statusStr,
			Data:   nil,
		}
		if dataMap, ok := data.(map[string]string); ok {
			resp.Data = dataMap
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			h.logger.Error("Failed to encode JSON fail response", "error", err)
		}
	}
}
