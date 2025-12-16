package service

import (
	"fmt"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/google/uuid"
)

type ReviewService struct {
	store ReviewStore
}

func NewReviewService(store ReviewStore) *ReviewService {
	return &ReviewService{store: store}
}

func (s *ReviewService) SaveReviews(rawReviews []domain.RawReview) ([]domain.Review, error) {
	reviews := make([]domain.Review, 0, len(rawReviews))
	for _, r := range rawReviews {
		review := domain.Review{
			ID: uuid.New().String(),
			RawReview: domain.RawReview{
				ContentID:   r.ContentID,
				UserID:      r.UserID,
				Title:       r.Title,
				Genres:      r.Genres,
				Tags:        r.Tags,
				Description: r.Description,
				Director:    r.Director,
				Actors:      r.Actors,
				Origins:     r.Origins,
				Duration:    r.Duration,
				Released:    r.Released,
				Review:      r.Review,
				Score:       r.Score,
			},
		}
		s.store.AddReview(review)
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (s *ReviewService) DeleteReview(reviewId string) error {
	err := s.store.DeleteReview(reviewId)
	if err != nil {
		return fmt.Errorf("failed to delete review %s: %w", reviewId, err)
	}
	return nil
}
