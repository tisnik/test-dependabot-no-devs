package reviews

import (
	"errors"
	"fmt"

	apperrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/google/uuid"
)

type ReviewService struct {
	repo ReviewRepository
}

func NewReviewService(repo ReviewRepository) *ReviewService {
	return &ReviewService{
		repo: repo,
	}
}

func (s *ReviewService) Ingest(raws []dto.RawReviewDTO) (dto.ReviewsDTO, error) {
	revs := make([]dto.ReviewDTO, len(raws))
	for i, raw := range raws {
		id := uuid.NewString()
		content, review := mapRawReviewToDomain(raw, id)
		err := s.repo.Save(content, review)
		if err != nil {
			return dto.ReviewsDTO{}, fmt.Errorf("failed to save review: %w", err)
		}
		revs[i] = dto.ReviewDTO{
			ID:           id,
			RawReviewDTO: raw,
		}
	}
	return dto.ReviewsDTO{Reviews: revs}, nil
}

func (s *ReviewService) Delete(id string) error {
	err := s.repo.Delete(id)
	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			return apperrors.NewNotFoundError("reviewId", "Review with such ID not found.")
		}
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

func mapRawReviewToDomain(raw dto.RawReviewDTO, id string) (Content, Review) {
	content := Content{
		ContentID:   raw.ContentID,
		Title:       raw.Title,
		Genres:      raw.Genres,
		Tags:        raw.Tags,
		Description: raw.Description,
		Director:    raw.Director,
		Actors:      raw.Actors,
		Origins:     raw.Origins,
		Duration:    raw.Duration,
		Released:    raw.Released,
	}

	review := Review{
		ReviewID:  id,
		ContentID: raw.ContentID,
		UserID:    raw.UserID,
		Comment:   raw.ReviewComment,
		Score:     raw.Score,
	}

	return content, review
}
