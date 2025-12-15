package repository

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/request"
	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/response"
	"github.com/google/uuid"
)

type Repository struct {
	mu   sync.RWMutex
	data map[string]request.Review
}

func NewRepository() *Repository {
	return &Repository{
		mu:   sync.RWMutex{},
		data: make(map[string]request.Review),
	}
}

func (r *Repository) SaveReview(reviews request.RawReviewsRequest) response.Reviews {
	newReviews := response.Reviews{
		Data: make([]response.Review, len(reviews.Data.Reviews)),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, review := range reviews.Data.Reviews {
		id := uuid.NewString()
		r.data[id] = review
		newReviews.Data[i] = response.Review{
			ID:          id,
			ContentID:   review.ContentID,
			UserID:      review.UserID,
			Title:       review.Title,
			Genres:      review.Genres,
			Tags:        review.Tags,
			Description: review.Description,
			Director:    review.Director,
			Actors:      review.Actors,
			Origins:     review.Origins,
			Duration:    review.Duration,
			Released:    review.Released,
			Review:      review.Review,
			Score:       review.Score,
		}
	}
	slog.Info("Repo size", "size", len(r.data))
	return newReviews
}

func (r *Repository) DeleteReview(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.data[id]
	if !ok {
		return errors.New("review not found")
	}

	delete(r.data, id)
	slog.Info("Repo size", "size", len(r.data))
	return nil
}

func (r *Repository) GetAllReviews() []request.Review {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reviews := make([]request.Review, 0, len(r.data))
	for _, review := range r.data {
		reviews = append(reviews, review)
	}

	return reviews
}

func (r *Repository) GetReviewsByUserID(userID string) []request.Review {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userReviews []request.Review
	for _, review := range r.data {
		if review.UserID == userID {
			userReviews = append(userReviews, review)
		}
	}

	return userReviews
}
