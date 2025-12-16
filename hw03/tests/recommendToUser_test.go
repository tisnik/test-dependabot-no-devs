package tests_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendContentToUser(t *testing.T) {
	t.Parallel()

	t.Run("Bad Id", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(http.MethodGet, "/recommendations/users/veryBadId123/content", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp domain.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFail, resp.Status)

		data := resp.Data
		errorsMap := map[string]string{
			"userId": "ID is not a valid UUID.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp domain.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFail, resp.Status)

		data := resp.Data
		assert.Len(t, data, 1)
		errorsMap := map[string]string{
			"userId": "User with such ID not found.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Invalid limit and offset", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content?limit=-1&offset=invalid",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp domain.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFail, resp.Status)

		data := resp.Data
		assert.Len(t, data, 2)
		errorsMap := map[string]string{
			"limit":  "Limit must be non-negative integer.",
			"offset": "Offset must be non-negative integer.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Single review, nothing to recommend", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)

		store.AddReview(ReviewA)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Empty(t, returnedRecommendations)
	})

	t.Run("one user available, no recommendations", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Empty(t, returnedRecommendations)
	})

	t.Run("2 users, no films in common, no recommendations", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)
		store.AddReview(ReviewD)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Empty(t, returnedRecommendations)
	})

	t.Run("2 users, 1 recommendation", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)
		store.AddReview(ReviewD)
		store.AddReview(ReviewE)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Len(t, returnedRecommendations, 1)

		compareReviewRecommendation(t, ReviewD, returnedRecommendations[0])
	})

	t.Run("2 available reviews, for the same film, 1 recommendation expected", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewE)
		store.AddReview(ReviewD)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Len(t, returnedRecommendations, 1)

		compareReviewRecommendation(t, ReviewD, returnedRecommendations[0])
	})

	t.Run("Limit and offset test", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)

		store.AddReview(ReviewD)
		store.AddReview(ReviewE)
		store.AddReview(ReviewF)

		store.AddReview(ReviewG)
		store.AddReview(ReviewH)
		store.AddReview(ReviewI)
		store.AddReview(ReviewJ)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/users/2f99df7d-751c-40c9-aeea-8be8cd7bfa98/content?limit=2&offset=1",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp domain.SuccessResponse[domain.Recommendations]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedRecommendations := resp.Data.Recommendations
		assert.Len(t, returnedRecommendations, 2)

		// order is I H D J
		compareReviewRecommendation(t, ReviewH, returnedRecommendations[0])
		compareReviewRecommendation(t, ReviewD, returnedRecommendations[1])
	})
}
