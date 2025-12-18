package service

import (
	"context"
	"fmt"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/google/uuid"
)

const highRatingThreshold = 4

// RecommendationService provides movie recommendation logic.
type RecommendationService struct {
	storage storage.Storage
}

func NewRecommendationService(storage storage.Storage) *RecommendationService {
	return &RecommendationService{
		storage: storage,
	}
}

func (s *RecommendationService) GetRecommendationsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.Recommendation, error) {
	reviews, err := s.storage.GetReviewsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews for user %s: %w", userID, err)
	}

	var recommendations []model.Recommendation
	for _, review := range reviews {
		if review.Rating >= highRatingThreshold {
			recommendations = append(recommendations, model.Recommendation{
				MovieID: review.MovieID,
				Score:   float64(review.Rating),
			})
		}
	}

	return recommendations, nil
}
