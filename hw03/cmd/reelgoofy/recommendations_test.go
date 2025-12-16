package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RecResponse struct {
	Status string `json:"status"`
	Data   struct {
		Recommendations []struct {
			ContentID string `json:"contentId"`
			Title     string `json:"title"`
		} `json:"recommendations"`
	} `json:"data"`
}

func seedReviews(t *testing.T, e *echo.Echo, reviewsJSON string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBufferString(reviewsJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "Failed to seed data")
}

// -------------------------------------------------------------------------
// Content-to-Content Tests
// -------------------------------------------------------------------------

func TestRecommendContent_Success(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	payload := `{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-1111-1111-1111-111111111111",
					"title": "Interstellar",
					"genres": ["Sci-Fi"],
					"director": "Christopher Nolan",
					"score": 90,
					"review": "Cool"
				},
				{
					"contentId": "aaaaaaaa-bbbb-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "11111111-2222-1111-1111-111111111111",
					"title": "The Matrix",
					"genres": ["Sci-Fi"],
					"director": "Lilly Wachowski",
					"score": 91,
					"review": "Cool"
				},
				{
					"contentId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "22222222-2222-2222-2222-222222222222",
					"title": "Oppenheimer",
					"genres": ["Historical"],
					"director": "Christopher Nolan",
					"score": 90,
					"review": "Boooom"
				},
				{
					"contentId": "bbbbbbbb-cccc-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "22222222-3333-2222-2222-222222222222",
					"title": "Inception",
					"genres": ["Sci-Fi"],
					"director": "Christopher Nolan",
					"score": 90,
					"review": "Dreamy"
				}
			]
		}
	}`
	seedReviews(t, e, payload)

	targetURL := "/api/v1/recommendations/content/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/content"
	req := httptest.NewRequest(http.MethodGet, targetURL, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp RecResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, "success", resp.Status)

	// Inception - score = 90; similarity = 5 (director) + 4 (genre); score + similarity = 99
	assert.Equal(t, "Inception", resp.Data.Recommendations[0].Title)

	// Oppenheimer - score = 90; similarity = 5 (director); score + similarity = 95 (higher because higher similarity)
	assert.Equal(t, "Oppenheimer", resp.Data.Recommendations[1].Title)

	// The Matrix - score = 91; similarity = 4 (genre); score + similarity = 95
	assert.Equal(t, "The Matrix", resp.Data.Recommendations[2].Title)
}

func TestRecommendContent_NotFound(t *testing.T) {
	t.Parallel()
	e := SetupServer() // Empty DB

	// Valid UUID, but no reviews exist for it
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/recommendations/content/ffffffff-ffff-ffff-ffff-ffffffffffff/content",
		nil,
	)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "fail", resp.Status)
	assert.Equal(t, "Content with such ID not found.", resp.Data["contentId"])
}

func TestRecommendContent_Validation_InvalidUUID(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/content/INVALID-UUID/content", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "ID is not a valid UUID.", resp.Data["contentId"])
}

// -------------------------------------------------------------------------
// User-to-Content Tests
// -------------------------------------------------------------------------

