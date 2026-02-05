package reviews

import (
	"errors"
	"slices"
	"sync"

	"github.com/course-go/reelgoofy/internal/entity"
	"github.com/google/uuid"
)

var ErrReviewNotFound = errors.New("review with given ID does not exist")

type Repository struct {
	mutex   sync.Mutex
	reviews []entity.Review
}

func NewRepository() *Repository {
	return &Repository{
		reviews: make([]entity.Review, 0),
	}
}

func (repository *Repository) InsertReview(review entity.Review) (newReview entity.Review) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	review.ID = uuid.New()
	repository.reviews = append(repository.reviews, review)
	return review
}

func (repository *Repository) DeleteReview(id uuid.UUID) (err error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	index := slices.IndexFunc(repository.reviews, func(review entity.Review) bool {
		return id == review.ID
	})
	if index == -1 {
		return ErrReviewNotFound
	}

	repository.reviews = slices.Delete(repository.reviews, index, index+1)
	return nil
}

func (repository *Repository) GetReviews() (reviews []entity.Review) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	return repository.reviews
}
