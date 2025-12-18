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
	"github.com/google/uuid"
)

// TestCreateReview tests the ingest endpoint for reviews.
func TestCreateReview(t *testing.T) {
	t.Parallel()

	memStorage := storage.NewMemoryStorage()
	reviewService := service.NewReviewService(memStorage)
	apiHandler := handler.NewHandler(reviewService, nil)

	review := model.Review{
		UserID:  uuid.New(),
		MovieID: uuid.New(),
		Rating:  8,
	}
	body, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("failed to marshal review: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	apiHandler.CreateReview(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}
