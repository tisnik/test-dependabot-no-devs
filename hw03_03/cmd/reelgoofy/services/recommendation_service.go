package services

import (
	"sort"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/structs"
)

const scoreNormalizer = 100.0

type RecommendationService struct {
	ReviewService *ReviewService
}

type scoredContent struct {
	contentID string
	score     float64
}

func NewRecommendationService(reviewService *ReviewService) *RecommendationService {
	return &RecommendationService{
		ReviewService: reviewService,
	}
}

// RecommendContentToContent gives score based on shared users and genres then sorts by score.
func (s *RecommendationService) RecommendContentToContent(
	contentID string,
	limit, offset int,
) []structs.Recommendation {
	targetReviews := s.ReviewService.GetReviewsByContent(contentID)
	if len(targetReviews) == 0 {
		return []structs.Recommendation{}
	}

	userSet := make(map[string]bool)
	genreSet := make(map[string]bool)

	for _, review := range targetReviews {
		userSet[review.UserID] = true
		for _, genre := range review.Genres {
			genreSet[genre] = true
		}
	}

	scoreMap := make(map[string]float64)
	titleMap := make(map[string]string)
	allReviews := s.ReviewService.GetAllReviews()

	for _, review := range allReviews {
		if review.ContentID == contentID {
			continue
		}

		score := 0.0
		if userSet[review.UserID] {
			score += float64(review.Score) / scoreNormalizer
		}

		for _, genre := range review.Genres {
			if genreSet[genre] {
				score += 0.5
			}
		}

		scoreMap[review.ContentID] += score
		titleMap[review.ContentID] = review.Title
	}

	scored := make([]scoredContent, 0, len(scoreMap))
	for cID, score := range scoreMap {
		scored = append(scored, scoredContent{cID, score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return paginateRecommendations(scored, titleMap, limit, offset)
}

// RecommendContentToUser takes ratings on genres and then sorts based on score for genres.
func (s *RecommendationService) RecommendContentToUser(
	userID string,
	limit, offset int,
) []structs.Recommendation {
	userReviews := s.ReviewService.GetReviewsByUser(userID)
	if len(userReviews) == 0 {
		return []structs.Recommendation{}
	}

	genreScores := make(map[string]float64)
	viewedContent := make(map[string]bool)

	for _, review := range userReviews {
		viewedContent[review.ContentID] = true
		normalizedScore := float64(review.Score) / scoreNormalizer

		for _, genre := range review.Genres {
			genreScores[genre] += normalizedScore
		}
	}

	scoreMap := make(map[string]float64)
	titleMap := make(map[string]string)
	allReviews := s.ReviewService.GetAllReviews()

	for _, review := range allReviews {
		if viewedContent[review.ContentID] {
			continue
		}

		score := 0.0
		for _, genre := range review.Genres {
			if genreScore, exists := genreScores[genre]; exists {
				score += genreScore
			}
		}

		if score > 0 {
			scoreMap[review.ContentID] = score
			titleMap[review.ContentID] = review.Title
		}
	}

	scored := make([]scoredContent, 0, len(scoreMap))
	for cID, score := range scoreMap {
		scored = append(scored, scoredContent{cID, score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return paginateRecommendations(scored, titleMap, limit, offset)
}

func paginateRecommendations(
	scored []scoredContent,
	titleMap map[string]string,
	limit, offset int,
) []structs.Recommendation {
	if offset >= len(scored) {
		return []structs.Recommendation{}
	}

	end := offset + limit
	if limit <= 0 || end > len(scored) {
		end = len(scored)
	}

	var recommendations []structs.Recommendation
	for i := offset; i < end; i++ {
		recommendations = append(recommendations, structs.Recommendation{
			ID:    scored[i].contentID,
			Title: titleMap[scored[i].contentID],
		})
	}

	return recommendations
}
