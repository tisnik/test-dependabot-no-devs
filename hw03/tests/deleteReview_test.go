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

func TestDeleteReview(t *testing.T) {
	t.Parallel()
	t.Run("Bad Id", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(http.MethodDelete, "/reviews/veryBadId123", nil)
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
			"reviewId": "ID is not a valid UUID.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()
		router, _, _ := generateSetup(t)
		req := httptest.NewRequest(http.MethodDelete, "/reviews/937b33bf-066a-44f7-9a9b-d65071d27270", nil)
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
			"reviewId": "Review with such ID not found.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Delete one", func(t *testing.T) {
		t.Parallel()
		router, service, store := generateSetup(t)

		_, _ = service.SaveReviews([]domain.RawReview{RawReviewA, RawReviewD})
		store.AddReview(ReviewB)
		store.AddReview(ReviewC)
		assert.Len(t, store.GetReviewsByUser("2f99df7d-751c-40c9-aeea-8be8cd7bfa98"), 4)

		req := httptest.NewRequest(http.MethodDelete, "/reviews/237b33bf-066a-44f7-9a9b-d65071d27270", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.SuccessResponse[any]
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		data := resp.Data
		assert.Nil(t, data)

		assert.Len(t, store.GetReviewsByUser("2f99df7d-751c-40c9-aeea-8be8cd7bfa98"), 3)
		survivor, ok := store.GetReview("237b33bf-066a-44f7-9a9b-d65071d27278")
		assert.True(t, ok)
		assert.Equal(t, "Trainspotting", survivor.Title)
	})
}
