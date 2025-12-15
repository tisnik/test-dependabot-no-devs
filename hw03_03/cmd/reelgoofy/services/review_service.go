package services

import (
	"errors"
	"sync"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/structs"
	"github.com/google/uuid"
)

type ReviewService struct {
	reviews map[string]structs.Review
	mu      sync.RWMutex
}

func NewReviewService() *ReviewService {
	return &ReviewService{
		reviews: make(map[string]structs.Review),
	}
}

func (s *ReviewService) IngestReviews(rawReviews []structs.RawReview) ([]structs.Review, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reviews := make([]structs.Review, 0, len(rawReviews))

	for _, raw := range rawReviews {
		if !isValidUUID(raw.ContentID) || !isValidUUID(raw.UserID) {
			return nil, errors.New("invalid UUID format")
		}

		if raw.Score < 0 || raw.Score > 100 {
			return nil, errors.New("score must be between 0 and 100")
		}

		review := structs.Review{
			ID:          uuid.New().String(),
			ContentID:   raw.ContentID,
			UserID:      raw.UserID,
			Title:       raw.Title,
			Genres:      raw.Genres,
			Tags:        raw.Tags,
			Description: raw.Description,
			Director:    raw.Director,
			Actors:      raw.Actors,
			Origins:     raw.Origins,
			Duration:    raw.Duration,
			Released:    raw.Released,
			Review:      raw.Review,
			Score:       raw.Score,
		}

		s.reviews[review.ID] = review
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (s *ReviewService) DeleteReview(reviewID string) error {
	if !isValidUUID(reviewID) {
		return errors.New("invalid UUID format")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.reviews[reviewID]
	if !exists {
		return errors.New("review not found")
	}

	delete(s.reviews, reviewID)
	return nil
}

func (s *ReviewService) GetReviewsByContent(contentID string) []structs.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reviews []structs.Review
	for _, review := range s.reviews {
		if review.ContentID == contentID {
			reviews = append(reviews, review)
		}
	}

	return reviews
}

func (s *ReviewService) GetReviewsByUser(userID string) []structs.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reviews []structs.Review
	for _, review := range s.reviews {
		if review.UserID == userID {
			reviews = append(reviews, review)
		}
	}

	return reviews
}

func (s *ReviewService) GetAllReviews() []structs.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviews := make([]structs.Review, 0, len(s.reviews))
	for _, review := range s.reviews {
		reviews = append(reviews, review)
	}

	return reviews
}

func (s *ReviewService) ContentExists(contentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, review := range s.reviews {
		if review.ContentID == contentID {
			return true
		}
	}

	return false
}

func (s *ReviewService) UserExists(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, review := range s.reviews {
		if review.UserID == userID {
			return true
		}
	}

	return false
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
