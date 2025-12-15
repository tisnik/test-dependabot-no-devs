package services

import (
	"fmt"
	"log/slog"

	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/request"
	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/response"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	validator  *validator.Validate
	repository *repository.Repository
}

func NewService(validator *validator.Validate, repository *repository.Repository) Service {
	return Service{
		validator:  validator,
		repository: repository,
	}
}

func (s Service) CreateReviews(reviews request.RawReviewsRequest) response.Reviews {
	slog.Info("Create review service called")

	return s.repository.SaveReview(reviews)
}

func (s Service) DeleteReview(id string) error {
	slog.Info("Delete review service called")

	err := s.repository.DeleteReview(id)
	if err != nil {
		return fmt.Errorf("failed to delete review with id %s: %w", id, err)
	}
	return nil
}
