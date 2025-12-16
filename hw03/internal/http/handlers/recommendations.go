package handlers

import (
	"net/http"

	"github.com/course-go/reelgoofy/internal/recommendations"
	"github.com/go-chi/chi/v5"
)

type RecommendationsHandler struct {
	service *recommendations.RecommendationService
}

func NewRecommendationsHandler(service *recommendations.RecommendationService) *RecommendationsHandler {
	return &RecommendationsHandler{
		service: service,
	}
}

func (h *RecommendationsHandler) RecommendContentToContent(w http.ResponseWriter, r *http.Request) {
	contentID := chi.URLParam(r, "contentId")
	err := validateUUID("contentId", contentID)
	if err != nil {
		respondWithError(w, err)
		return
	}

	limit, offset := getPaginationParams(r)

	result, err := h.service.RecommendContentToContent(contentID, limit, offset)
	if err != nil {
		respondWithError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

func (h *RecommendationsHandler) RecommendContentToUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	err := validateUUID("userId", userID)
	if err != nil {
		respondWithError(w, err)
		return
	}

	limit, offset := getPaginationParams(r)

	result, err := h.service.RecommendContentToUser(userID, limit, offset)
	if err != nil {
		respondWithError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}
