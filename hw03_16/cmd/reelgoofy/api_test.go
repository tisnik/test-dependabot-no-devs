package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestServer() http.Handler {
	server := NewServer()
	r := chi.NewRouter()
	HandlerFromMuxWithBaseURL(server, r, "/api/v1")
	return r
}

func makeRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func ptr[T any](v T) *T {
	return &v
}

func TestIngestReviews_Valid(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Title:     ptr("The Matrix"),
					Genres:    &[]string{"sci-fi", "action"},
					Review:    "Amazing!",
					Score:     95,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	reviews, ok := data["reviews"].([]any)
	require.True(t, ok)
	assert.Len(t, reviews, 1)

	review, ok := reviews[0].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, review["id"])
	assert.Equal(t, "937b33bf-066a-44f7-9a9b-d65071d27270", review["contentId"])
}

func TestIngestReviews_Multiple(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "Good",
					Score:     90,
				},
				{
					ContentId: "75fe91a7-9ebc-4029-97f7-9d99a059348d",
					UserId:    "1668d2a8-6344-4389-9797-c381c55b080b",
					Review:    "Great",
					Score:     85,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestIngestReviews_InvalidContentID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "invalid-uuid",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "Test",
					Score:     80,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "fail", response["status"])
}

func TestIngestReviews_InvalidUserID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "invalid",
					Review:    "Test",
					Score:     80,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteReview_Valid(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "Test",
					Score:     80,
				},
			},
		},
	}

	createRR := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	require.Equal(t, http.StatusCreated, createRR.Code)

	var createResp map[string]any
	_ = json.Unmarshal(createRR.Body.Bytes(), &createResp)

	data, ok := createResp["data"].(map[string]any)
	require.True(t, ok)
	reviews, ok := data["reviews"].([]any)
	require.True(t, ok)
	review, ok := reviews[0].(map[string]any)
	require.True(t, ok)
	reviewID, ok := review["id"].(string)
	require.True(t, ok)

	deleteRR := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/"+reviewID, nil)
	assert.Equal(t, http.StatusOK, deleteRR.Code)
}

func TestDeleteReview_InvalidUUID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/invalid", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteReview_NotFound(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/733b9f08-710d-4abf-93f9-7353ed2b4e08", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRecommendContentToContent_Valid(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "4dd103ad-54d5-4bd2-a90e-57ac52e0259a",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Genres:    &[]string{"sci-fi"},
					Review:    "Great",
					Score:     90,
				},
				{
					ContentId: "75fe91a7-9ebc-4029-97f7-9d99a059348d",
					UserId:    "1668d2a8-6344-4389-9797-c381c55b080b",
					Genres:    &[]string{"sci-fi"},
					Review:    "Good",
					Score:     85,
				},
			},
		},
	}

	createRR := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	require.Equal(t, http.StatusCreated, createRR.Code)

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/content/4dd103ad-54d5-4bd2-a90e-57ac52e0259a/content",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	recommendations, ok := data["recommendations"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, recommendations, "Should recommend similar content")
}

func TestRecommendContentToContent_InvalidUUID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(t, handler, http.MethodGet, "/api/v1/recommendations/content/invalid/content", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRecommendContentToUser_Valid(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa9a/content",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRecommendContentToUser_InvalidUUID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(t, handler, http.MethodGet, "/api/v1/recommendations/users/invalid/content", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIngestReviews_EmptyList(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.True(t, rr.Code == http.StatusCreated || rr.Code == http.StatusBadRequest)
}

func TestIngestReviews_EmptyReviewText(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "",
					Score:     80,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIngestReviews_MissingContentId(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "Test",
					Score:     80,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIngestReviews_MultipleBatch(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reviews := []RawReview{}
	for i := range 5 {
		reviews = append(reviews, RawReview{
			ContentId: "937b33bf-066a-44f7-9a9b-d65071d2727" + string(rune('0'+i)),
			UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
			Review:    "Movie review",
			Score:     80 + i,
		})
	}

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &reviews,
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	reviewsList, ok := data["reviews"].([]any)
	require.True(t, ok)
	assert.Len(t, reviewsList, 5)
}

func TestDeleteReview_Twice(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "Test",
					Score:     80,
				},
			},
		},
	}

	createRR := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	require.Equal(t, http.StatusCreated, createRR.Code)

	var createResp map[string]any
	_ = json.Unmarshal(createRR.Body.Bytes(), &createResp)

	data, ok := createResp["data"].(map[string]any)
	require.True(t, ok)
	reviewsList, ok := data["reviews"].([]any)
	require.True(t, ok)
	review, ok := reviewsList[0].(map[string]any)
	require.True(t, ok)
	reviewID, ok := review["id"].(string)
	require.True(t, ok)

	deleteRR := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/"+reviewID, nil)
	assert.Equal(t, http.StatusOK, deleteRR.Code)

	deleteRR2 := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/"+reviewID, nil)
	assert.Equal(t, http.StatusNotFound, deleteRR2.Code)
}

