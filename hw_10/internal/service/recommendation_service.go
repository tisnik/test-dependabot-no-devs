package service

import (
	"fmt"
	"slices"

	hanlerdto "github.com/course-go/reelgoofy/internal/handler/dto"
	"github.com/course-go/reelgoofy/internal/repository"
	"github.com/course-go/reelgoofy/internal/repository/dto"
	"github.com/google/uuid"
)

type RecommendationService struct {
	repository *repository.ReviewRepository
}

func NewRecommendationService(repository *repository.ReviewRepository) *RecommendationService {
	return &RecommendationService{
		repository: repository,
	}
}

func (s *RecommendationService) GetRecommendationsByUser(
	userId uuid.UUID,
	limit int,
	offset int,
) ([]hanlerdto.RecommendationDTO, error) {
	movies := s.repository.FindMovies()
	usedMovieIds := make(map[uuid.UUID]struct{})

	directors, err := s.repository.FindDirectorsByUser(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find directors: %w", err)
	}
	moviesMathingDirectors, usedMovieIds := s.getRecommendedByDirectors(directors, movies, usedMovieIds)

	genres, err := s.repository.FindGenresByUser(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find genres: %w", err)
	}
	moviesMathingGenres, usedMovieIds := s.getRecommendedByGenres(genres, movies, usedMovieIds)

	otherMovies := make([]dto.MovieDataDTO, 0)
	for _, movie := range movies {
		_, used := usedMovieIds[movie.ContentId]
		if used {
			continue
		}
		otherMovies = append(otherMovies, movie)
	}

	slices.SortFunc(otherMovies, func(a dto.MovieDataDTO, b dto.MovieDataDTO) int {
		return b.AverageRating() - a.AverageRating()
	})

	recommended := make([]dto.MovieDataDTO, 0, len(moviesMathingDirectors)+len(moviesMathingGenres)+len(otherMovies))
	recommended = append(recommended, moviesMathingDirectors...)
	recommended = append(recommended, moviesMathingGenres...)
	recommended = append(recommended, otherMovies...)
	if offset >= len(recommended) {
		return make([]hanlerdto.RecommendationDTO, 0), nil
	}
	recommended = recommended[offset:]

	if limit > 0 && limit < len(recommended) {
		recommended = recommended[:limit]
	}

	result := make([]hanlerdto.RecommendationDTO, 0, len(recommended))
	for _, movie := range recommended {
		result = append(result, hanlerdto.RecommendationDTO{
			Id:    movie.ContentId,
			Title: movie.Title,
		})
	}

	return result, nil
}

func (s *RecommendationService) GetRecommendationsByContent(
	contentId uuid.UUID,
	limit int,
	offset int,
) ([]hanlerdto.RecommendationDTO, error) {
	contentMovie, err := s.repository.FindMovieByContentId(contentId)
	if err != nil {
		return nil, fmt.Errorf("failed to find movie: %w", err)
	}

	movies := s.repository.FindMovies()
	usedMovieIds := make(map[uuid.UUID]struct{})
	usedMovieIds[contentMovie.ContentId] = struct{}{}

	directors := make([]string, 0)
	directors = append(directors, contentMovie.Director)
	moviesMathingDirectors, usedMovieIds := s.getRecommendedByDirectors(directors, movies, usedMovieIds)

	moviesMathingGenres, usedMovieIds := s.getRecommendedByGenres(contentMovie.Genres, movies, usedMovieIds)

	otherMovies := make([]dto.MovieDataDTO, 0)
	for _, movie := range movies {
		_, used := usedMovieIds[movie.ContentId]
		if used {
			continue
		}
		otherMovies = append(otherMovies, movie)
	}

	slices.SortFunc(otherMovies, func(a dto.MovieDataDTO, b dto.MovieDataDTO) int {
		return b.AverageRating() - a.AverageRating()
	})

	recommended := make([]dto.MovieDataDTO, 0, len(moviesMathingDirectors)+len(moviesMathingGenres)+len(otherMovies))
	recommended = append(recommended, moviesMathingDirectors...)
	recommended = append(recommended, moviesMathingGenres...)
	recommended = append(recommended, otherMovies...)
	if offset >= len(recommended) {
		return make([]hanlerdto.RecommendationDTO, 0), nil
	}
	recommended = recommended[offset:]

	if limit > 0 && limit < len(recommended) {
		recommended = recommended[:limit]
	}

	result := make([]hanlerdto.RecommendationDTO, 0, len(recommended))
	for _, movie := range recommended {
		result = append(result, hanlerdto.RecommendationDTO{
			Id:    movie.ContentId,
			Title: movie.Title,
		})
	}

	return result, nil
}

func (s *RecommendationService) getRecommendedByDirectors(
	directors []string,
	movies []dto.MovieDataDTO,
	excludedSet map[uuid.UUID]struct{},
) ([]dto.MovieDataDTO, map[uuid.UUID]struct{}) {
	moviesMathingDirectors := make([]dto.MovieDataDTO, 0)
	for _, movie := range movies {
		for _, director := range directors {
			if movie.Director == director {
				_, used := excludedSet[movie.ContentId]
				if used {
					break
				}
				moviesMathingDirectors = append(moviesMathingDirectors, movie)
				excludedSet[movie.ContentId] = struct{}{}
				break
			}
		}
	}

	slices.SortFunc(moviesMathingDirectors, func(a dto.MovieDataDTO, b dto.MovieDataDTO) int {
		return b.AverageRating() - a.AverageRating()
	})

	return moviesMathingDirectors, excludedSet
}

func (s *RecommendationService) getRecommendedByGenres(
	genres []string,
	movies []dto.MovieDataDTO,
	excludedSet map[uuid.UUID]struct{},
) ([]dto.MovieDataDTO, map[uuid.UUID]struct{}) {
	moviesMathingGenres := make([]dto.MovieDataDTO, 0)
	for _, movie := range movies {
		for _, genre := range movie.Genres {
			if slices.Contains(genres, genre) {
				_, used := excludedSet[movie.ContentId]
				if used {
					break
				}
				moviesMathingGenres = append(moviesMathingGenres, movie)
				excludedSet[movie.ContentId] = struct{}{}
				break
			}
		}
	}

	slices.SortFunc(moviesMathingGenres, func(a dto.MovieDataDTO, b dto.MovieDataDTO) int {
		return b.AverageRating() - a.AverageRating()
	})

	return moviesMathingGenres, excludedSet
}
