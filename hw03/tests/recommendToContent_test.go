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

func TestRecommendContentToContent(t *testing.T) {
	t.Parallel()
	t.Run("Bad Id", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(http.MethodGet, "/recommendations/content/veryBadId123/content", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp domain.FailResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFail, resp.Status)

		data := resp.Data
		assert.Len(t, data, 1)
		errorsMap := map[string]string{
			"contentId": "ID is not a valid UUID.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content",
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
			"contentId": "Content with such ID not found.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Invalid limit and offset", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content?limit=-1&offset=invalid",
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
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content",
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

	t.Run("one available review, one recommendation", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content",
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
		compareReviewRecommendation(t, ReviewB, returnedRecommendations[0])
	})

	t.Run("3 available reviews, 2 recommendations", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)
		store.AddReview(ReviewD)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content",
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
		assert.Len(t, returnedRecommendations, 2)

		// higher score first
		compareReviewRecommendation(t, ReviewB, returnedRecommendations[1])
		compareReviewRecommendation(t, ReviewC, returnedRecommendations[0])
		// the other user did not make a review for the ReviewA, so his review D should not be included
	})

	t.Run("2 available reviews, for the same film, 1 recommendation expected", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		store.AddReview(ReviewA)
		store.AddReview(ReviewB)
		store.AddReview(ReviewE)
		store.AddReview(ReviewF)

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content",
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

		compareReviewRecommendation(t, ReviewB, returnedRecommendations[0])
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

		req := httptest.NewRequest(
			http.MethodGet,
			"/recommendations/content/937b33bf-066a-44f7-9a9b-d65071d27270/content?limit=2&offset=1",
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

		// order is D C F B
		compareReviewRecommendation(t, ReviewC, returnedRecommendations[0])
		compareReviewRecommendation(t, ReviewF, returnedRecommendations[1])
	})
}
