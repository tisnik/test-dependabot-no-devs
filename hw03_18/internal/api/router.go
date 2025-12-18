package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/recommend"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

func NewRouter(repo *repository.InMemoryRepository) http.Handler {
	r := chi.NewRouter()
	engine := recommend.NewEngine(repo)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/reviews", func(w http.ResponseWriter, r *http.Request) { handleIngestReviews(w, r, repo) })
		api.Delete(
			"/reviews/{reviewId}",
			func(w http.ResponseWriter, r *http.Request) { handleDeleteReview(w, r, repo) },
		)
		api.Get(
			"/recommendations/content/{contentId}/content",
			func(w http.ResponseWriter, r *http.Request) { handleRecommendContentToContent(w, r, engine) },
		)
		api.Get(
			"/recommendations/users/{userId}/content",
			func(w http.ResponseWriter, r *http.Request) { handleRecommendContentToUser(w, r, engine) },
		)
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	// Encode first to avoid writing headers when encoding fails
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func handleIngestReviews(w http.ResponseWriter, r *http.Request, repo *repository.InMemoryRepository) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			model.ReviewsErrorResponse{
				Status:  model.StatusError,
				Message: "unable to read body",
				Code:    http.StatusInternalServerError,
			},
		)
		return
	}
	var req model.RawReviewsRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.ReviewsFailResponse{Status: model.StatusFail, Data: map[string]string{"payload": "Invalid JSON."}},
		)
		return
	}
	if len(req.Data.Reviews) == 0 {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.ReviewsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"reviews": "Must contain at least one review."},
			},
		)
		return
	}
	errs := make(map[string]string)
	valid := make([]model.RawReview, 0, len(req.Data.Reviews))
	for i, rr := range req.Data.Reviews {
		_, e := uuid.Parse(rr.ContentID)
		if e != nil {
			errs["contentId"] = "ID is not a valid UUID."
		}
		_, e = uuid.Parse(rr.UserID)
		if e != nil {
			errs["userId"] = "ID is not a valid UUID."
		}
		if rr.Released != "" {
			_, e = time.Parse("2006-01-02", rr.Released)
			if e != nil {
				errs["released"] = "Invalid date formats."
			}
		}
		if rr.Score < 0 || rr.Score > 100 {
			errs["score"] = "Score must be between 0 and 100."
		}
		if rr.Review == "" {
			errs["review"] = "Review text required."
		}
		if len(errs) == 0 {
			valid = append(valid, rr)
		} else if i == 0 {
			break
		}
	}
	if len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, model.ReviewsFailResponse{Status: model.StatusFail, Data: errs})
		return
	}
	added, _ := repo.AddReviews(valid)
	writeJSON(
		w,
		http.StatusCreated,
		model.ReviewsSuccessResponse{Status: model.StatusSuccess, Data: model.ReviewsResponseData{Reviews: added}},
	)
}

func handleDeleteReview(w http.ResponseWriter, r *http.Request, repo *repository.InMemoryRepository) {
	id := chi.URLParam(r, "reviewId")
	_, err := uuid.Parse(id)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.ReviewsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"reviewId": "ID is not a valid UUID."},
			},
		)
		return
	}
	ok, _ := repo.DeleteReview(id)
	if !ok {
		writeJSON(
			w,
			http.StatusNotFound,
			model.ReviewsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"reviewId": "Review with such ID not found."},
			},
		)
		return
	}
	writeJSON(w, http.StatusOK, model.GenericSuccessResponse{Status: model.StatusSuccess, Data: nil})
}

func parseLimitOffset(r *http.Request) (int, int, map[string]string) {
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	limit := defaultLimit
	offset := 0
	errs := make(map[string]string)
	if limitStr != "" {
		parsedLimit, convErr := strconv.Atoi(limitStr)
		if convErr != nil || parsedLimit < 0 {
			errs["limit"] = "limit must be a non-negative integer."
		} else {
			limit = parsedLimit
		}
	}
	if offsetStr != "" {
		parsedOffset, convErr := strconv.Atoi(offsetStr)
		if convErr != nil || parsedOffset < 0 {
			errs["offset"] = "offset must be a non-negative integer."
		} else {
			offset = parsedOffset
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if len(errs) > 0 {
		return 0, 0, errs
	}
	return limit, offset, nil
}

func handleRecommendContentToContent(w http.ResponseWriter, r *http.Request, eng *recommend.Engine) {
	contentID := chi.URLParam(r, "contentId")
	_, err := uuid.Parse(contentID)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.RecommendationsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"contentId": "ID is not a valid UUID."},
			},
		)
		return
	}
	limit, offset, perr := parseLimitOffset(r)
	if perr != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.RecommendationsFailResponse{Status: model.StatusFail, Data: perr},
		)
		return
	}
	recs := eng.RecommendContentToContent(contentID, limit, offset)
	if recs == nil {
		writeJSON(
			w,
			http.StatusNotFound,
			model.RecommendationsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"contentId": "Content with such ID not found."},
			},
		)
		return
	}
	writeJSON(
		w,
		http.StatusOK,
		model.RecommendationsSuccessResponse{
			Status: model.StatusSuccess,
			Data:   model.RecommendationsResponseData{Recommendations: recs},
		},
	)
}

func handleRecommendContentToUser(w http.ResponseWriter, r *http.Request, eng *recommend.Engine) {
	userID := chi.URLParam(r, "userId")
	_, err := uuid.Parse(userID)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.RecommendationsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"userId": "ID is not a valid UUID."},
			},
		)
		return
	}
	limit, offset, perr := parseLimitOffset(r)
	if perr != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			model.RecommendationsFailResponse{Status: model.StatusFail, Data: perr},
		)
		return
	}
	recs := eng.RecommendContentToUser(userID, limit, offset)
	if recs == nil {
		writeJSON(
			w,
			http.StatusNotFound,
			model.RecommendationsFailResponse{
				Status: model.StatusFail,
				Data:   map[string]string{"userId": "User with such ID not found."},
			},
		)
		return
	}
	writeJSON(
		w,
		http.StatusOK,
		model.RecommendationsSuccessResponse{
			Status: model.StatusSuccess,
			Data:   model.RecommendationsResponseData{Recommendations: recs},
		},
	)
}
