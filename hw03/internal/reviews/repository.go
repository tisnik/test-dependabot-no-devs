package reviews

import "errors"

var (
	ErrReviewNotFound  = errors.New("review not found")
	ErrContentNotFound = errors.New("content not found")
)

type ReviewRepository interface {
	Save(content Content, review Review) error
	Delete(reviewID string) error
	GetAllReviews() ([]Review, error)
	GetReview(reviewID string) (Review, error)
	GetReviewsByUserID(userID string) ([]Review, error)
	GetReviewsByContentID(contentID string) ([]Review, error)
	GetAllContents() ([]Content, error)
	GetContent(contentID string) (Content, error)
}
