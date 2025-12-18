package internal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal"
	"github.com/course-go/reelgoofy/internal/recommendation"
	"github.com/course-go/reelgoofy/internal/rest"
	"github.com/course-go/reelgoofy/internal/review"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngestReviews(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		payload        string
		expectedCode   int
		expectedStatus rest.ResponseStatus
	}{
		{
			name: "Valid Review Ingestion",
			payload: `{
				"data": {
					"reviews": [
						{
							"contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
							"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
							"review": "Great movie!",
							"score": 85
						}
					]
				}
			}`,
			expectedCode:   http.StatusCreated,
			expectedStatus: "success",
		},
		{
			name: "Invalid UUID Format (400)",
			payload: `{
				"data": {
					"reviews": [
						{
							"contentId": "NOT-A-UUID",
							"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
							"review": "Bad ID",
							"score": 50
						}
					]
				}
			}`,
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
		{
			name: "Missing Required Fields (400)",
			payload: `{
				"data": {
					"reviews": [
						{
							"review": "Missing IDs"
						}
					]
				}
			}`,
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := internal.SetupRouter()
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/api/v1/reviews",
				bytes.NewBufferString(tt.payload),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			var resp rest.Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

func TestDeleteReview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		reviewID       string
		expectedCode   int
		expectedStatus rest.ResponseStatus
	}{
		{
			name:           "Successful Deletion",
			reviewID:       "",
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name:           "Review Not Found",
			reviewID:       "99999999-9999-9999-9999-999999999999",
			expectedCode:   http.StatusNotFound,
			expectedStatus: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := internal.SetupRouter()

			id := sendReviewForDeletion(router)

			if tt.reviewID == "" {
				tt.reviewID = id
			}

			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodDelete,
				"/api/v1/reviews/"+tt.reviewID,
				nil,
			)

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			var resp rest.Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

func TestRecommendContentToContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		contentID      string
		queryString    string
		expectedCode   int
		expectedStatus rest.ResponseStatus
	}{
		{
			name:           "Default Params",
			contentID:      "937b33bf-066a-44f7-9a9b-d65071d27270",
			queryString:    "",
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name:           "Custom Pagination",
			contentID:      "937b33bf-066a-44f7-9a9b-d65071d27270",
			queryString:    "?limit=5&offset=10",
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name:           "Invalid Limit",
			contentID:      "937b33bf-066a-44f7-9a9b-d65071d27270",
			queryString:    "?limit=five",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
		{
			name:           "Invalid Offset",
			contentID:      "937b33bf-066a-44f7-9a9b-d65071d27270",
			queryString:    "?offset=ten",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
		{
			name:           "Invalid UUID",
			contentID:      "invalid-uuid-string",
			queryString:    "",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := internal.SetupRouter()

			url := "/api/v1/recommendations/content/" + tt.contentID + "/content" + tt.queryString

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			var resp rest.Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

func TestRecommendContentToUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		userID         string
		queryString    string
		expectedCode   int
		expectedStatus rest.ResponseStatus
	}{
		{
			name:           "Valid Request",
			userID:         "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
			queryString:    "?limit=10",
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name:           "Invalid Limit",
			userID:         "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
			queryString:    "?limit=abc",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
		{
			name:           "Invalid UUID",
			userID:         "invalid-uuid-string",
			queryString:    "",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
		{
			name:           "Invalid Offset",
			userID:         "937b33bf-066a-44f7-9a9b-d65071d27270",
			queryString:    "?offset=ten",
			expectedCode:   http.StatusBadRequest,
			expectedStatus: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := internal.SetupRouter()

			url := "/api/v1/recommendations/user/" + tt.userID + "/content" + tt.queryString

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			var resp rest.Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

type RecDataWrapper struct {
	Recommendations []recommendation.Recommendation `json:"recommendations"`
}

type RecResponse struct {
	Status string         `json:"status"`
	Data   RecDataWrapper `json:"data"`
}

func TestRecommendationsWithRealData(t *testing.T) {
	t.Parallel()

	t.Run("Recommend Content-to-Content (Matrix -> Inception)", func(t *testing.T) {
		t.Parallel()
		router := internal.SetupRouter()

		seedTestReviews(t, router)

		targetContentID := "11111111-0000-0000-0000-000000000001"

		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/api/v1/recommendations/content/"+targetContentID+"/content",
			nil,
		)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp RecResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "success", resp.Status)

		assert.NotEmpty(t, resp.Data.Recommendations, "Should return recommendations")

		foundInception := foundFilmInResponse(t, resp, "Inception")
		assert.True(t, foundInception, "Should have recommended Inception based on shared user reviews")
	})

	t.Run("Recommend User-to-Content", func(t *testing.T) {
		t.Parallel()
		router := internal.SetupRouter()

		seedTestReviews(t, router)

		userID := "33333333-0000-0000-0000-000000000001"
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/api/v1/recommendations/user/"+userID+"/content",
			nil,
		)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp RecResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		log.Print("Recommendations Response: ")
		log.Println(resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)

		assert.NotEmpty(t, resp.Data.Recommendations, "Should return recommendations")

		foundTitanic := foundFilmInResponse(t, resp, "Titanic")

		assert.True(t, foundTitanic, "Should have recommended Titanic based on shared user reviews")
	})
}

type reviewResponse struct {
	Status  string                     `json:"status"`
	Data    map[string][]review.Review `json:"data,omitempty"`
	Message string                     `json:"message,omitempty"`
}

func sendReviewForDeletion(router *chi.Mux) string {
	reviewPayload := `{
				"data": {
					"reviews": [
						{
							"contentId": "937b33bf-066a-44f7-9a9b-d65071d272ab",
							"userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9b",
							"review": "Great movie!",
							"score": 85
						}
					]
				}
			}`

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/reviews",
		bytes.NewBufferString(reviewPayload),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	var jResponse reviewResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &jResponse)
	log.Println(jResponse)
	return jResponse.Data["reviews"][0].Id
}

var seedPayload = map[string]any{
	"data": map[string]any{
		"reviews": []map[string]any{
			{
				"contentId":   "11111111-0000-0000-0000-000000000001",
				"userId":      "33333333-0000-0000-0000-000000000001",
				"review":      "Mind blowing concept.",
				"score":       95,
				"title":       "The Matrix",
				"director":    "Lana Wachowski",
				"genres":      []string{"Sci-Fi", "Action"},
				"actors":      []string{"Keanu Reeves", "Laurence Fishburne"},
				"origins":     []string{"USA", "Australia"},
				"duration":    8160,
				"released":    "1999-03-31",
				"tags":        []string{"cyberpunk", "simulation", "dystopia"},
				"description": "A computer hacker learns about the true nature of his reality.",
			},
			{
				"contentId":   "11111111-0000-0000-0000-000000000002",
				"userId":      "33333333-0000-0000-0000-000000000001",
				"review":      "Dreams within dreams.",
				"score":       90,
				"title":       "Inception",
				"director":    "Christopher Nolan",
				"genres":      []string{"Sci-Fi", "Action", "Thriller"},
				"actors":      []string{"Leonardo DiCaprio", "Joseph Gordon-Levitt"},
				"origins":     []string{"USA", "UK"},
				"duration":    8880,
				"released":    "2010-07-16",
				"tags":        []string{"dream", "thief", "subconscious"},
				"description": "A thief who steals corporate secrets through the use of dream-sharing technology.",
			},

			{
				"contentId":   "22222222-0000-0000-0000-000000000001",
				"userId":      "33333333-0000-0000-0000-000000000002",
				"review":      "A tragic love story.",
				"score":       88,
				"title":       "Titanic",
				"director":    "James Cameron",
				"genres":      []string{"Romance", "Drama"},
				"actors":      []string{"Leonardo DiCaprio", "Kate Winslet"},
				"origins":     []string{"USA"},
				"duration":    11700,
				"released":    "1997-12-19",
				"tags":        []string{"shipwreck", "iceberg", "love"},
				"description": "A seventeen-year-old aristocrat falls in love with a kind but poor artist.",
			},
			{
				"contentId":   "22222222-0000-0000-0000-000000000002",
				"userId":      "33333333-0000-0000-0000-000000000002",
				"review":      "So romantic and sad.",
				"score":       92,
				"title":       "The Notebook",
				"director":    "Nick Cassavetes",
				"genres":      []string{"Romance", "Drama"},
				"actors":      []string{"Ryan Gosling", "Rachel McAdams"},
				"origins":     []string{"USA"},
				"duration":    7380,
				"released":    "2004-06-25",
				"tags":        []string{"memory", "elderly", "flashback"},
				"description": "A poor yet passionate young man falls in love with a rich young woman.",
			},

			{
				"contentId":   "11111111-0000-0000-0000-000000000001",
				"userId":      "33333333-0000-0000-0000-000000000003",
				"review":      "Classic sci-fi.",
				"score":       100,
				"title":       "The Matrix",
				"director":    "Lana Wachowski",
				"genres":      []string{"Sci-Fi", "Action"},
				"actors":      []string{"Keanu Reeves", "Laurence Fishburne"},
				"origins":     []string{"USA", "Australia"},
				"duration":    8160,
				"released":    "1999-03-31",
				"tags":        []string{"cyberpunk", "simulation", "dystopia"},
				"description": "A computer hacker learns about the true nature of his reality.",
			},
			{
				"contentId":   "11111111-0000-0000-0000-000000000002",
				"userId":      "33333333-0000-0000-0000-000000000003",
				"review":      "Nolan is genius.",
				"score":       85,
				"title":       "Inception",
				"director":    "Christopher Nolan",
				"genres":      []string{"Sci-Fi", "Action", "Thriller"},
				"actors":      []string{"Leonardo DiCaprio", "Joseph Gordon-Levitt"},
				"origins":     []string{"USA", "UK"},
				"duration":    8880,
				"released":    "2010-07-16",
				"tags":        []string{"dream", "thief", "subconscious"},
				"description": "A thief who steals corporate secrets through the use of dream-sharing technology.",
			},
		},
	},
}

func seedTestReviews(t *testing.T, router http.Handler) {
	t.Helper()

	payloadBytes, err := json.Marshal(seedPayload)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/reviews",
		bytes.NewBuffer(payloadBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Seeding failed")
}

func foundFilmInResponse(t *testing.T, resp RecResponse, filmTitle string) bool {
	t.Helper()

	foundFilm := false
	for _, rec := range resp.Data.Recommendations {
		if rec.Title == filmTitle {
			foundFilm = true
			break
		}
	}
	return foundFilm
}
