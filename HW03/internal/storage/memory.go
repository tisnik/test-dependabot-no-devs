package storage

import (
	"errors"
	"sync"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/google/uuid"
)

// ReviewStore provides thread-safe in-memory storage for reviews with efficient
// lookups by review ID, content ID, or user ID. Uses read-write mutex for
// concurrent access and maintains indexes for fast content and user queries.
type ReviewStore struct {
	mu           sync.RWMutex
	reviews      map[uuid.UUID]models.Review
	contentIndex map[uuid.UUID][]uuid.UUID
	userIndex    map[uuid.UUID][]uuid.UUID
}

func NewReviewStore() *ReviewStore {
	return &ReviewStore{
		reviews:      make(map[uuid.UUID]models.Review),
		contentIndex: make(map[uuid.UUID][]uuid.UUID),
		userIndex:    make(map[uuid.UUID][]uuid.UUID),
	}
}

// Add adds a new review to the store.
func (s *ReviewStore) Add(review models.Review) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews[review.ID] = review

	s.contentIndex[review.ContentID] = append(s.contentIndex[review.ContentID], review.ID)
	s.userIndex[review.UserID] = append(s.userIndex[review.UserID], review.ID)
}

// Delete removes a review from the store.
func (s *ReviewStore) Delete(reviewID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[reviewID]
	if !exists {
		return errors.New("review with such ID not found")
	}

	delete(s.reviews, reviewID)

	s.removeFromIndex(s.contentIndex, review.ContentID, reviewID)

	s.removeFromIndex(s.userIndex, review.UserID, reviewID)

	return nil
}

// Get retrieves a review by ID.
func (s *ReviewStore) Get(reviewID uuid.UUID) (models.Review, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	review, exists := s.reviews[reviewID]
	return review, exists
}

// GetAll returns all reviews.
func (s *ReviewStore) GetAll() []models.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviews := make([]models.Review, 0, len(s.reviews))
	for _, review := range s.reviews {
		reviews = append(reviews, review)
	}
	return reviews
}

// GetByContentID returns all reviews for a specific content.
func (s *ReviewStore) GetByContentID(contentID uuid.UUID) []models.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviewIDs, exists := s.contentIndex[contentID]
	if !exists {
		return []models.Review{}
	}

	reviews := make([]models.Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		if review, ok := s.reviews[id]; ok {
			reviews = append(reviews, review)
		}
	}
	return reviews
}

// GetByUserID returns all reviews by a specific user.
func (s *ReviewStore) GetByUserID(userID uuid.UUID) []models.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviewIDs, exists := s.userIndex[userID]
	if !exists {
		return []models.Review{}
	}

	reviews := make([]models.Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		if review, ok := s.reviews[id]; ok {
			reviews = append(reviews, review)
		}
	}
	return reviews
}

// ContentExists checks if a content has any reviews.
func (s *ReviewStore) ContentExists(contentID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.contentIndex[contentID]
	return exists
}

// UserExists checks if user has any reviews.
func (s *ReviewStore) UserExists(userID uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.userIndex[userID]
	return exists
}

// removeFromIndex removes a reviewID from an index.
func (s *ReviewStore) removeFromIndex(index map[uuid.UUID][]uuid.UUID, key, reviewID uuid.UUID) {
	ids := index[key]
	for i, id := range ids {
		if id == reviewID {
			index[key] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(index[key]) == 0 {
		delete(index, key)
	}
}
