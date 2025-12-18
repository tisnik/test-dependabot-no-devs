package service

import (
	"context"
	"fmt"
	"time"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/google/uuid"
)

// ReviewService provides operations for reviews.
type ReviewService struct {
	storage storage.Storage
}

func NewReviewService(storage storage.Storage) *ReviewService {
	return &ReviewService{
		storage: storage,
	}
}

func (s *ReviewService) Create(ctx context.Context, review model.Review) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	if review.CreatedAt.IsZero() {
		review.CreatedAt = time.Now()
	}

	err := s.storage.CreateReview(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}
