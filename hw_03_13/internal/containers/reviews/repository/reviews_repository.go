package repository

import (
	"sync"

	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/services"
	"github.com/google/uuid"
)

type ReviewRepository interface {
	AddReview(r dto.RawReview) dto.Review
	GetReviews() []dto.Review
	GetReviewById(id string) (*dto.Review, bool)
	DeleteReview(id string) bool
	GetFirstReviewByContentID(contentID uuid.UUID) (*dto.Review, bool)
	GetReviewsByUserId(userId uuid.UUID) (*[]dto.Review, bool)
}

// Repository is an in-memory storage implementation.
// Used as a temporary solution before database integration.
type Repository struct {
	mu         sync.RWMutex
	reviewsMap map[string]dto.Review
}

func NewReviewRepository() *Repository {
	return &Repository{
		reviewsMap: make(map[string]dto.Review),
	}
}

// AddReview receives a user-created RawReview, creates a SimilarityVector
// and stores it as a new Review into the storage.
func (rr *Repository) AddReview(r dto.RawReview) dto.Review {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	id := uuid.New().String()
	vec := services.BuildVector(r)
	review := dto.Review{
		ID:               id,
		RawReview:        r,
		SimilarityVector: vec,
	}
	rr.reviewsMap[id] = review
	return review
}

// GetReviews returns a collection of all reviews.
func (rr *Repository) GetReviews() []dto.Review {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	allReviews := make([]dto.Review, 0, len(rr.reviewsMap))
	for _, v := range rr.reviewsMap {
		allReviews = append(allReviews, v)
	}

	return allReviews
}

// GetReviewById returns a collection of all reviews.
func (rr *Repository) GetReviewById(id string) (*dto.Review, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	_, ok := rr.reviewsMap[id]
	if ok {
		val := rr.reviewsMap[id]
		return &val, true
	}

	return nil, false
}

// DeleteReview deletes a review by ID if possible.
func (rr *Repository) DeleteReview(id string) bool {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if _, ok := rr.reviewsMap[id]; ok {
		delete(rr.reviewsMap, id)
		return true
	}
	return false
}

// GetFirstReviewByContentID returns first review with matching contentID. First is returned since
// no validation is checking for duplicit keys. This will be replaced by query select when DB is
// implemented.
func (rr *Repository) GetFirstReviewByContentID(contentID uuid.UUID) (*dto.Review, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for _, r := range rr.reviewsMap {
		if r.ContentId == contentID.String() {
			return &r, true
		}
	}

	return nil, false
}

// GetReviewsByUserId returns first a collection of reviews by the same author (with matching userID).
// This will be replaced by query select when DB is implemented.
func (rr *Repository) GetReviewsByUserId(userId uuid.UUID) (*[]dto.Review, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	var reviews []dto.Review

	for _, r := range rr.reviewsMap {
		if r.UserId == userId.String() {
			reviews = append(reviews, r)
		}
	}

	if len(reviews) == 0 {
		return nil, false
	}

	return &reviews, true
}
