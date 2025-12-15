package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

type Database struct {
	Reviews map[string]Review
	mu      sync.Mutex
}
type Review struct {
	ID          string   `json:"id"`
	ContentID   string   `json:"contentId"`
	UserID      string   `json:"userId"`
	Title       string   `json:"title"`
	Genres      []string `json:"genres"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Director    string   `json:"director"`
	Actors      []string `json:"actors"`
	Origins     []string `json:"origins"`
	Duration    int      `json:"duration"`
	Released    string   `json:"released"`
	ReviewText  string   `json:"review"`
	Score       int      `json:"score"`
}
type ReviewsRequest struct {
	Data ReviewsData `json:"data"`
}
type ReviewsData struct {
	Reviews []Review `json:"reviews"`
}
type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type RecommendationsData struct {
	Recommendations []Recommendation `json:"recommendations"`
}
type Response struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    any    `json:"data"`
}

type Server struct {
	db *Database
}

func main() {
	s := &Server{
		db: &Database{
			Reviews: make(map[string]Review),
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/reviews", s.ingestReviews)
	mux.HandleFunc("DELETE /api/v1/reviews/{id}", s.deleteReview)
	mux.HandleFunc("GET /api/v1/recommendations/content/{contentId}/content", s.recommendContentToContent)
	mux.HandleFunc("GET /api/v1/recommendations/users/{userId}/content", s.recommendContentToUser)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) ingestReviews(w http.ResponseWriter, r *http.Request) {
	request := ReviewsRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeInternalError(w, err)
		return
	}
	reqErrors := validateRequest(request)
	if len(reqErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status: "fail",
			Data:   reqErrors,
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()

	savedReviews := make([]Review, 0, len(request.Data.Reviews))
	for _, review := range request.Data.Reviews {
		id := uuid.NewString()
		review.ID = id
		s.db.Reviews[id] = review
		savedReviews = append(savedReviews, review)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := Response{
		Status: "success",
		Data: ReviewsData{
			Reviews: savedReviews,
		},
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeInternalError(w, err)
	}
}

func validateRequest(request ReviewsRequest) map[string]string {
	validationErrors := make(map[string]string)
	for _, review := range request.Data.Reviews {
		_, err := uuid.Parse(review.ContentID)
		if err != nil {
			validationErrors["contentId"] = "ID is not a valid UUID."
		}
		_, err = uuid.Parse(review.UserID)
		if err != nil {
			validationErrors["userId"] = "ID is not a valid UUID."
		}
		if review.Released != "" {
			_, err = time.Parse("2006-01-02", review.Released)
			if err != nil {
				validationErrors["released"] = "Invalid date formats."
			}
		}
		if review.ReviewText == "" {
			validationErrors["review"] = "Review content is required."
		}
		if review.Score == 0 {
			validationErrors["score"] = "Review score is required."
		}
	}
	return validationErrors
}

func (s *Server) deleteReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status: "fail",
			Data: struct {
				ReviewId string `json:"reviewId"`
			}{ReviewId: "ID is not a valid UUID."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	_, ok := s.db.Reviews[id]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := Response{
			Status: "fail",
			Data: struct {
				ReviewId string `json:"reviewId"`
			}{ReviewId: "Review with such ID not found."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	delete(s.db.Reviews, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := Response{
		Status: "success",
		Data:   nil,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeInternalError(w, err)
	}
}

func (s *Server) recommendContentToContent(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("contentId")
	_, err := uuid.Parse(contentID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status: "fail",
			Data: struct {
				ContentId string `json:"contentId"`
			}{ContentId: "ID is not a valid UUID."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	limit, offset, ok := getParams(w, r)
	if !ok {
		return
	}

	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	var targetGenres []string
	foundTarget := false

	for _, review := range s.db.Reviews {
		if review.ContentID == contentID {
			targetGenres = review.Genres
			foundTarget = true
			break
		}
	}
	if !foundTarget {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := Response{
			Status: "fail",
			Data: struct {
				ContentId string `json:"contentId"`
			}{ContentId: "Content with such ID not found."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	seenMovies := map[string]bool{}
	recommendations := s.findRecommendations(contentID, limit, offset, targetGenres, seenMovies)
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Status: "success",
		Data: RecommendationsData{
			Recommendations: recommendations,
		},
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeInternalError(w, err)
	}
}

func (s *Server) recommendContentToUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	_, err := uuid.Parse(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status: "fail",
			Data: struct {
				UserId string `json:"userId"`
			}{UserId: "ID is not a valid UUID."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	limit, offset, ok := getParams(w, r)
	if !ok {
		return
	}

	s.db.mu.Lock()
	defer s.db.mu.Unlock()

	seenMovies := map[string]bool{}
	var userGenres []string
	userExists := false
	for _, review := range s.db.Reviews {
		if review.UserID == userID {
			userExists = true
			seenMovies[review.ContentID] = true
			userGenres = append(userGenres, review.Genres...)
		}
	}
	if !userExists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := Response{
			Status: "fail",
			Data: struct {
				UserId string `json:"userId"`
			}{UserId: "User with such ID not found."},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeInternalError(w, err)
		}
		return
	}
	recommendations := s.findRecommendations(userID, limit, offset, userGenres, seenMovies)
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Status: "success",
		Data: RecommendationsData{
			Recommendations: recommendations,
		},
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeInternalError(w, err)
	}
}

func getParams(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit := 20
	offset := 0

	if val := r.URL.Query().Get("limit"); val != "" {
		l, err := strconv.Atoi(val)
		if err == nil && l >= 0 {
			limit = l
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := Response{
				Status: "fail",
				Data: struct {
					Limit string `json:"limit"`
				}{Limit: "Limit must be non-negative integer."},
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				writeInternalError(w, err)
			}
			return 0, 0, false
		}
	}
	if val := r.URL.Query().Get("offset"); val != "" {
		o, err := strconv.Atoi(val)
		if err == nil && o >= 0 {
			offset = o
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := Response{
				Status: "fail",
				Data: struct {
					Offset string `json:"offset"`
				}{Offset: "Offset must be non-negative integer."},
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				writeInternalError(w, err)
			}
			return 0, 0, false
		}
	}
	return limit, offset, true
}

func (s *Server) findRecommendations(
	id string,
	limit int,
	offset int,
	targetGenres []string,
	seenMovies map[string]bool,
) []Recommendation {
	var rec []Recommendation
	addedToRec := make(map[string]bool)
	for _, review := range s.db.Reviews {
		if seenMovies[review.ContentID] {
			continue
		}
		if review.ContentID == id {
			continue
		}
		if addedToRec[review.ContentID] {
			continue
		}
		hasCommonGenre := false
		for _, genre := range review.Genres {
			if slices.Contains(targetGenres, genre) {
				hasCommonGenre = true
				break
			}
		}
		if hasCommonGenre {
			rec = append(rec, Recommendation{
				ID:    review.ContentID,
				Title: review.Title,
			})
			addedToRec[review.ContentID] = true
		}
	}
	total := len(rec)
	start := min(offset, total)
	end := min(offset+limit, total)
	return rec[start:end]
}

func writeInternalError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := ErrorResponse{
		Status:  "error",
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
		Data:    map[string]string{},
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}
