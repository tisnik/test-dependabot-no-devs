package main

import "sync"

type Storage struct {
	mu      sync.RWMutex
	reviews map[string]*Review
}

func NewStorage() *Storage {
	return &Storage{
		reviews: make(map[string]*Review),
	}
}

func (s *Storage) Add(review *Review) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reviews[review.Id] = review
}

func (s *Storage) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.reviews[id]; exists {
		delete(s.reviews, id)
		return true
	}
	return false
}

func (s *Storage) GetAll() []*Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Review, 0, len(s.reviews))
	for _, review := range s.reviews {
		result = append(result, review)
	}
	return result
}

func (s *Storage) GetByContentID(contentID string) []*Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Review
	for _, review := range s.reviews {
		if review.ContentId == contentID {
			result = append(result, review)
		}
	}
	return result
}

func (s *Storage) GetByUserID(userID string) []*Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Review
	for _, review := range s.reviews {
		if review.UserId == userID {
			result = append(result, review)
		}
	}
	return result
}
