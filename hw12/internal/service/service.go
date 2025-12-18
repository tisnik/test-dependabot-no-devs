package service

import (
	"context"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/google/uuid"
)

type Reviewer interface {
	Create(ctx context.Context, review model.Review) error
}

type Recommender interface {
	GetRecommendationsForUser(ctx context.Context, userID uuid.UUID) ([]model.Recommendation, error)
}
