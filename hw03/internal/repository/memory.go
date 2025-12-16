package repository

import (
	"errors"
	"sync"

	"github.com/Markek1/reelgoofy/internal/domain"
)

var ErrReviewNotFound = errors.New("review not found")

type MemoryRepository struct {
	mu      sync.RWMutex
	reviews map[domain.ReviewID]domain.Review
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		reviews: make(map[domain.ReviewID]domain.Review),
	}
}

func (r *MemoryRepository) Save(review domain.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reviews[review.ID] = review
	return nil
}

func (r *MemoryRepository) Delete(id domain.ReviewID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.reviews[id]; !exists {
		return ErrReviewNotFound
	}

	delete(r.reviews, id)
	return nil
}

func (r *MemoryRepository) Get(id domain.ReviewID) (domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	review, exists := r.reviews[id]
	if !exists {
		return domain.Review{}, ErrReviewNotFound
	}
	return review, nil
}

func (r *MemoryRepository) GetAll() ([]domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reviews := make([]domain.Review, 0, len(r.reviews))
	for _, review := range r.reviews {
		reviews = append(reviews, review)
	}
	return reviews, nil
}
