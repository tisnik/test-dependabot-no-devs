package repository

import (
	"fmt"

	"github.com/course-go/reelgoofy/internal/domain"
	"github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/repository/dto"
	"github.com/google/uuid"
)

type ReviewRepository struct {
	data map[uuid.UUID]domain.Review
}

func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{
		data: make(map[uuid.UUID]domain.Review),
	}
}

func (r *ReviewRepository) SaveBatch(reviews []domain.Review) ([]domain.Review, error) {
	for i, review := range reviews {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to generate UUID: %w", err)
		}
		r.data[uuid] = review
		reviews[i].Id = uuid
	}
	return reviews, nil
}

func (r *ReviewRepository) Delete(id uuid.UUID) error {
	_, exists := r.data[id]
	if !exists {
		return errors.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

func (r *ReviewRepository) FindMovies() []dto.MovieDataDTO {
	movies := make(map[uuid.UUID]dto.MovieDataDTO)
	for _, review := range r.data {
		movie, exists := movies[review.ContentId]
		if !exists {
			movie = dto.MovieDataDTO{
				ContentId: review.ContentId,
				Title:     review.Title,
				Director:  review.Director,
				Genres:    review.Genres,
				Ratings:   make([]int, 0),
			}
		}
		movie.Ratings = append(movie.Ratings, review.Score)
		movies[review.ContentId] = movie
	}

	result := make([]dto.MovieDataDTO, 0, len(movies))
	for _, movie := range movies {
		result = append(result, movie)
	}

	return result
}

func (r *ReviewRepository) FindMovieByContentId(contentId uuid.UUID) (dto.MovieDataDTO, error) {
	matchedReviews := make([]domain.Review, 0)
	for _, review := range r.data {
		if review.ContentId == contentId {
			matchedReviews = append(matchedReviews, review)
		}
	}

	var movie dto.MovieDataDTO
	if len(matchedReviews) > 0 {
		sampleReview := matchedReviews[0]
		movie.ContentId = sampleReview.ContentId
		movie.Title = sampleReview.Title
		movie.Director = sampleReview.Director
		movie.Genres = sampleReview.Genres
	} else {
		return dto.MovieDataDTO{}, errors.ErrNotFound
	}

	for _, review := range matchedReviews {
		movie.Ratings = append(movie.Ratings, review.Score)
	}

	return movie, nil
}

func (r *ReviewRepository) FindGenresByUser(id uuid.UUID) ([]string, error) {
	if !r.findUserExists(id) {
		return nil, errors.ErrNotFound
	}

	genreSet := make(map[string]struct{})
	for _, review := range r.data {
		if review.UserId == id {
			for _, genre := range review.Genres {
				genreSet[genre] = struct{}{}
			}
		}
	}

	genres := make([]string, 0, len(genreSet))
	for genre := range genreSet {
		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *ReviewRepository) FindDirectorsByUser(id uuid.UUID) ([]string, error) {
	if !r.findUserExists(id) {
		return nil, errors.ErrNotFound
	}

	directorSet := make(map[string]struct{})
	for _, review := range r.data {
		if review.UserId == id {
			directorSet[review.Director] = struct{}{}
		}
	}

	directors := make([]string, 0, len(directorSet))
	for director := range directorSet {
		directors = append(directors, director)
	}

	return directors, nil
}

func (r *ReviewRepository) findUserExists(userId uuid.UUID) bool {
	for _, review := range r.data {
		if review.UserId == userId {
			return true
		}
	}
	return false
}
