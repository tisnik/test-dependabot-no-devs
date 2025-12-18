package service

import (
	"fmt"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/google/uuid"
)

type ReviewService struct {
	repository *repository.ReviewRepository
}

func NewReviewService(repository *repository.ReviewRepository) *ReviewService {
	return &ReviewService{
		repository: repository,
	}
}

func (s *ReviewService) CreateReviews(reviews []domain.Review) ([]domain.Review, error) {
	savedReviews, err := s.repository.SaveBatch(reviews)
	if err != nil {
		return nil, fmt.Errorf("failed to save reviews: %w", err)
	}
	return savedReviews, nil
}

func (s *ReviewService) DeleteReview(id uuid.UUID) error {
	err := s.repository.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}
