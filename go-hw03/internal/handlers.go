package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type App struct {
	Store *Storage
}

func NewApp() *App {
	return &App{
		Store: NewStorage(),
	}
}

func (app *App) Routing(r chi.Router) {
	r.Post("/api/v1/reviews", app.IngestReviews)
	r.Delete("/api/v1/reviews/{reviewId}", app.DeleteReviews)

	r.Get("/api/v1/recommendations/content/{contentId}/content", app.RecommendContendToContent)
	r.Get("/api/v1/recommendations/user/{userId}/content", app.RecommendContendToUser)
}

func (app *App) IngestReviews(w http.ResponseWriter, r *http.Request) {
	var request RawReviewsRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"json": "Invalid json in request"}})
		return
	}

	if len(request.Data.Reviews) == 0 {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"reviews": "No reviews found"}})
		return
	}
	payload := make([]Review, 0, len(request.Data.Reviews))

	// decompose
	for _, review := range request.Data.Reviews {
		// validate ContentIDs
		_, err = uuid.Parse(review.ContentID)
		if err != nil {
			sendJSON(
				w,
				http.StatusBadRequest,
				JSendFail{Status: "fail", Data: map[string]string{"contentID": "ID is not a valid UUID."}})
			return
		}

		// validate UserIDs
		_, err = uuid.Parse(review.UserID)
		if err != nil {
			sendJSON(
				w,
				http.StatusBadRequest,
				JSendFail{Status: "fail", Data: map[string]string{"userID": "ID is not a valid UUID."}})
			return
		}

		// validate Time format
		if review.Released != "" {
			_, err = time.Parse("2006-01-02", review.Released)
			if err != nil {
				sendJSON(
					w,
					http.StatusBadRequest,
					JSendFail{Status: "fail", Data: map[string]string{"released": "Invalid date formats."}})
			}
		}

		// generate ID param to Review
		id := uuid.New().String()
		completedReview := Review{
			ID:        id,
			RawReview: review,
		}

		// Store to DB
		app.Store.AddReview(completedReview)
		payload = append(payload, completedReview)
	}

	sendJSON(
		w,
		http.StatusCreated,
		JSendSuccess{Status: "success", Data: Reviews{Reviews: payload}})
}

func (app *App) DeleteReviews(w http.ResponseWriter, r *http.Request) {
	reviewId := chi.URLParam(r, "reviewId")
	_, err := uuid.Parse(reviewId)
	if err != nil {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"reviewId": "ID is not a valid UUID."}})
		return
	}

	err = app.Store.DeleteReview(reviewId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			sendJSON(
				w,
				http.StatusNotFound,
				JSendFail{Status: "fail", Data: map[string]string{"reviewId": "Review with such ID not found."}})
		}

		sendJSON(
			w,
			http.StatusInternalServerError,
			JSendError{
				Status:  "error",
				Message: "Unable to communicate with database",
				Code:    http.StatusInternalServerError,
			},
		)
		return
	}

	sendJSON(
		w,
		http.StatusOK,
		JSendSuccess{Status: "success"})
}

func (app *App) RecommendContendToContent(w http.ResponseWriter, r *http.Request) {
	contentId := chi.URLParam(r, "contentId")
	_, err := uuid.Parse(contentId)
	if err != nil {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"contentID": "ID is not a valid UUID."}})
		return
	}

	limit, offset, err := limitOffsetParse(r)
	if err != nil {
		sendJSON(w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"limit/offset": "Invalid limit/offset"}})
		return
	}

	if !app.Store.HasContent(contentId) {
		sendJSON(w,
			http.StatusNotFound,
			JSendFail{Status: "fail", Data: map[string]string{"contentId": "Content with such ID not found."}})
		return
	}

	list := RecommendContentToContent(app.Store, contentId, limit, offset)
	sendJSON(
		w,
		http.StatusOK,
		JSendSuccess{Status: "success", Data: Recommendations{Recommendations: list}})
}

func (app *App) RecommendContendToUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	_, err := uuid.Parse(userId)
	if err != nil {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"userId": "ID is not a valid UUID."}})
		return
	}

	limit, offset, err := limitOffsetParse(r)
	if err != nil {
		sendJSON(
			w,
			http.StatusBadRequest,
			JSendFail{Status: "fail", Data: map[string]string{"limit/offset": "Invalid limit/offset"}})
		return
	}

	if !app.Store.HasUser(userId) {
		sendJSON(
			w,
			http.StatusNotFound,
			JSendFail{Status: "fail", Data: map[string]string{"userId": "User with such ID not found."}})
		return
	}

	list := RecommendContentToUser(app.Store, userId, limit, offset)
	sendJSON(
		w,
		http.StatusOK,
		JSendSuccess{Status: "success", Data: Recommendations{Recommendations: list}})
}

func limitOffsetParse(r *http.Request) (limit, offset int, err error) {
	limit, err = converter("limit", r)
	if err != nil {
		return 0, 0, err
	}

	offset, err = converter("offset", r)
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

func converter(term string, r *http.Request) (value int, err error) {
	statement := r.URL.Query().Get(term)
	if statement != "" {
		value, err := strconv.Atoi(statement)
		if err != nil || value < 0 {
			return 0, fmt.Errorf("%s is not a valid value", term)
		} else {
			return value, nil
		}
	}

	// can reach?
	return 0, nil
}

func sendJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}
