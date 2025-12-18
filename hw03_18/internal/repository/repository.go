package repository

import (
	"sync"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/google/uuid"
)

// Repository defines the persistence contract for reviews.
// For this assignment it is implemented purely in-memory.
type Repository interface {
	AddReviews(raw []model.RawReview) ([]model.Review, map[string]string)
	DeleteReview(id string) (bool, map[string]string)
}

// InMemoryRepository stores reviews and simple lookup indexes.
type InMemoryRepository struct {
	mu        sync.RWMutex
	reviews   map[string]model.Review
	byContent map[string][]model.Review
	byUser    map[string][]model.Review
}

// NewInMemoryRepository constructs a fresh empty repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		reviews:   make(map[string]model.Review),
		byContent: make(map[string][]model.Review),
		byUser:    make(map[string][]model.Review),
	}
}

func (repo *InMemoryRepository) AddReviews(raw []model.RawReview) ([]model.Review, map[string]string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	out := make([]model.Review, 0, len(raw))
	for _, rr := range raw {
		id := uuid.New().String()
		review := model.Review{ID: id, RawReview: rr}
		repo.reviews[id] = review
		repo.byContent[rr.ContentID] = append(repo.byContent[rr.ContentID], review)
		repo.byUser[rr.UserID] = append(repo.byUser[rr.UserID], review)
		out = append(out, review)
	}
	return out, nil
}

func (repo *InMemoryRepository) DeleteReview(id string) (bool, map[string]string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	r, ok := repo.reviews[id]
	if !ok {
		return false, nil
	}
	delete(repo.reviews, id)
	cSlice := repo.byContent[r.ContentID]
	for i := range cSlice {
		if cSlice[i].ID == id {
			cSlice = append(cSlice[:i], cSlice[i+1:]...)
			break
		}
	}
	if len(cSlice) == 0 {
		delete(repo.byContent, r.ContentID)
	} else {
		repo.byContent[r.ContentID] = cSlice
	}
	uSlice := repo.byUser[r.UserID]
	for i := range uSlice {
		if uSlice[i].ID == id {
			uSlice = append(uSlice[:i], uSlice[i+1:]...)
			break
		}
	}
	if len(uSlice) == 0 {
		delete(repo.byUser, r.UserID)
	} else {
		repo.byUser[r.UserID] = uSlice
	}
	return true, nil
}

func (repo *InMemoryRepository) ReviewsForContent(contentID string) []model.Review {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	return append([]model.Review(nil), repo.byContent[contentID]...)
}

func (repo *InMemoryRepository) ReviewsForUser(userID string) []model.Review {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	return append([]model.Review(nil), repo.byUser[userID]...)
}

func (repo *InMemoryRepository) AllContentIDs() []string {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	ids := make([]string, 0, len(repo.byContent))
	for k := range repo.byContent {
		ids = append(ids, k)
	}
	return ids
}
