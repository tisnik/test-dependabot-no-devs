package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/services"
	"github.com/google/uuid"
)

type RecommendationHandler struct {
	service *services.RecommendationService
}

func NewRecommendationHandler(service *services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: service}
}

func (h *RecommendationHandler) RecommendContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/recommendations/content/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "content" {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"path": "Invalid path format",
		})
		return
	}

	contentID := parts[0]
	_, err := uuid.Parse(contentID)
	if err != nil {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"contentId": "ID is not a valid UUID.",
		})
		return
	}

	if !h.service.ReviewService.ContentExists(contentID) {
		RespondFail(w, http.StatusNotFound, map[string]any{
			"contentId": "Content with such ID not found.",
		})
		return
	}

	limit, offset := parsePagination(r)
	recommendations := h.service.RecommendContentToContent(contentID, limit, offset)
	RespondRecommendations(w, recommendations)
}

func (h *RecommendationHandler) RecommendToUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/recommendations/users/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "content" {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"path": "Invalid path format",
		})
		return
	}

	userID := parts[0]
	_, err := uuid.Parse(userID)
	if err != nil {
		RespondFail(w, http.StatusBadRequest, map[string]any{
			"userId": "ID is not a valid UUID.",
		})
		return
	}

	if !h.service.ReviewService.UserExists(userID) {
		RespondFail(w, http.StatusNotFound, map[string]any{
			"userId": "User with such ID not found.",
		})
		return
	}

	limit, offset := parsePagination(r)
	recommendations := h.service.RecommendContentToUser(userID, limit, offset)
	RespondRecommendations(w, recommendations)
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 10
	offset = 0

	l := r.URL.Query().Get("limit")
	if l != "" {
		parsed, err := strconv.Atoi(l)
		if err == nil && parsed >= 0 {
			limit = parsed
		}
	}

	o := r.URL.Query().Get("offset")
	if o != "" {
		parsed, err := strconv.Atoi(o)
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
