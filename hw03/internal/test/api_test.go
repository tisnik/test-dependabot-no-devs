package test_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Markek1/reelgoofy/internal/api"
	"github.com/Markek1/reelgoofy/internal/domain"
	"github.com/Markek1/reelgoofy/internal/repository"
	"github.com/Markek1/reelgoofy/internal/service"
	"github.com/google/uuid"
)

func setupRouter() http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo := repository.NewMemoryRepository()
	svc := service.NewService(repo, logger)
	return api.NewRouter(svc, logger)
}

func TestIngestReviews(t *testing.T) {
	t.Parallel()
	router := setupRouter()

	c1 := uuid.NewString()
	u1 := uuid.NewString()

	reqBody := fmt.Sprintf(`
	{
		"data": {
			"reviews": [
				{
					"contentId": "%s",
					"userId": "%s",
					"review": "Good",
					"score": 80,
					"genres": ["Action"]
				}
			]
		}
	}`, c1, u1)

	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var resp domain.ReviewsSuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != domain.StatusSuccess {
		t.Errorf("Expected status success, got %s", resp.Status)
	}

	if len(resp.Data.Reviews) != 1 {
		t.Errorf("Expected 1 review, got %d", len(resp.Data.Reviews))
	}
}

func TestRecommendContent(t *testing.T) {
	t.Parallel()
	router := setupRouter()

	movieA := uuid.NewString()
	movieB := uuid.NewString()
	movieC := uuid.NewString()
	u1 := uuid.NewString()
	u2 := uuid.NewString()
	u3 := uuid.NewString()

	reqBody := fmt.Sprintf(`
	{
		"data": {
			"reviews": [
				{"contentId": "%s", "userId": "%s", "review": "A", "score": 80, "genres": ["Action"]},
				{"contentId": "%s", "userId": "%s", "review": "B", "score": 80, "genres": ["Action"]},
				{"contentId": "%s", "userId": "%s", "review": "C", "score": 80, "genres": ["Romance"]}
			]
		}
	}`, movieA, u1, movieB, u2, movieC, u3)

	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewBufferString(reqBody))
	router.ServeHTTP(httptest.NewRecorder(), req)

	reqRec := httptest.NewRequest(http.MethodGet, "/recommendations/content/"+movieA+"/content", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, reqRec)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp domain.RecommendationsSuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Data.Recommendations) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(resp.Data.Recommendations))
	}

	if resp.Data.Recommendations[0].ID != domain.ContentID(movieB) {
		t.Errorf("Expected recommendation %s, got %s", movieB, resp.Data.Recommendations[0].ID)
	}
}

func TestRecommendUser(t *testing.T) {
	t.Parallel()
	router := setupRouter()

	movieA := uuid.NewString()
	movieB := uuid.NewString()
	u1 := uuid.NewString()
	u2 := uuid.NewString()

	reqBody := fmt.Sprintf(`
	{
		"data": {
			"reviews": [
				{"contentId": "%s", "userId": "%s", "review": "A", "score": 90, "genres": ["Action"]},
				{"contentId": "%s", "userId": "%s", "review": "B", "score": 80, "genres": ["Action"]}
			]
		}
	}`, movieA, u1, movieB, u2)

	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewBufferString(reqBody))
	router.ServeHTTP(httptest.NewRecorder(), req)

	reqRec := httptest.NewRequest(http.MethodGet, "/recommendations/users/"+u1+"/content", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, reqRec)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp domain.RecommendationsSuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Data.Recommendations) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(resp.Data.Recommendations))
	}

	if resp.Data.Recommendations[0].ID != domain.ContentID(movieB) {
		t.Errorf("Expected recommendation %s, got %s", movieB, resp.Data.Recommendations[0].ID)
	}
}

func TestDeleteReview(t *testing.T) {
	t.Parallel()
	router := setupRouter()

	c1 := uuid.NewString()
	u1 := uuid.NewString()

	reqBody := fmt.Sprintf(
		`{"data": {"reviews": [{"contentId": "%s", "userId": "%s", "review": "R", "score": 50}]}}`,
		c1,
		u1,
	)
	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp domain.ReviewsSuccessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	reviewID := resp.Data.Reviews[0].ID

	reqDel := httptest.NewRequest(http.MethodDelete, "/reviews/"+string(reviewID), nil)
	recDel := httptest.NewRecorder()
	router.ServeHTTP(recDel, reqDel)

	if recDel.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recDel.Code)
	}

	reqDel2 := httptest.NewRequest(http.MethodDelete, "/reviews/"+string(reviewID), nil)
	recDel2 := httptest.NewRecorder()
	router.ServeHTTP(recDel2, reqDel2)

	if recDel2.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", recDel2.Code)
	}
}
