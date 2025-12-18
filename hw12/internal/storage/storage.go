package storage

import (
	"context"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/google/uuid"
)

// Storage is a composite interface for all storage operations.
type Storage interface {
	CreateReview(ctx context.Context, review model.Review) error
	GetReviewsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Review, error)
}
