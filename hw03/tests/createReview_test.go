package tests_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddReview(t *testing.T) {
	t.Parallel()
	t.Run("bad IDs", func(t *testing.T) {
		t.Parallel()
		review := domain.RawReview{
			ContentID:   "3",
			UserID:      "4",
			Title:       "One Flew over the Cuckoo's Nest",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "Christopher Nolan",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       75,
		}
		router, _, _ := generateSetup(t)
		payload := map[string]any{
			"data": map[string]any{
				"reviews": []domain.RawReview{review},
			},
		}
		b, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(b))
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
			"contentId": "ID is not a valid UUID.",
			"userId":    "ID is not a valid UUID.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("bad date", func(t *testing.T) {
		t.Parallel()
		review := domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27270",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
			Title:       "One Flew over the Cuckoo's Nest",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "Christopher Nolan",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "very bad date",
			Review:      "I really enjoyed this one.",
			Score:       75,
		}
		router, _, _ := generateSetup(t)

		payload := map[string]any{
			"data": map[string]any{
				"reviews": []domain.RawReview{review},
			},
		}
		b, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(b))
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
			"released": "Invalid date formats.",
		}
		assert.Equal(t, errorsMap, data)
	})

	t.Run("Single review", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		payload := map[string]any{
			"data": map[string]any{
				"reviews": []domain.RawReview{RawReviewA},
			},
		}
		b, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp domain.ReviewsResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		returnedReviews := resp.Data.Reviews
		assert.Len(t, returnedReviews, 1)
		returnedReview := returnedReviews[0]

		compareRawReturned(t, RawReviewA, returnedReview)

		dbReview, ok := store.GetReview(returnedReview.ID)
		assert.True(t, ok)
		assert.Equal(t, "One Flew over the Cuckoo's Nest", dbReview.Title)
	})

	t.Run("Multiple reviews", func(t *testing.T) {
		t.Parallel()
		router, _, store := generateSetup(t)
		payload := map[string]any{
			"data": map[string]any{
				"reviews": []domain.RawReview{RawReviewA, RawReviewB, RawReviewC},
			},
		}
		b, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/reviews", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp domain.ReviewsResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		returnedReviews := resp.Data.Reviews
		assert.Equal(t, domain.StatusSuccess, resp.Status)
		assert.Len(t, returnedReviews, 3)
		returnedReviewA := returnedReviews[0]
		returnedReviewB := returnedReviews[1]
		returnedReviewC := returnedReviews[2]

		compareRawReturned(t, RawReviewA, returnedReviewA)
		compareRawReturned(t, RawReviewB, returnedReviewB)
		compareRawReturned(t, RawReviewC, returnedReviewC)

		userReviews := store.GetReviewsByUser("2f99df7d-751c-40c9-aeea-8be8cd7bfa98")
		assert.Len(t, userReviews, 2)
	})
}
