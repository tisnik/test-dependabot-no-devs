package repository

import (
	"github.com/course-go/reelgoofy/internal/db"
	"github.com/course-go/reelgoofy/internal/models"
	"github.com/google/uuid"
)

type ReviewRepository struct {
	db *db.MemoryDB
}

func NewReviewRepository(db *db.MemoryDB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) CreateReview(review models.RawReview) (models.Review, error) {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	for reviewId, existingReview := range r.db.Reviews {
		if existingReview.UserID == review.UserID && existingReview.ContentID == review.ContentID {
			r.db.Reviews[reviewId] = review
			return models.Review{
				ID:        reviewId,
				RawReview: review,
			}, nil
		}
	}

	newId := uuid.NewString()
	r.db.Reviews[newId] = review
	return models.Review{
		ID:        newId,
		RawReview: review,
	}, nil
}

func (r *ReviewRepository) DeleteReview(id string) error {
	r.db.Mu.Lock()
	defer r.db.Mu.Unlock()

	if _, exists := r.db.Reviews[id]; exists {
		delete(r.db.Reviews, id)
		return nil
	}

	return models.ErrNotFound
}

func (r *ReviewRepository) GetAllReviews() []models.RawReview {
	r.db.Mu.RLock()
	defer r.db.Mu.RUnlock()

	allReviews := make([]models.RawReview, 0, len(r.db.Reviews))
	for _, rawReview := range r.db.Reviews {
		allReviews = append(allReviews, rawReview)
	}
	return allReviews
}
