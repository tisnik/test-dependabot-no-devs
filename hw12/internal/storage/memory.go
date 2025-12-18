package storage

import (
	"context"
	"sync"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/google/uuid"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
type MemoryStorage struct {
	mu      sync.RWMutex
	reviews map[uuid.UUID][]model.Review
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		reviews: make(map[uuid.UUID][]model.Review),
	}
}

func (s *MemoryStorage) CreateReview(ctx context.Context, review model.Review) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews[review.UserID] = append(s.reviews[review.UserID], review)
	return nil
}

func (s *MemoryStorage) GetReviewsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Review, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviews, ok := s.reviews[userID]
	if !ok {
		return []model.Review{}, nil
	}
	return reviews, nil
}
