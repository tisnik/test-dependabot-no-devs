package http

import (
	"net/http"

	"github.com/course-go/reelgoofy/internal/containers/recommendations/services"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/core/request"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/go-chi/chi/v5"
)

// RecommendationRequestHandler represents a controller.
type RecommendationRequestHandler struct {
	repo repository.ReviewRepository
}

// NewRecommendationRequestHandler is a constructor of RecommendationRequestHandler.
func NewRecommendationRequestHandler(repo repository.ReviewRepository) *RecommendationRequestHandler {
	return &RecommendationRequestHandler{
		repo: repo,
	}
}

// GetContentRecommendations handles HTTP requests for content-based recommendations.
// It takes a content ID from URL parameters, validates it, and returns similar content.
func (r *RecommendationRequestHandler) GetContentRecommendations(write http.ResponseWriter, req *http.Request) {
	idParam := "contentId"

	contentId := chi.URLParam(req, idParam)
	id, errMap := request.ValidateUuid(contentId, idParam)
	if errMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, errMap, "")
		return
	}

	attributes, validationErrMap := GetValidatedAttributes(req)
	if validationErrMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, validationErrMap, "")
		return
	}

	rec, serviceErrMap := services.RecommendByContentId(id, r.repo, attributes.Limit, attributes.Offset)
	if serviceErrMap != nil {
		response.JSONResponse(write, http.StatusNotFound, serviceErrMap, "")
		return
	}

	resp := Format(rec)
	response.JSONResponse(write, http.StatusOK, resp, "")
}

// GetUserRecommendations handles HTTP requests for user-based recommendations.
// It takes a user ID from URL parameters, validates it, and returns similar content.
func (r *RecommendationRequestHandler) GetUserRecommendations(write http.ResponseWriter, req *http.Request) {
	idParam := "userId"

	userId := chi.URLParam(req, idParam)
	id, errMap := request.ValidateUuid(userId, idParam)
	if errMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, errMap, string(response.InvalidDataMessage))
		return
	}

	attributes, validationErrMap := GetValidatedAttributes(req)
	if validationErrMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, validationErrMap, "")
		return
	}

	rec, serviceErrMap := services.RecommendByUserId(id, r.repo, attributes.Limit, attributes.Offset)
	if serviceErrMap != nil {
		response.JSONResponse(write, http.StatusBadRequest, serviceErrMap, "")
		return
	}

	resp := Format(rec)
	response.JSONResponse(write, http.StatusOK, resp, "")
}
