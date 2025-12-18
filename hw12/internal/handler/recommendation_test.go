package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal/handler"
	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestGetRecommendations tests the recommendation endpoint.
func TestGetRecommendations(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemoryStorage()
	reviewService := service.NewReviewService(memStorage)
	recommendationService := service.NewRecommendationService(memStorage)
	apiHandler := handler.NewHandler(reviewService, recommendationService)

	testUserID := uuid.New()
	movie1 := uuid.New()
	movie2 := uuid.New()

	reviews := []model.Review{
		{UserID: testUserID, MovieID: movie1, Rating: 9},
		{UserID: testUserID, MovieID: movie2, Rating: 8},
		{UserID: uuid.New(), MovieID: movie1, Rating: 5},
	}

	for _, review := range reviews {
		body, err := json.Marshal(review)
		if err != nil {
			t.Fatalf("failed to marshal review: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		apiHandler.CreateReview(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("setting up test data failed")
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/recommendations/"+testUserID.String(), nil)
	rr := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Get("/recommendations/{userID}", apiHandler.GetRecommendations)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var recommendations []model.Recommendation
	err := json.NewDecoder(rr.Body).Decode(&recommendations)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if len(recommendations) == 0 {
		t.Errorf("expected some recommendations, but got none")
	}
}