func TestRecommendUser_Success(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// 1. Arrange
	// User liked Adventure.
	// User liked and did not like Action
	// User did not like Romance
	// All other movies have same rating
	// Recommendation should give Adventure Movie, Action Movie, Romance Movie in that order
	userID := "11111111-1111-1111-1111-111111111111"
	payload := fmt.Sprintf(`{
		"data": {
			"reviews": [
				{
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "%s",
					"title": "Action Adventure Movie",
					"genres": ["Action", "Adventure"],
					"score": 90,
					"review": "Love it"
				},
				{
					"contentId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "%s",
					"title": "Action Romance Movie",
					"genres": ["Action", "Romance"],
					"score": 10,
					"review": "Love it"
				},
				{
					"contentId": "bbbbbbbb-cccc-cccc-cccc-bbbbbbbbbbbb",
					"userId": "99999999-9999-9999-9999-999999999999",
					"title": "Action Movie",
					"genres": ["Action"],
					"score": 50,
					"review": "Also good"
				},
				{
					"contentId": "bbbbbbbb-cccc-bbbb-cccc-bbbbbbbbbbbb",
					"userId": "99999999-8888-9999-9999-999999999999",
					"title": "Adventure Movie",
					"genres": ["Adventure"],
					"score": 50,
					"review": "Also good"
				},
				{
					"contentId": "bbbbbbbb-cccc-cccc-bbbb-bbbbbbbbbbbb",
					"userId": "99999999-9999-8888-9999-999999999999",
					"title": "Romance Movie",
					"genres": ["Romance"],
					"score": 50,
					"review": "Also good"
				}
			]
		}
	}`, userID, userID)
	seedReviews(t, e, payload)

	targetURL := "/api/v1/recommendations/users/" + userID + "/content"
	req := httptest.NewRequest(http.MethodGet, targetURL, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp RecResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.NotEmpty(t, resp.Data)
	assert.Equal(t, "Adventure Movie", resp.Data.Recommendations[0].Title)
	assert.Equal(t, "Action Movie", resp.Data.Recommendations[1].Title)
	assert.Equal(t, "Romance Movie", resp.Data.Recommendations[2].Title)
}

func TestRecommendUser_NotFound(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Valid UUID, but this user has never submitted a review
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/recommendations/users/ffffffff-ffff-ffff-ffff-ffffffffffff/content",
		nil,
	)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp FailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "User with such ID not found.", resp.Data["userId"])
}

// -------------------------------------------------------------------------
// Pagination Tests
// -------------------------------------------------------------------------

func TestRecommendation_Pagination(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	// Seed 3 movies connected to Source
	sourceID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	payload := `{
		"data": {
			"reviews": [
				{ 
					"contentId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"userId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"title": "Source",
					"genres": ["Action"],
					"score": 10,
					"review": "."
				},
				{ 
					"contentId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"userId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"title": "Rec 1", 
					"genres": ["Action"],
					"score": 90,
					"review": "."
				},
				{
					"contentId": "cccccccc-cccc-cccc-cccc-cccccccccccc",
					"userId": "cccccccc-cccc-cccc-cccc-cccccccccccc",
					"title": "Rec 2", 
					"genres": ["Action"],
					"score": 80,
					"review": "."
				},
				{
					"contentId": "dddddddd-dddd-dddd-dddd-dddddddddddd",
					"userId": "dddddddd-dddd-dddd-dddd-dddddddddddd",
					"title": "Rec 3", 
					"genres": ["Action"],
					"score": 70,
					"review": "."
				}
			]
		}
	}`
	seedReviews(t, e, payload)

	// Request Limit=2 Offset=1
	url := "/api/v1/recommendations/content/" + sourceID + "/content?limit=2&offset=1"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp RecResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	// Should only return 2 items, even though 3 are available
	assert.Len(t, resp.Data.Recommendations, 2)
	assert.Equal(t, "Rec 2", resp.Data.Recommendations[0].Title)
	assert.Equal(t, "Rec 3", resp.Data.Recommendations[1].Title)
}

func TestRecommendation_Pagination_Validation(t *testing.T) {
	t.Parallel()
	e := SetupServer()

	reqLimit := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/recommendations/content/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/content?limit=-1&offset=-2",
		nil,
	)
	recLimit := httptest.NewRecorder()
	e.ServeHTTP(recLimit, reqLimit)

	assert.Equal(t, http.StatusBadRequest, recLimit.Code)

	var respLimit FailResponse
	_ = json.Unmarshal(recLimit.Body.Bytes(), &respLimit)
	assert.Equal(t, "fail", respLimit.Status)
	assert.Equal(t, "Offset must be non-negative integer.", respLimit.Data["offset"])
	assert.Equal(t, "Limit must be non-negative integer.", respLimit.Data["limit"])
}
