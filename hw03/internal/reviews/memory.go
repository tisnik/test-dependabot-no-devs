package reviews

import (
	"sync"
)

type MemoryReviewRepository struct {
	mu       sync.RWMutex
	contents map[string]Content
	reviews  map[string]Review
}

func NewMemoryReviewRepository() *MemoryReviewRepository {
	return &MemoryReviewRepository{
		contents: make(map[string]Content),
		reviews:  make(map[string]Review),
	}
}

func (r *MemoryReviewRepository) Save(content Content, review Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.contents[content.ContentID]
	if !exists {
		r.contents[content.ContentID] = content
	} else if content.Title == "" {
		r.contents[content.ContentID] = content
	}

	r.reviews[review.ReviewID] = review
	return nil
}

func (r *MemoryReviewRepository) Delete(reviewID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reviews[reviewID]; !ok {
		return ErrReviewNotFound
	}
	delete(r.reviews, reviewID)
	return nil
}

func (r *MemoryReviewRepository) GetReview(reviewID string) (Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	review, ok := r.reviews[reviewID]
	if !ok {
		return Review{}, ErrReviewNotFound
	}
	return review, nil
}

func (r *MemoryReviewRepository) GetAllReviews() ([]Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reviews := make([]Review, 0, len(r.reviews))
	for _, review := range r.reviews {
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (r *MemoryReviewRepository) GetReviewsByUserID(userID string) ([]Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reviews := make([]Review, 0)
	for _, review := range r.reviews {
		if review.UserID == userID {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

func (r *MemoryReviewRepository) GetReviewsByContentID(contentID string) ([]Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reviews := make([]Review, 0)
	for _, review := range r.reviews {
		if review.ContentID == contentID {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

func (r *MemoryReviewRepository) GetContent(contentID string) (Content, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	content, ok := r.contents[contentID]
	if !ok {
		return Content{}, ErrContentNotFound
	}
	return content, nil
}

func (r *MemoryReviewRepository) GetAllContents() ([]Content, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	contents := make([]Content, 0, len(r.contents))
	for _, content := range r.contents {
		contents = append(contents, content)
	}
	return contents, nil
}
