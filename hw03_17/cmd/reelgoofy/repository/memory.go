package repository

import (
	"sync"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
)

type ReviewTable struct {
	mu      sync.RWMutex
	reviews map[string]structures.Review
}

func NewReviewTable() *ReviewTable {
	return &ReviewTable{reviews: make(map[string]structures.Review)}
}

func (table *ReviewTable) Add(review structures.Review) {
	table.mu.Lock()
	defer table.mu.Unlock()
	table.reviews[review.ID] = review
}

func (r *ReviewTable) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.reviews[id]
	if ok {
		delete(r.reviews, id)
		return true
	}
	return false
}

func (r *ReviewTable) GetAll() []structures.Review {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := []structures.Review{}
	for _, rev := range r.reviews {
		all = append(all, rev)
	}
	return all
}
