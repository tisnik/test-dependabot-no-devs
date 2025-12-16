package internal_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"pb173_hw03/internal"
)

type ingestTestCase struct {
	Name         string
	File         string
	ExpectedCode int
}

//nolint:gocognit
func TestIngestReviews(t *testing.T) {
	t.Parallel()

	tests := []ingestTestCase{
		{"Valid", "testdata/testdataI/review_valid.json", http.StatusCreated},
		{"InvalidActor", "testdata/testdataI/review_invalid_actor.json", http.StatusBadRequest},
		{"InvalidGenre", "testdata/testdataI/review_invalid_genre.json", http.StatusBadRequest},
		{"InvalidTag", "testdata/testdataI/review_invalid_tag.json", http.StatusBadRequest},
		{"InvalidOrigin", "testdata/testdataI/review_invalid_origin.json", http.StatusBadRequest},
		{"InvalidReleased", "testdata/testdataI/review_invalid_released.json", http.StatusBadRequest},
		{"InvalidScore", "testdata/testdataI/review_invalid_score.json", http.StatusBadRequest},
		{"MissingReview", "testdata/testdataI/review_missing_review.json", http.StatusBadRequest},
		{"InvalidUserID", "testdata/testdataI/review_invalid_userid.json", http.StatusBadRequest},
		{"InvalidContentID", "testdata/testdataI/review_invalid_contentid.json", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			handler := internal.NewHandler()

			var testData struct {
				Request          map[string]any `json:"request"`
				ExpectedResponse map[string]any `json:"expectedResponse"`
			}

			loadJSONFile(t, tc.File, &testData)

			body, err := json.Marshal(testData.Request)
			if err != nil {
				t.Fatalf("error should nil but got %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.IngestReviews(w, req)

			resp := w.Result()
			assert.Equal(t, tc.ExpectedCode, resp.StatusCode)

			var jsonResp map[string]any
			err = json.NewDecoder(resp.Body).Decode(&jsonResp)
			if err != nil {
				t.Fatalf("error should nil but got %v", err)
			}

			if testData.ExpectedResponse["status"] != "success" {
				return
			}

			dataMap, ok := jsonResp["data"].(map[string]any)
			if !ok {
				t.Fatal(`jsonResp["data"] is not a map`)
			}

			reviewsArr, ok := dataMap["reviews"].([]any)
			if !ok || len(reviewsArr) == 0 {
				t.Fatal(`data["reviews"] is not a non-empty array`)
			}

			firstReview, ok := reviewsArr[0].(map[string]any)
			if !ok {
				t.Fatal("first review is not a map")
			}

			actualID, ok := firstReview["id"].(string)
			if !ok {
				t.Fatal("id field is not a string")
			}

			expDataMap, ok := testData.ExpectedResponse["data"].(map[string]any)
			if !ok {
				t.Fatal(`ExpectedResponse["data"] is not a map`)
			}

			expReviewsArr, ok := expDataMap["reviews"].([]any)
			if !ok || len(expReviewsArr) == 0 {
				t.Fatal(`ExpectedResponse["reviews"] is not a non-empty array`)
			}

			expReview, ok := expReviewsArr[0].(map[string]any)
			if !ok {
				t.Fatal("first expected review is not a map")
			}

			expReview["id"] = actualID

			assert.Equal(t, testData.ExpectedResponse, jsonResp)
		})
	}
}

//nolint:gosec
func loadJSONFile(t *testing.T, path string, v any) {
	t.Helper() // označí túto funkciu ako pomocnú pre testy

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open JSON file %q: %v", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	err = json.NewDecoder(f).Decode(v)
	if err != nil {
		t.Fatalf("failed to decode JSON file %q: %v", path, err)
	}
}
