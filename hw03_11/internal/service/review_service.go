package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/repository"
)

type ReviewService struct {
	logger     *slog.Logger
	validate   *validator.Validate
	repository *repository.Repository
}

var ErrReviewNotFound = errors.New("review with given UUID does not exist")

func NewReviewService(
	logger *slog.Logger,
	validate *validator.Validate,
	repository *repository.Repository,
) (rs *ReviewService) {
	return &ReviewService{
		logger:     logger,
		validate:   validate,
		repository: repository,
	}
}

func (rs *ReviewService) IngestReviews(rawReviews *[]model.RawReview) ([]model.Review, error) {
	if rawReviews == nil {
		return []model.Review{}, nil
	}

	for _, rawReview := range *rawReviews {
		err := rs.validate.Struct(rawReview)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	reviews := make([]model.Review, 0, len(*rawReviews))

	for _, rawReview := range *rawReviews {
		review := rs.repository.CreateReview(&rawReview)
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (rs *ReviewService) DeleteReview(reviewId string) (err error) {
	err = rs.validate.Var(reviewId, "uuid")
	if err != nil {
		return fmt.Errorf("invalid reviewId: %w", err)
	}

	if !rs.repository.DeleteReview(reviewId) {
		return ErrReviewNotFound
	}

	return nil
}
