package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/course-go/reelgoofy/internal/dto/request"
	"github.com/course-go/reelgoofy/internal/dto/response"
	"github.com/course-go/reelgoofy/internal/entity"
	reviews "github.com/course-go/reelgoofy/internal/repository"
	"github.com/course-go/reelgoofy/internal/service/recommendation"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	defaultLimit  = 20
	defaultOffset = 0
)

type ctxKey string

const (
	ctxLimit  ctxKey = "limit"
	ctxOffset ctxKey = "offset"
)

type API struct {
	repository *reviews.Repository
}

func NewRouter(repository *reviews.Repository) *chi.Mux {
	api := &API{repository: repository}

	mux := chi.NewRouter()
	mux.Route("/api/v1", func(r chi.Router) {
		r.Route("/reviews", func(r chi.Router) {
			r.Get("/", api.getReviews)
			r.Post("/", api.createReviews)
			r.Delete("/{id}", api.deleteReview)
		})
		r.Route("/recommendations", func(r chi.Router) {
			r.With(paginate).Get("/content/{contentId}/content", api.recommendContentToContent)
			r.With(paginate).Get("/users/{userId}/content", api.recommendContentToUser)
		})
	})
	return mux
}

func (api *API) getReviews(w http.ResponseWriter, r *http.Request) {
	revs := api.repository.GetReviews()

	reviewsJSON, err := json.Marshal(revs)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(reviewsJSON)
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (api *API) createReviews(w http.ResponseWriter, r *http.Request) {
	var req request.RawReviewsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		failResponse := response.ReviewsFailResponse{
			Status: response.Fail,
			Data:   map[string]any{"body": "Invalid request body format"},
		}
		marshalAndWrite(w, http.StatusBadRequest, failResponse)
		return
	}

	rawReviews := req.Data.Reviews
	for _, review := range rawReviews {
		err := validator.New().Struct(review)
		if err != nil {
			failResponse := response.ReviewsFailResponse{
				Status: response.Fail,
				Data:   map[string]any{"validation failed": err.Error()},
			}
			marshalAndWrite(w, http.StatusBadRequest, failResponse)
			return
		}
	}
	createdReviews := make([]entity.Review, 0, len(rawReviews))
	for _, rawReview := range rawReviews {
		review := entity.Review{
			ContentID:   rawReview.ContentID,
			UserID:      rawReview.UserID,
			Title:       rawReview.Title,
			Genres:      rawReview.Genres,
			Tags:        rawReview.Tags,
			Description: rawReview.Description,
			Director:    rawReview.Director,
			Actors:      rawReview.Actors,
			Origins:     rawReview.Origins,
			Duration:    rawReview.Duration,
			Released:    rawReview.Released,
			Review:      rawReview.Review,
			Score:       rawReview.Score,
		}
		createdReview := api.repository.InsertReview(review)
		createdReviews = append(createdReviews, createdReview)
	}

	successResponse := response.ReviewsSuccessResponse{
		Status: response.Success,
		Data:   &response.ReviewData{Reviews: createdReviews},
	}
	marshalAndWrite(w, http.StatusCreated, successResponse)
}

func (api *API) deleteReview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		failResponse := response.ReviewsFailResponse{
			Status: response.Fail,
			Data:   map[string]any{"reviewId": "invalid UUID format"},
		}
		marshalAndWrite(w, http.StatusBadRequest, failResponse)
		return
	}

	err = api.repository.DeleteReview(id)
	if err != nil {
		if errors.Is(err, reviews.ErrReviewNotFound) {
			failResponse := response.ReviewsFailResponse{
				Status: response.Fail,
				Data:   map[string]any{"reviewId": "review not found"},
			}
			marshalAndWrite(w, http.StatusNotFound, failResponse)
			return
		}
		errorResponse := response.ReviewsErrorResponse{
			Status:  response.Error,
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
			Data:    nil,
		}
		marshalAndWrite(w, http.StatusInternalServerError, errorResponse)
		return
	}

	successResponse := response.ReviewsSuccessResponse{
		Status: response.Success,
		Data:   nil,
	}
	marshalAndWrite(w, http.StatusOK, successResponse)
}

