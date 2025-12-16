package internal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"pb173_hw03/internal"
)

type recommendContentTestCase struct {
	Name         string
	File         string
	ExpectedCode int
}

func TestRecommendContentToContent(t *testing.T) {
	t.Parallel()

	tests := []recommendContentTestCase{
		{"NotFound", "testdata/testdataC/recommend_content_notfound.json", http.StatusNotFound},
		{"Valid", "testdata/testdataC/recommend_content_valid.json", http.StatusOK},
		{"Paging", "testdata/testdataC/recommend_content_paging.json", http.StatusOK},
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

			contentId, ok := testData.Request["contentId"].(string)
			if !ok {
				t.Fatalf("expected string for contentId but got %v", testData.Request["contentId"])
			}
			url := "/recommendations/content/" + contentId + "/content"

			params := internal.RecommendContentToContentParams{}
			if q, ok := testData.Request["query"].(map[string]any); ok && len(q) > 0 {
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

			handler.RecommendContentToContent(w, req, contentId, params)

			resp := w.Result()
			assert.Equal(t, tc.ExpectedCode, resp.StatusCode)

			var jsonResp map[string]any
			err := json.NewDecoder(resp.Body).Decode(&jsonResp)
			if err != nil {
				t.Fatalf("error should nil but got %v", err)
			}

			assert.Equal(t, testData.ExpectedResponse, jsonResp)
		})
	}
}
