package tests_test

import (
	"net/http"
	"testing"

	"github.com/course-go/reelgoofy/internal/api/router"
	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func generateSetup(t *testing.T) (http.Handler, *service.ReviewService, *service.MemoryStorage) {
	t.Helper()
	reviewStore := service.NewMemoryStorage()
	reviewService := service.NewReviewService(reviewStore)
	router := router.NewRouter(reviewService)
	return router, reviewService, reviewStore
}

func compareRawReturned(t *testing.T, raw domain.RawReview, returned domain.Review) {
	t.Helper()
	assert.Equal(t, raw.ContentID, returned.ContentID)
	assert.Equal(t, raw.UserID, returned.UserID)
	assert.Equal(t, raw.Title, returned.Title)
	assert.Equal(t, raw.Genres, returned.Genres)
	assert.Equal(t, raw.Tags, returned.Tags)
	assert.Equal(t, raw.Description, returned.Description)
	assert.Equal(t, raw.Director, returned.Director)
	assert.Equal(t, raw.Actors, returned.Actors)
	assert.Equal(t, raw.Origins, returned.Origins)
	assert.Equal(t, raw.Duration, returned.Duration)
	assert.Equal(t, raw.Released, returned.Released)
	assert.Equal(t, raw.Review, returned.Review)
	assert.Equal(t, raw.Score, returned.Score)

	_, err := uuid.Parse(returned.ID)
	assert.NoError(t, err)
}

func compareReviewRecommendation(t *testing.T, review domain.Review, recommendation domain.Recommendation) {
	t.Helper()
	assert.Equal(t, review.Title, recommendation.Title)
}
