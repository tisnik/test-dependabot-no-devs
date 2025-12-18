package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal/api"
	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/google/uuid"
)

func newRecommendationsTestServer() http.Handler {
	return api.NewRouter(repository.NewInMemoryRepository())
}

func TestRecommendContentToContentSuccess(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	c1 := uuid.New().String()
	c2 := uuid.New().String()
	c3 := uuid.New().String()
	u1 := uuid.New().String()
	ingest := model.RawReviewsRequest{}
	ingest.Data.Reviews = []model.RawReview{
		{
			ContentID: c1,
			UserID:    u1,
			Title:     "Alpha",
			Genres:    []string{"drama"},
			Tags:      []string{"friendship"},
			Review:    "Good",
			Score:     80,
			Released:  "2024-01-01",
		},
		{
			ContentID: c2,
			UserID:    u1,
			Title:     "Beta",
			Genres:    []string{"drama"},
			Tags:      []string{"friendship"},
			Review:    "Nice",
			Score:     75,
			Released:  "2024-01-02",
		},
		{
			ContentID: c3,
			UserID:    u1,
			Title:     "Gamma",
			Genres:    []string{"sci-fi"},
			Tags:      []string{"space"},
			Review:    "Ok",
			Score:     60,
			Released:  "2024-01-03",
		},
	}
	body, err := json.Marshal(ingest)
	if err != nil {
		t.Fatalf("failed to marshal ingest: %v", err)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(body)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("seed failed %d", rr.Code)
	}
	recRR := httptest.NewRecorder()
	h.ServeHTTP(
		recRR,
		httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+c1+"/content?limit=5", nil),
	)
	if recRR.Code != http.StatusOK {
		t.Errorf("expected 200 got %d", recRR.Code)
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Recommendations []struct {
				ID string `json:"id"`
			} `json:"recommendations"`
		} `json:"data"`
	}
	_ = json.Unmarshal(recRR.Body.Bytes(), &resp)
	if resp.Status != "success" {
		t.Errorf("expected success got %s", resp.Status)
	}
	if len(resp.Data.Recommendations) == 0 {
		t.Errorf("expected at least one recommendation")
	}
}

func TestRecommendContentToContentInvalidUUID(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/not-a-uuid/content", nil))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 got %d", rr.Code)
	}
}

func TestRecommendContentToContentNotFound(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	missing := uuid.New().String()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/"+missing+"/content", nil))
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
}

func TestRecommendContentToUserSuccess(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	u := uuid.New().String()
	c1 := uuid.New().String()
	c2 := uuid.New().String()
	c3 := uuid.New().String()
	ingest := model.RawReviewsRequest{}
	ingest.Data.Reviews = []model.RawReview{
		{
			ContentID: c1,
			UserID:    u,
			Title:     "Watched1",
			Genres:    []string{"thriller"},
			Tags:      []string{"dark"},
			Review:    "Great",
			Score:     90,
			Released:  "2024-02-01",
		},
		{
			ContentID: c2,
			UserID:    uuid.New().String(),
			Title:     "Cand2",
			Genres:    []string{"thriller"},
			Tags:      []string{"dark"},
			Review:    "Good",
			Score:     80,
			Released:  "2024-02-02",
		},
		{
			ContentID: c3,
			UserID:    uuid.New().String(),
			Title:     "Cand3",
			Genres:    []string{"comedy"},
			Tags:      []string{"fun"},
			Review:    "Okay",
			Score:     70,
			Released:  "2024-02-03",
		},
	}
	body, err := json.Marshal(ingest)
	if err != nil {
		t.Fatalf("failed to marshal ingest: %v", err)
	}
	seedRR := httptest.NewRecorder()
	h.ServeHTTP(seedRR, httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(body)))
	if seedRR.Code != http.StatusCreated {
		t.Fatalf("seed failed %d", seedRR.Code)
	}
	recRR := httptest.NewRecorder()
	h.ServeHTTP(recRR, httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/"+u+"/content?limit=5", nil))
	if recRR.Code != http.StatusOK {
		t.Errorf("expected 200 got %d", recRR.Code)
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Recommendations []struct {
				ID string `json:"id"`
			} `json:"recommendations"`
		} `json:"data"`
	}
	_ = json.Unmarshal(recRR.Body.Bytes(), &resp)
	if resp.Status != "success" {
		t.Errorf("expected success got %s", resp.Status)
	}
	for _, r := range resp.Data.Recommendations {
		if r.ID == c1 {
			t.Errorf("should not recommend already reviewed content")
		}
	}
}

func TestRecommendContentToUserInvalidUUID(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/not-a-uuid/content", nil))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 got %d", rr.Code)
	}
}

func TestRecommendContentToUserNotFound(t *testing.T) {
	t.Parallel()
	h := newRecommendationsTestServer()
	missing := uuid.New().String()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/users/"+missing+"/content", nil))
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
}
