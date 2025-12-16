package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HandlerImpl struct {
	Storage *Storage
}

func NewHandler() *HandlerImpl {
	return &HandlerImpl{
		Storage: NewStorage(),
	}
}

/* ---- Handlers ---- */

func (h *HandlerImpl) RecommendContentToContent(w http.ResponseWriter,
	r *http.Request,
	contentId string,
	params RecommendContentToContentParams,
) {
	defer writeErrorRecommendations(w, "Unable to communicate with database")

	err := validateUUID(contentId)
	if err != nil {
		writeFailRecommendations(w, http.StatusBadRequest, "contentId", "ID is not a valid UUID")
		return
	}

	errors := validatePagination(params.Offset, params.Limit)
	if len(errors) > 0 {
		writeFailMapRecommendations(w, http.StatusBadRequest, errors)
		return
	}

	reviews := h.Storage.GetReviewsByContentId(contentId)
	if len(reviews) == 0 {
		writeFailRecommendations(w, http.StatusNotFound, "contentId", "Content with such ID not found.")
		return
	}

	recsReviews := h.Storage.RecommendContentToContent(contentId, params.Limit, params.Offset)
	recs := make([]Recommendation, 0, len(recsReviews))
	for _, rev := range recsReviews {
		recs = append(recs, Recommendation{
			Id:    rev.ContentId,
			Title: rev.Title,
		})
	}

	writeSuccessRecommendations(w, http.StatusOK, recs)
}

func (h *HandlerImpl) RecommendContentToUser(
	w http.ResponseWriter,
	r *http.Request,
	userId string,
	params RecommendContentToUserParams,
) {
	defer writeErrorRecommendations(w, "Unable to communicate with database")

	err := validateUUID(userId)
	if err != nil {
		writeFailRecommendations(w, http.StatusBadRequest, "userId", "ID is not a valid UUID")
		return
	}

	errors := validatePagination(params.Offset, params.Limit)
	if len(errors) > 0 {
		writeFailMapRecommendations(w, http.StatusBadRequest, errors)
		return
	}

	reviews := h.Storage.GetReviewsByUserId(userId)
	if len(reviews) == 0 {
		writeFailRecommendations(w, http.StatusNotFound, "userId", "User with such ID not found.")
		return
	}

	recsReviews := h.Storage.RecommendContentToUser(userId, params.Limit, params.Offset)
	recs := make([]Recommendation, 0, len(recsReviews))
	for _, rev := range recsReviews {
		recs = append(recs, Recommendation{
			Id:    rev.ContentId,
			Title: rev.Title,
		})
	}
	writeSuccessRecommendations(w, http.StatusOK, recs)
}

func (h *HandlerImpl) IngestReviews(w http.ResponseWriter, r *http.Request) {
	defer writeErrorReviews(w, "Unable to communicate with database")

	var req IngestReviewsJSONRequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeFailReviews(w, http.StatusBadRequest, "error", "Supplied data format is invalid")
		return
	}

	createdReviews := make([]Review, 0, len(*req.Data.Reviews))
	for _, raw := range *req.Data.Reviews {
		if checks := checkIngestReviewReq(raw); len(checks) > 0 {
			writeFailMapReviews(w, http.StatusBadRequest, checks)
			return
		}

		review := Review{
			Id:          uuid.NewString(),
			UserId:      raw.UserId,
			ContentId:   raw.ContentId,
			Review:      raw.Review,
			Title:       raw.Title,
			Actors:      raw.Actors,
			Director:    raw.Director,
			Description: raw.Description,
			Duration:    raw.Duration,
			Genres:      raw.Genres,
			Origins:     raw.Origins,
			Released:    raw.Released,
			Score:       raw.Score,
			Tags:        raw.Tags,
		}
		h.Storage.AddReview(review)
		createdReviews = append(createdReviews, review)
	}
	writeSuccessReviews(w, http.StatusCreated, createdReviews)
}

func (h *HandlerImpl) DeleteReview(w http.ResponseWriter, r *http.Request, reviewId string) {
	defer writeErrorReviews(w, "Unable to communicate with database")

	err := validateUUID(reviewId)
	if err != nil {
		writeFailReviews(w, http.StatusBadRequest, "reviewId", "ID is not a valid UUID.")
		return
	}

	review := h.Storage.GetReviewById(reviewId)
	if review == nil {
		writeFailReviews(w, http.StatusNotFound, "reviewId", "Review with such ID not found.")
		return
	}

	h.Storage.DeleteReview(reviewId)

	writeSuccessReviews(w, http.StatusOK, []Review{*review})
}

