package services

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"

	"github.com/course-go/reelgoofy/internal/http/controllers/recommendations/dto/response"
	"github.com/course-go/reelgoofy/internal/http/controllers/reviews/dto/request"
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

type scoredRecommendation struct {
	ContentID string
	UserID    string
	Title     string
	Score     int
}

func (s Service) GetContentToContent(
	contentId string,
	limitStr string,
	offsetStr string,
) (response.Recommendations, error) {
	slog.Info("GetContentToContent service called")

	allReviews := s.repository.GetAllReviews()

	var targetReview *request.Review
	for i, review := range allReviews { // find stats of movie
		if review.ContentID == contentId {
			targetReview = &allReviews[i]
			break
		}
	}

	if targetReview == nil {
		return response.Recommendations{}, errors.New("ID not found")
	}

	recommendations := recommendedContent(*targetReview, contentId, allReviews)
	slog.Info("Recommendation ", "recom", recommendations)

	sort.Slice(recommendations, func(i int, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	start, end, err := getRangeOfRecommendations(limitStr, offsetStr, len(recommendations))
	if err != nil {
		return response.Recommendations{}, err
	}

	finalRecommendations := recommendations[start:end]

	responseRecommendation := make([]response.Recommendation, len(finalRecommendations))
	for i, rec := range finalRecommendations {
		responseRecommendation[i] = response.Recommendation{ID: rec.ContentID, Title: rec.Title}
	}

	return response.Recommendations{Recommendations: responseRecommendation}, nil
}

func (s Service) GetContentToUser(
	userId string,
	limitStr string,
	offsetStr string,
) (response.Recommendations, error) {
	slog.Info("GetContentToUser service called")

	allReviews := s.repository.GetAllReviews()
	userReviews := s.repository.GetReviewsByUserID(userId)

	if len(userReviews) == 0 {
		return response.Recommendations{}, errors.New("no reviews found for this user")
	}

	var highlyRatedContent []request.Review
	const highlyRatedScore = 75
	for i, review := range userReviews {
		if review.Score > highlyRatedScore { // leave only ratings 7+
			highlyRatedContent = append(highlyRatedContent, userReviews[i])
		}
	}

	if len(highlyRatedContent) == 0 {
		return response.Recommendations{}, errors.New(
			"no highly rated content found for this user to base recommendations on",
		)
	}

	aggregatedRecommendations := make(map[string]scoredRecommendation)
	for _, targetReview := range highlyRatedContent {
		recommendations := recommendedContent(targetReview, targetReview.ContentID, allReviews)

		for _, recom := range recommendations {
			if recom.UserID == userId { // skip already reviewed
				continue
			}
			existingRecom := aggregatedRecommendations[recom.ContentID]
			existingRecom.ContentID = recom.ContentID
			existingRecom.UserID = recom.UserID
			existingRecom.Title = recom.Title
			existingRecom.Score += recom.Score // adding score to movies
			aggregatedRecommendations[recom.ContentID] = existingRecom
		}
	}

	aggregatedToSliceRec := make([]scoredRecommendation, 0, len(aggregatedRecommendations))
	for _, rec := range aggregatedRecommendations {
		aggregatedToSliceRec = append(aggregatedToSliceRec, rec)
	}

	sort.Slice(aggregatedToSliceRec, func(i int, j int) bool {
		return aggregatedToSliceRec[i].Score > aggregatedToSliceRec[j].Score
	})

	start, end, err := getRangeOfRecommendations(limitStr, offsetStr, len(aggregatedToSliceRec))
	if err != nil {
		return response.Recommendations{}, err
	}

	finalRecommendations := aggregatedToSliceRec[start:end]

	responseRecommendation := make([]response.Recommendation, len(finalRecommendations))
	for i, rec := range finalRecommendations {
		responseRecommendation[i] = response.Recommendation{ID: rec.ContentID, Title: rec.Title}
	}

	return response.Recommendations{Recommendations: responseRecommendation}, nil
}

func recommendedContent(
	targetReview request.Review,
	contentId string,
	allReviews []request.Review,
) []scoredRecommendation {
	targetGenres := make(map[string]struct{})
	for _, genre := range targetReview.Genres {
		targetGenres[genre] = struct{}{}
	}
	targetTags := make(map[string]struct{})
	for _, tag := range targetReview.Tags {
		targetTags[tag] = struct{}{}
	}

	var recommendations []scoredRecommendation
	for _, review := range allReviews {
		if review.ContentID == contentId {
			continue
		}

		recommendationScore := 0

		// Genres
		for _, genre := range review.Genres {
			if _, found := targetGenres[genre]; found {
				recommendationScore += 3
			}
		}

		// Tags
		for _, tag := range review.Tags {
			if _, found := targetTags[tag]; found {
				recommendationScore += 2
			}
		}

		// Director
		if review.Director != "" && review.Director == targetReview.Director {
			recommendationScore++
		}

		if recommendationScore > 0 {
			recommendations = append(recommendations, scoredRecommendation{
				ContentID: review.ContentID,
				UserID:    review.UserID,
				Title:     review.Title,
				Score:     recommendationScore,
			})
		}
	}

	return recommendations
}

func getRangeOfRecommendations(
	limitStr string,
	offsetStr string,
	lenOfRecom int,
) (int, int, error) {
	var err error

	limit := 20 // Default limit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return -1, -1, fmt.Errorf("failed to parse limit: %w", err)
		}
		if limit <= 0 {
			return -1, -1, errors.New("limit cannot be negative or zero")
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return -1, -1, fmt.Errorf("failed to parse offset: %w", err)
		}
		if offset < 0 {
			return -1, -1, errors.New("offser cannot be negative")
		}
	}

	start := offset
	start = min(start, lenOfRecom)

	end := start + limit
	end = min(end, lenOfRecom)

	return start, end, nil
}
