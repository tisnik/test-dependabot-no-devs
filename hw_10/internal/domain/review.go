package domain

import (
	"errors"

	"github.com/course-go/reelgoofy/internal/handler/dto"
	"github.com/google/uuid"
	"github.com/hardfinhq/go-date"
)

type Review struct {
	Id          uuid.UUID
	ContentId   uuid.UUID
	UserId      uuid.UUID
	Title       string
	Genres      []string
	Tags        []string
	Description string
	Director    string
	Actors      []string
	Origins     []string
	Duration    int
	Released    date.Date
	ReviewText  string
	Score       int
}

func FromReviewDTO(dto dto.ReviewDTO) (Review, error) {
	contentId, err := uuid.Parse(dto.ContentID)
	if err != nil {
		return Review{}, errors.New("invalid contentId")
	}

	userId, err := uuid.Parse(dto.UserID)
	if err != nil {
		return Review{}, errors.New("invalid userId")
	}

	released, err := date.FromString(dto.Released)
	if err != nil {
		return Review{}, errors.New("invalid released date")
	}

	return Review{
		ContentId:   contentId,
		UserId:      userId,
		Title:       dto.Title,
		Genres:      dto.Genres,
		Tags:        dto.Tags,
		Description: dto.Description,
		Director:    dto.Director,
		Actors:      dto.Actors,
		Origins:     dto.Origins,
		Duration:    dto.Duration,
		Released:    released,
		ReviewText:  dto.ReviewText,
		Score:       dto.Score,
	}, nil
}

func ToReviewDTO(review Review) dto.ReviewDTO {
	return dto.ReviewDTO{
		Id:          review.Id.String(),
		ContentID:   review.ContentId.String(),
		UserID:      review.UserId.String(),
		Title:       review.Title,
		Genres:      review.Genres,
		Tags:        review.Tags,
		Description: review.Description,
		Director:    review.Director,
		Actors:      review.Actors,
		Origins:     review.Origins,
		Duration:    review.Duration,
		Released:    review.Released.Format("2006-01-02"),
		ReviewText:  review.ReviewText,
		Score:       review.Score,
	}
}
