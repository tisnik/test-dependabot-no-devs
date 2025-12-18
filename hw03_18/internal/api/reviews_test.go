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

const statusSuccess = "success"

func newTestServer() http.Handler {
	repo := repository.NewInMemoryRepository()
	return api.NewRouter(repo)
}

func TestIngestReviewsSuccess(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	contentID := uuid.New().String()
	userID := uuid.New().String()
	reqBody := model.RawReviewsRequest{}
	reqBody.Data.Reviews = []model.RawReview{{
		ContentID: contentID,
		UserID:    userID,
		Title:     "Test Movie",
		Review:    "Pretty good",
		Score:     85,
		Released:  "2024-10-10",
	}}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(buf))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", rr.Code)
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Reviews []struct {
				ID string `json:"id"`
			} `json:"reviews"`
		} `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("invalid JSON response: %v", err)
	}
	if resp.Status != statusSuccess {
		t.Errorf("expected status success, got %s", resp.Status)
	}
	if len(resp.Data.Reviews) != 1 {
		t.Errorf("expected 1 review returned, got %d", len(resp.Data.Reviews))
	}
	if resp.Data.Reviews[0].ID == "" {
		t.Errorf("expected generated review id to be non-empty")
	}
}

func TestIngestReviewsInvalidUUIDs(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	reqBody := `{"data":{"reviews":[{"contentId":"not-a-uuid","userId":"also-bad","review":"r","score":10}]}}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(reqBody))
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
	var resp struct {
		Status string            `json:"status"`
		Data   map[string]string `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("invalid JSON response: %v", err)
	}
	if resp.Status != "fail" {
		t.Errorf("expected status fail, got %s", resp.Status)
	}
	if _, ok := resp.Data["contentId"]; !ok {
		t.Errorf("expected contentId validation error present")
	}
	if _, ok := resp.Data["userId"]; !ok {
		t.Errorf("expected userId validation error present")
	}
}

func TestIngestReviewsInvalidDate(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	contentID := uuid.New().String()
	userID := uuid.New().String()
	reqBody := model.RawReviewsRequest{}
	reqBody.Data.Reviews = []model.RawReview{{
		ContentID: contentID,
		UserID:    userID,
		Review:    "Bad date",
		Score:     50,
		Released:  "2024-13-99",
	}}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(buf))
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestIngestReviewsEmptyList(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	reqBody := `{"data":{"reviews":[]}}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(reqBody))
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty reviews list, got %d", rr.Code)
	}
}

func TestDeleteReviewInvalidUUID(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/not-a-uuid", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid uuid, got %d", rr.Code)
	}
}

func TestDeleteReviewUnknown(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	unknown := uuid.New().String()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+unknown, nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for unknown review id, got %d", rr.Code)
	}
}

func TestDeleteReviewSuccess(t *testing.T) {
	t.Parallel()
	handler := newTestServer()
	contentID := uuid.New().String()
	userID := uuid.New().String()
	reqBody := model.RawReviewsRequest{}
	reqBody.Data.Reviews = []model.RawReview{{
		ContentID: contentID,
		UserID:    userID,
		Review:    "Nice",
		Score:     90,
		Released:  "2024-10-10",
	}}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	ingestRR := httptest.NewRecorder()
	ingestReq := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(buf))
	handler.ServeHTTP(ingestRR, ingestReq)
	if ingestRR.Code != http.StatusCreated {
		t.Fatalf("expected initial ingest 201, got %d", ingestRR.Code)
	}
	// Parse returned ID.
	var resp struct {
		Data struct {
			Reviews []struct {
				ID string `json:"id"`
			} `json:"reviews"`
		} `json:"data"`
	}
	_ = json.Unmarshal(ingestRR.Body.Bytes(), &resp)
	if len(resp.Data.Reviews) == 0 || resp.Data.Reviews[0].ID == "" {
		t.Fatalf("expected returned review with id")
	}
	id := resp.Data.Reviews[0].ID
	delRR := httptest.NewRecorder()
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+id, nil)
	handler.ServeHTTP(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Errorf("expected 200 OK on delete, got %d", delRR.Code)
	}
}
