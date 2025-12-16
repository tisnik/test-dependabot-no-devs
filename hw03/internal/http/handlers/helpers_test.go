package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apphttp "github.com/course-go/reelgoofy/internal/http"
	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/recommendations"
	"github.com/course-go/reelgoofy/internal/reviews"
	"github.com/course-go/reelgoofy/internal/status"
)

const testdataRoot = "../../utils/test/testdata"

func setupComponents(t *testing.T) (*reviews.MemoryReviewRepository, http.Handler) {
	t.Helper()

	repo := reviews.NewMemoryReviewRepository()
	revSvc := reviews.NewReviewService(repo)
	recSvc := recommendations.NewRecommendationService(repo)
	r := apphttp.NewRouter(revSvc, recSvc)

	return repo, r
}

func setupRecommendationTest(t *testing.T, jsonPath string) http.Handler {
	t.Helper()

	repo := reviews.NewMemoryReviewRepository()
	revSvc := reviews.NewReviewService(repo)
	recSvc := recommendations.NewRecommendationService(repo)
	r := apphttp.NewRouter(revSvc, recSvc)

	fullPath := filepath.Join(testdataRoot, jsonPath)
	fullPath = filepath.Clean(fullPath)
	if !strings.HasPrefix(fullPath, testdataRoot) {
		t.Fatalf("failed to setup recommendations testdataRoot: %s", fullPath)
	}

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("error reading file %s: %v", fullPath, err)
	}

	var req dto.RawReviewsRequest
	err = json.Unmarshal(bytes, &req)
	if err != nil {
		t.Fatal("Error unmarshalling", jsonPath)
	}

	_, err = revSvc.Ingest(req.Data.Reviews)
	if err != nil {
		t.Fatal("Error ingesting", jsonPath)
	}

	return r
}

func executeRequest(r http.Handler, method, url string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func assertSuccess[T any](t *testing.T, rec *httptest.ResponseRecorder) dto.Response[T] {
	t.Helper()
	return assertResponse[T](t, rec, http.StatusOK, status.Success)
}

func assertCreated[T any](t *testing.T, rec *httptest.ResponseRecorder) dto.Response[T] {
	t.Helper()
	return assertResponse[T](t, rec, http.StatusCreated, status.Success)
}

func assertFail(t *testing.T, rec *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()
	assertResponse[dto.FailDataDTO](t, rec, expectedCode, status.Fail)
}

func assertResponse[T any](
	t *testing.T,
	rec *httptest.ResponseRecorder,
	expectedCode int,
	expectedStatus status.Status,
) dto.Response[T] {
	t.Helper()

	if rec.Code != expectedCode {
		t.Fatalf("expected code %d, got %d", expectedCode, rec.Code)
	}

	var resp dto.Response[T]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, body: %s", err, rec.Body.String())
	}

	if resp.Status != expectedStatus {
		t.Fatalf("expected status %s, got %s", status.Success, resp.Status)
	}

	return resp
}