func (api *API) recommendContentToContent(w http.ResponseWriter, r *http.Request) {
	contentIdStr := chi.URLParam(r, "contentId")
	contentId, err := uuid.Parse(contentIdStr)
	if err != nil {
		failResponse := response.RecommendationsFailResponse{
			Status: response.Fail,
			Data:   map[string]any{"contentId": "invalid UUID format"},
		}
		marshalAndWrite(w, http.StatusBadRequest, failResponse)
		return
	}

	service := recommendation.NewRecommendationService(api.repository)
	recommendations, err := service.RecommendByContentID(contentId)
	if err != nil {
		if errors.Is(err, recommendation.ErrContentNotFound) {
			failResponse := response.RecommendationsFailResponse{
				Status: response.Fail,
				Data:   map[string]any{"contentId": "content with given ID not found"},
			}
			marshalAndWrite(w, http.StatusNotFound, failResponse)
			return
		}
		errorResponse := response.RecommendationsErrorResponse{
			Status:  response.Error,
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
			Data:    nil,
		}
		marshalAndWrite(w, http.StatusInternalServerError, errorResponse)
		return
	}

	limit, ok := r.Context().Value(ctxLimit).(int)
	if !ok {
		limit = defaultLimit
	}
	offset, ok := r.Context().Value(ctxOffset).(int)
	if !ok {
		offset = defaultOffset
	}
	pagedRecommendations := recommendations[min(offset, len(recommendations)):min(offset+limit, len(recommendations))]

	recommendationResponse := make([]response.Recommendation, 0, len(pagedRecommendations))
	for _, r := range pagedRecommendations {
		recommendationResponse = append(recommendationResponse, response.Recommendation{
			ID:    r.ContentID,
			Title: r.Title,
		})
	}
	successResponse := response.RecommendationsSuccessResponse{
		Status: response.Success,
		Data:   response.RecommendationData{Recommendations: recommendationResponse},
	}
	marshalAndWrite(w, http.StatusOK, successResponse)
}

func (api *API) recommendContentToUser(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		failResponse := response.RecommendationsFailResponse{
			Status: response.Fail,
			Data:   map[string]any{"userId": "invalid UUID format"},
		}
		marshalAndWrite(w, http.StatusBadRequest, failResponse)
		return
	}

	service := recommendation.NewRecommendationService(api.repository)
	recommendations, err := service.RecommendByUserID(userId)
	if err != nil {
		if errors.Is(err, recommendation.ErrUserHasNoReviews) {
			failResponse := response.RecommendationsFailResponse{
				Status: response.Fail,
				Data:   map[string]any{"userId": "user has no reviews"},
			}
			marshalAndWrite(w, http.StatusNotFound, failResponse)
			return
		}
		errorResponse := response.RecommendationsErrorResponse{
			Status:  response.Error,
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
			Data:    nil,
		}
		marshalAndWrite(w, http.StatusInternalServerError, errorResponse)
		return
	}

	limit, ok := r.Context().Value(ctxLimit).(int)
	if !ok {
		limit = defaultLimit
	}
	offset, ok := r.Context().Value(ctxOffset).(int)
	if !ok {
		offset = defaultOffset
	}
	pagedRecommendations := recommendations[min(offset, len(recommendations)):min(offset+limit, len(recommendations))]

	recommendationResponse := make([]response.Recommendation, 0, len(pagedRecommendations))
	for _, r := range pagedRecommendations {
		recommendationResponse = append(recommendationResponse, response.Recommendation{
			ID:    r.ContentID,
			Title: r.Title,
		})
	}
	successResponse := response.RecommendationsSuccessResponse{
		Status: response.Success,
		Data:   response.RecommendationData{Recommendations: recommendationResponse},
	}
	marshalAndWrite(w, http.StatusOK, successResponse)
}

func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := defaultLimit
		offset := defaultOffset

		lim := r.URL.Query().Get(string(ctxLimit))
		if lim != "" {
			parsedLimit, err := strconv.Atoi(lim)
			if err == nil && parsedLimit >= 0 {
				limit = parsedLimit
			}
		}
		off := r.URL.Query().Get(string(ctxOffset))
		if off != "" {
			parsedOffset, err := strconv.Atoi(off)
			if err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		ctx := context.WithValue(r.Context(), ctxLimit, limit)
		ctx = context.WithValue(ctx, ctxOffset, offset)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func marshalAndWrite(w http.ResponseWriter, status int, payload any) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