/* ---- Helpers ---- */

// Request validation helpers

func validateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("uuid parse error: %w", err)
	}
	return nil
}

func validatePagination(offset, limit *int) map[string]string {
	errors := make(map[string]string)
	if offset != nil && *offset < 0 {
		errors["offset"] = "Offset must be non-negative integer."
	}
	if limit != nil && *limit < 0 {
		errors["limit"] = "Limit must be non-negative integer."
	}
	return errors
}

func validateParams(name string, params *[]string, errors map[string]string) {
	if params != nil {
		for _, s := range *params {
			if strings.TrimSpace(s) == "" {
				errors[name] = "Cannot contain empty strings."
				break
			}
		}
	}
}

func checkIngestReviewReq(raw RawReview) map[string]string {
	errors := make(map[string]string)

	_, err := uuid.Parse(raw.ContentId)
	if err != nil {
		errors["contentId"] = "ID is not a valid UUID."
	}

	_, err = uuid.Parse(raw.UserId)
	if err != nil {
		errors["userId"] = "ID is not a valid UUID."
	}
	if strings.TrimSpace(raw.Review) == "" {
		errors["review"] = "Review is required."
	}
	// My choice of score scale 0-100
	if raw.Score == nil {
		errors["score"] = "Score is required."
	}
	if raw.Score != nil && (*raw.Score < 0 || *raw.Score > 100) {
		errors["score"] = "Score must be between 0 and 100."
	}

	if raw.Released != nil && *raw.Released != "" {
		_, err := time.Parse("2006-01-02", *raw.Released)
		if err != nil {
			errors["released"] = "Invalid date formats."
		}
	}

	validateParams("actors", raw.Actors, errors)
	validateParams("genres", raw.Genres, errors)
	validateParams("tags", raw.Tags, errors)
	validateParams("origins", raw.Origins, errors)

	return errors
}

func writeJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// Review response helpers

func writeFailReviews(w http.ResponseWriter, status int, field string, message string) {
	writeJSON(w, status, ReviewsFailResponse{
		Data:   map[string]any{field: message},
		Status: Fail,
	})
}

func writeFailMapReviews(w http.ResponseWriter, status int, errors map[string]string) {
	data := make(map[string]any, len(errors))
	for k, v := range errors {
		data[k] = v
	}
	writeJSON(w, status, ReviewsFailResponse{
		Data:   data,
		Status: Fail,
	})
}

func writeSuccessReviews(w http.ResponseWriter, status int, reviews []Review) {
	writeJSON(w, status, ReviewsSuccessResponse{
		Data:   Reviews{Reviews: &reviews},
		Status: Success,
	})
}

func writeErrorReviews(w http.ResponseWriter, message string) {
	if rec := recover(); rec != nil {
		code := http.StatusInternalServerError
		emptyData := map[string]any{}
		writeJSON(w, code, ReviewsErrorResponse{
			Code:    &code,
			Data:    &emptyData,
			Message: message,
			Status:  Error,
		})
	}
}

// Recomendations response helpers

func writeSuccessRecommendations(w http.ResponseWriter, status int, recs []Recommendation) {
	writeJSON(w, status, RecommendationsSuccessResponse{
		Data:   Recommendations{Recommendations: &recs},
		Status: Success,
	})
}

func writeFailRecommendations(w http.ResponseWriter, status int, field string, message string) {
	writeJSON(w, status, RecommendationsFailResponse{
		Data:   map[string]any{field: message},
		Status: Fail,
	})
}

func writeFailMapRecommendations(w http.ResponseWriter, status int, errors map[string]string) {
	data := make(map[string]any, len(errors))
	for k, v := range errors {
		data[k] = v
	}
	writeJSON(w, status, RecommendationsFailResponse{
		Data:   data,
		Status: Fail,
	})
}

func writeErrorRecommendations(w http.ResponseWriter, message string) {
	rec := recover()
	if rec != nil {
		code := http.StatusInternalServerError
		emptyData := map[string]any{}
		writeJSON(w, code, RecommendationsErrorResponse{
			Code:    &code,
			Data:    &emptyData,
			Message: message,
			Status:  Error,
		})
	}
}
