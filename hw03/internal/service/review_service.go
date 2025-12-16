package service

import (
	"fmt"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/repository"
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) IngestReviews(reviews []models.RawReview) ([]models.Review, error) {
	results := make([]models.Review, 0, len(reviews))

	for _, rawReview := range reviews {
		createdReview, err := s.repo.CreateReview(rawReview)
		if err != nil {
			return nil, fmt.Errorf("error creating review: %w", err)
		}
		results = append(results, createdReview)
	}

	return results, nil
}

func (s *ReviewService) DeleteReview(id string) error {
	err := s.repo.DeleteReview(id)
	if err != nil {
		return fmt.Errorf("error in deletion of review: %w", err)
	}
	return nil
}