func TestRecommendations_EmptyWhenNoData(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/content/4dd103ad-54d5-4bd2-a90e-57ac52e0259a/content",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	recommendations, ok := data["recommendations"].([]any)
	require.True(t, ok)
	assert.Empty(t, recommendations)
}

func TestRecommendations_WithPagination(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/content/4dd103ad-54d5-4bd2-a90e-57ac52e0259a/content?limit=10&offset=0",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestIngestReviews_UniqueIDs(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId: "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:    "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Review:    "First",
					Score:     90,
				},
				{
					ContentId: "75fe91a7-9ebc-4029-97f7-9d99a059348d",
					UserId:    "1668d2a8-6344-4389-9797-c381c55b080b",
					Review:    "Second",
					Score:     85,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	require.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatal("data is not a map")
	}
	reviewsList, ok := data["reviews"].([]any)
	if !ok {
		t.Fatal("reviewsList is not a slice")
	}

	review1, ok := reviewsList[0].(map[string]any)
	if !ok {
		t.Fatal("review1 is not a map")
	}
	review2, ok := reviewsList[1].(map[string]any)
	if !ok {
		t.Fatal("review2 is not a map")
	}

	id1, ok := review1["id"].(string)
	if !ok {
		t.Fatal("review1 id is not a string")
	}
	id2, ok := review2["id"].(string)
	if !ok {
		t.Fatal("review2 id is not a string")
	}

	assert.NotEqual(t, id1, id2, "IDs should be unique")
}

func TestDeleteReview_EmptyID(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(t, handler, http.MethodDelete, "/api/v1/reviews/", nil)
	assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusBadRequest)
}

func TestIngestReviews_WithAllFields(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	reqBody := RawReviewsRequest{
		Data: &RawReviews{
			Reviews: &[]RawReview{
				{
					ContentId:   "937b33bf-066a-44f7-9a9b-d65071d27270",
					UserId:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
					Title:       ptr("Inception"),
					Genres:      &[]string{"sci-fi", "thriller"},
					Tags:        &[]string{"dreams", "complex"},
					Description: ptr("A thief who steals corporate secrets"),
					Director:    ptr("Christopher Nolan"),
					Actors:      &[]string{"Leonardo DiCaprio", "Tom Hardy"},
					Origins:     &[]string{"USA", "UK"},
					Duration:    ptr(8880),
					Released:    ptr("2010-07-16"),
					Review:      "Mind-blowing movie!",
					Score:       95,
				},
			},
		},
	}

	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/reviews", reqBody)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	reviewsList, ok := data["reviews"].([]any)
	require.True(t, ok)
	review, ok := reviewsList[0].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "Inception", review["title"])
	assert.Equal(t, "Christopher Nolan", review["director"])
}

func TestRecommendContentToUser_WithPagination(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa9a/content?limit=5&offset=0",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRecommendContentToContent_NonExistent(t *testing.T) {
	t.Parallel()
	handler := createTestServer()

	rr := makeRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/recommendations/content/99999999-0000-0000-0000-000000000000/content",
		nil,
	)
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok)
	recommendations, ok := data["recommendations"].([]any)
	require.True(t, ok)
	assert.Empty(t, recommendations)
}
