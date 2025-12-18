package store

import (
	"sync"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/google/uuid"
)

type Repository interface {
	AddReviews(raw []model.RawReview) ([]model.Review, map[string]string)
	DeleteReview(id string) (bool, map[string]string)
}

type InMemoryRepository struct {
	mu        sync.RWMutex
	reviews   map[string]model.Review
	byContent map[string][]model.Review
	byUser    map[string][]model.Review
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		reviews:   make(map[string]model.Review),
		byContent: make(map[string][]model.Review),
		byUser:    make(map[string][]model.Review),
	}
}

func (repository *InMemoryRepository) AddReviews(raw []model.RawReview) ([]model.Review, map[string]string) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	out := make([]model.Review, 0, len(raw))
	for _, rr := range raw {
		id := uuid.New().String()
		review := model.Review{ID: id, RawReview: rr}
		repository.reviews[id] = review
		repository.byContent[rr.ContentID] = append(repository.byContent[rr.ContentID], review)
		repository.byUser[rr.UserID] = append(repository.byUser[rr.UserID], review)
		out = append(out, review)
	}
	return out, nil
}

func (repository *InMemoryRepository) DeleteReview(id string) (bool, map[string]string) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	r, ok := repository.reviews[id]
	if !ok {
		return false, nil
	}
	delete(repository.reviews, id)
	cSlice := repository.byContent[r.ContentID]
	for i := range cSlice {
		if cSlice[i].ID == id {
			cSlice = append(cSlice[:i], cSlice[i+1:]...)
			break
		}
	}
	if len(cSlice) == 0 {
		delete(repository.byContent, r.ContentID)
	} else {
		repository.byContent[r.ContentID] = cSlice
	}
	uSlice := repository.byUser[r.UserID]
	for i := range uSlice {
		if uSlice[i].ID == id {
			uSlice = append(uSlice[:i], uSlice[i+1:]...)
			break
		}
	}
	if len(uSlice) == 0 {
		delete(repository.byUser, r.UserID)
	} else {
		repository.byUser[r.UserID] = uSlice
	}
	return true, nil
}

func (repository *InMemoryRepository) ReviewsForContent(contentID string) []model.Review {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	return append([]model.Review(nil), repository.byContent[contentID]...)
}

func (repository *InMemoryRepository) ReviewsForUser(userID string) []model.Review {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	return append([]model.Review(nil), repository.byUser[userID]...)
}

func (repository *InMemoryRepository) AllContentIDs() []string {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	ids := make([]string, 0, len(repository.byContent))
	for k := range repository.byContent {
		ids = append(ids, k)
	}
	return ids
}
