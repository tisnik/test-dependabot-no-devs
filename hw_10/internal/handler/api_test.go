package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/handler"
	"github.com/course-go/reelgoofy/internal/handler/dto"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

func setupRouter() (*chi.Mux, *repository.ReviewRepository) {
	repo := repository.NewReviewRepository()
	reviewService := service.NewReviewService(repo)
	recommendationService := service.NewRecommendationService(repo)

	reviewHandler := handler.NewReviewHandler(reviewService)
	recommendationsHandler := handler.NewRecommendationsHandler(recommendationService)

	r := chi.NewRouter()
	r.Route("/reviews", func(r chi.Router) {
		r.Post("/", reviewHandler.CreateReview)
		r.Delete("/{id}", reviewHandler.DeleteReview)
	})
	r.Route("/recommendations", func(r chi.Router) {
		r.Get("/users/{userId}/content", recommendationsHandler.GetRecommendationsByUser)
		r.Get("/content/{contentId}/content", recommendationsHandler.GetRecommendationsByContent)
	})
	return r, repo
}

func TestCreateReviews(t *testing.T) {
	t.Parallel()
	router, _ := setupRouter()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		review := dto.ReviewDTO{
			ContentID:   uuid.New().String(),
			UserID:      uuid.New().String(),
			Title:       "Test Movie",
			Genres:      []string{"Action"},
			Tags:        []string{"Fun"},
			Director:    "Test Director",
			Actors:      []string{"Actor A"},
			Origins:     []string{"USA"},
			Duration:    120,
			Released:    "2023-01-01",
			ReviewText:  "Great movie",
			Score:       85,
			Description: "A test movie description",
		}

		reqBody := dto.CreateReviewRequest{
			Data: dto.CreateReviewData{
				Reviews: []dto.ReviewDTO{review},
			},
		}

		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/reviews", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var resp dto.CreateReviewResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Status != statusSuccess {
			t.Errorf("expected status success, got %s", resp.Status)
		}
		if len(resp.Data.Reviews) != 1 {
			t.Errorf("expected 1 review, got %d", len(resp.Data.Reviews))
		}
		if resp.Data.Reviews[0].Id == "" {
			t.Error("expected review ID to be set")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/reviews",
			bytes.NewBufferString("invalid json"),
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Errorf("failed to unmarshal fail response: %v. Body: %s", err, w.Body.String())
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		t.Parallel()
		review := dto.ReviewDTO{
			UserID:      uuid.New().String(),
			Title:       "Test Movie",
			Score:       85,
			Description: "Description",
		}
		reqBody := dto.CreateReviewRequest{
			Data: dto.CreateReviewData{
				Reviews: []dto.ReviewDTO{review},
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/reviews", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Error("expected data to be a map")
		}
		if _, exists := data["ContentID"]; !exists {
			t.Error("expected validation error for ContentID")
		}
	})
}

func TestDeleteReview(t *testing.T) {
	t.Parallel()
	router, repo := setupRouter()

	review := domain.Review{
		Id:        uuid.New(),
		ContentId: uuid.New(),
		UserId:    uuid.New(),
		Title:     "Test Movie",
		Score:     10,
	}
	saved, _ := repo.SaveBatch([]domain.Review{review})
	reviewId := saved[0].Id

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			"/reviews/"+reviewId.String(),
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			"/reviews/"+reviewId.String(),
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Error("expected data to be a map")
		}
		if _, exists := data["reviewId"]; !exists {
			t.Error("expected error message for reviewId")
		}
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/reviews/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}

		if resp.Data == nil {
			t.Error("expected data field in fail response")
		}
	})
}

func TestGetRecommendationsByUser(t *testing.T) {
	t.Parallel()
	router, repo := setupRouter()

	userId := uuid.New()
	contentId := uuid.New()

	reviews := []domain.Review{
		{
			Id:        uuid.New(),
			ContentId: contentId,
			UserId:    userId,
			Title:     "User's Favorite",
			Director:  "Director A",
			Genres:    []string{"Drama"},
			Score:     90,
		},
		{
			Id:        uuid.New(),
			ContentId: uuid.New(),
			UserId:    uuid.New(), // Different user
			Title:     "Recommended Movie",
			Director:  "Director A", // Same director
			Genres:    []string{"Action"},
			Score:     95,
		},
	}
	_, _ = repo.SaveBatch(reviews)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/users/"+userId.String()+"/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp dto.RecommendationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Status != statusSuccess {
			t.Errorf("expected status success, got %s", resp.Status)
		}
		if len(resp.Data.Recommendations) == 0 {
			t.Error("expected recommendations, got none")
		} else if resp.Data.Recommendations[0].Title != "Recommended Movie" {
			t.Errorf("expected title 'Recommended Movie', got '%s'", resp.Data.Recommendations[0].Title)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/users/"+uuid.New().String()+"/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Error("expected data to be a map")
		}
		if _, exists := data["userId"]; !exists {
			t.Error("expected error message for userId")
		}
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/users/invalid-uuid/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		if resp.Data == nil {
			t.Error("expected data field in fail response")
		}
	})
}

func TestGetRecommendationsByContent(t *testing.T) {
	t.Parallel()
	router, repo := setupRouter()
	contentId := uuid.New()

	reviews := []domain.Review{
		{
			Id:        uuid.New(),
			ContentId: contentId,
			UserId:    uuid.New(),
			Title:     "Source Movie",
			Director:  "Director B",
			Genres:    []string{"Sci-Fi"},
			Score:     90,
		},
		{
			Id:        uuid.New(),
			ContentId: uuid.New(),
			UserId:    uuid.New(),
			Title:     "Similar Movie",
			Director:  "Director B", // Same director
			Genres:    []string{"Action"},
			Score:     80,
		},
	}
	_, _ = repo.SaveBatch(reviews)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/content/"+contentId.String()+"/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp dto.RecommendationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Status != statusSuccess {
			t.Errorf("expected status success, got %s", resp.Status)
		}
		if len(resp.Data.Recommendations) == 0 {
			t.Error("expected recommendations, got none")
		} else if resp.Data.Recommendations[0].Title != "Similar Movie" {
			t.Errorf("expected title 'Similar Movie', got '%s'", resp.Data.Recommendations[0].Title)
		}
	})

	t.Run("ContentNotFound", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/content/"+uuid.New().String()+"/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Error("expected data to be a map")
		}
		if _, exists := data["contentId"]; !exists {
			t.Error("expected error message for contentId")
		}
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/recommendations/content/invalid-uuid/content",
			nil,
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp dto.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to unmarshal fail response: %v", err)
		}
		if resp.Status != statusFail {
			t.Errorf("expected status fail, got %s", resp.Status)
		}
		if resp.Data == nil {
			t.Error("expected data field in fail response")
		}
	})
}
