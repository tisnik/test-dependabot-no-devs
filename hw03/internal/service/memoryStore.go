package service

import (
	"errors"
	"slices"
	"sync"

	"github.com/course-go/reelgoofy/internal/domain"
)

const initCapacity = 30

var (
	ErrInvalidJsonParams = errors.New("invalid json parameters")
	ErrNotFound          = errors.New("not found")
	ErrInternal          = errors.New("internal error")
)

type ReviewStore interface {
	AddReview(review domain.Review)
	GetReviewsByUser(userId string) []domain.Review
	GetReviewsByUserMap(userId string) map[string]domain.Review
	GetReviewsByContent(contentId string) []domain.Review
	DeleteReview(id string) error
}

type MemoryStorage struct {
	mu             sync.RWMutex
	reviews        map[string]domain.Review
	contentReviews map[string][]domain.Review
	userReviews    map[string][]domain.Review
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		reviews:        make(map[string]domain.Review, initCapacity),
		contentReviews: make(map[string][]domain.Review, initCapacity),
		userReviews:    make(map[string][]domain.Review, initCapacity),
	}
}

func (m *MemoryStorage) AddReview(review domain.Review) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, filmExists := m.contentReviews[review.ContentID]
	m.reviews[review.ID] = review
	if !filmExists {
		m.contentReviews[review.ContentID] = make([]domain.Review, 0, initCapacity)
	}
	m.contentReviews[review.ContentID] = insertSorted(m.contentReviews[review.ContentID], review)
	_, userHasReview := m.userReviews[review.UserID]
	if !userHasReview {
		m.userReviews[review.UserID] = make([]domain.Review, 0, initCapacity)
	}
	m.userReviews[review.UserID] = insertSorted(m.userReviews[review.UserID], review)
}

func (m *MemoryStorage) DeleteReview(reviewId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	review, ok := m.reviews[reviewId]
	if !ok {
		return ErrNotFound
	}
	delete(m.reviews, reviewId)
	m.contentReviews[review.ContentID] = deleteFromArray(m.contentReviews[review.ContentID], reviewId)
	m.userReviews[review.UserID] = deleteFromArray(m.userReviews[review.UserID], reviewId)
	return nil
}

func (m *MemoryStorage) GetReviewsByUser(userId string) []domain.Review {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copyReviews := make([]domain.Review, len(m.userReviews[userId]))
	copy(copyReviews, m.userReviews[userId])
	return copyReviews
}

func (m *MemoryStorage) GetReviewsByUserMap(userId string) map[string]domain.Review {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copyReviews := make(map[string]domain.Review, len(m.userReviews[userId]))
	for _, review := range m.userReviews[userId] {
		copyReviews[review.ContentID] = review
	}
	return copyReviews
}

func (m *MemoryStorage) GetReviewsByContent(contentId string) []domain.Review {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copyReviews := make([]domain.Review, len(m.contentReviews[contentId]))
	copy(copyReviews, m.contentReviews[contentId])
	return copyReviews
}

func insertSorted(reviews []domain.Review, review domain.Review) []domain.Review {
	for index, current := range reviews {
		if review.Score > current.Score {
			return slices.Insert(reviews, index, review)
		}
	}
	return slices.Insert(reviews, len(reviews), review)
}

func deleteFromArray(reviews []domain.Review, reviewId string) []domain.Review {
	for index, current := range reviews {
		if current.ID == reviewId {
			return append(reviews[:index], reviews[index+1:]...)
		}
	}
	return reviews
}

func (m *MemoryStorage) GetReview(id string) (domain.Review, bool) {
	review, ok := m.reviews[id]
	return review, ok
}
