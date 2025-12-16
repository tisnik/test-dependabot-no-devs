package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func RecommendContentToContentHandler(s *service.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "contentId")
		_, err := uuid.Parse(id)
		if id == "" || err != nil {
			WriteFailResponse(w, http.StatusBadRequest, map[string]string{"contentId": "ID is not a valid UUID."})
			return
		}
		recommendations, validationErrors, err := getRecommendationsBase(r, s.RecommendContentToContent, id)
		if errors.Is(err, service.ErrInvalidJsonParams) {
			WriteFailResponse(w, http.StatusBadRequest, validationErrors)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			WriteFailResponse(w, http.StatusNotFound, map[string]string{"contentId": "Content with such ID not found."})
			return
		}
		if err != nil {
			WriteErrorResponse(w)
			return
		}
		response := domain.SuccessResponse[domain.Recommendations]{
			Status: domain.StatusSuccess,
			Data: domain.Recommendations{
				Recommendations: recommendations,
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func RecommendContentToUserHandler(s *service.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userId")
		w.Header().Set("Content-Type", "application/json")
		_, err := uuid.Parse(id)
		if id == "" || err != nil {
			WriteFailResponse(w, http.StatusBadRequest, map[string]string{"userId": "ID is not a valid UUID."})
			return
		}
		recommendations, validationErrors, err := getRecommendationsBase(r, s.RecommendContentToUser, id)
		if errors.Is(err, service.ErrInvalidJsonParams) {
			WriteFailResponse(w, http.StatusBadRequest, validationErrors)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			WriteFailResponse(w, http.StatusNotFound, map[string]string{"userId": "User with such ID not found."})
			return
		}
		if err != nil {
			WriteErrorResponse(w)
			return
		}
		response := domain.SuccessResponse[domain.Recommendations]{
			Status: domain.StatusSuccess,
			Data: domain.Recommendations{
				Recommendations: recommendations,
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func getRecommendationsBase(r *http.Request, getter func(string) (
	[]domain.Recommendation, error),
	key string,
) ([]domain.Recommendation, map[string]string, error) {
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	limit := -1
	var offset int
	errors := make(map[string]string, fieldsToValidate)
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 0 {
			errors["limit"] = "Limit must be non-negative integer."
		}
		limit = l
	}
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			errors["offset"] = "Offset must be non-negative integer."
		}
		offset = o
	}
	if len(errors) > 0 {
		return nil, errors, service.ErrInvalidJsonParams
	}
	recommendations, err := getter(key)
	if err != nil {
		return nil, errors, err
	}
	var result []domain.Recommendation
	if offset > len(recommendations) {
		return []domain.Recommendation{}, errors, nil
	}
	if limit != -1 {
		result = recommendations[offset:min(offset+limit, len(recommendations))]
	} else {
		result = recommendations[offset:]
	}
	return result, errors, nil
}
