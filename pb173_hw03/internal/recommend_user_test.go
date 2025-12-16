package internal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"pb173_hw03/internal"
)

type recommendUserTestCase struct {
	Name         string
	File         string
	ExpectedCode int
}

func TestRecommendContentToUser(t *testing.T) {
	t.Parallel()

	tests := []recommendUserTestCase{
		{"NotFound", "testdata/testdataU/recommend_user_notfound.json", http.StatusNotFound},
		{"Valid", "testdata/testdataU/recommend_user_valid.json", http.StatusOK},
		{"OffsetLimit", "testdata/testdataU/recommend_user_paging.json", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			handler := internal.NewHandler()

			var testData struct {
				Request          map[string]any   `json:"request"`
				Preload          []map[string]any `json:"preload"`
				ExpectedResponse map[string]any   `json:"expectedResponse"`
			}

			loadJSONFile(t, tc.File, &testData)

			preloadReviews(t, handler, testData.Preload)

			userId, ok := testData.Request["userId"].(string)
			if !ok {
				t.Fatalf("expected string for userId but got %#v", testData.Request["userId"])
			}

			url := "/recommendations/users/" + userId + "/content"

			params := internal.RecommendContentToUserParams{}
			if q, ok := testData.Request["query"].(map[string]any); ok {
				var sb strings.Builder
				sb.WriteString("?")
				first := true
				for key, val := range q {
					if !first {
						sb.WriteString("&")
					}
					first = false
					sb.WriteString(key)
					sb.WriteString("=")
					sb.WriteString(fmt.Sprintf("%v", val))
				}
				url += sb.String()

				if limit, ok := q["limit"].(float64); ok {
					v := int(limit)
					params.Limit = &v
				}
				if offset, ok := q["offset"].(float64); ok {
					v := int(offset)
					params.Offset = &v
				}
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.RecommendContentToUser(w, req, userId, params)

			resp := w.Result()
			assert.Equal(t, tc.ExpectedCode, resp.StatusCode)

			var jsonResp map[string]any
			err := json.NewDecoder(resp.Body).Decode(&jsonResp)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			assert.Equal(t, testData.ExpectedResponse, jsonResp)
		})
	}
}

func preloadReviews(t *testing.T, handler *internal.HandlerImpl, preload []map[string]any) {
	t.Helper()

	for _, r := range preload {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("failed to marshal preload review: %v", err)
		}

		var raw internal.RawReview
		err = json.Unmarshal(data, &raw)
		if err != nil {
			t.Fatalf("failed to unmarshal preload review: %v", err)
		}

		review := internal.Review{
			Id:          uuid.NewString(),
			UserId:      raw.UserId,
			ContentId:   raw.ContentId,
			Review:      raw.Review,
			Title:       raw.Title,
			Actors:      raw.Actors,
			Director:    raw.Director,
			Description: raw.Description,
			Duration:    raw.Duration,
			Genres:      raw.Genres,
			Origins:     raw.Origins,
			Released:    raw.Released,
			Score:       raw.Score,
			Tags:        raw.Tags,
		}

		handler.Storage.AddReview(review)
	}
}
