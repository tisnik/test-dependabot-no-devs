package service

import (
	"sort"

	"github.com/course-go/reelgoofy/internal/domain"
)

const (
	initRecommendationWithScoreCapacity = 30
)

func (s *ReviewService) RecommendContentToContent(contentId string) ([]domain.Recommendation, error) {
	reviews := s.store.GetReviewsByContent(contentId)
	recommendationWithScores := make(map[string]domain.RecommendationWithScore, initRecommendationWithScoreCapacity)
	if len(reviews) == 0 {
		return nil, ErrNotFound
	}
	for _, review := range reviews {
		newRecommendationWithScores := s.getUserContentExcludeContent(review.UserID, contentId)
		for key, value := range newRecommendationWithScores {
			current, ok := recommendationWithScores[key]
			if !ok || current.Score < value.Score {
				recommendationWithScores[key] = value
			}
		}
	}
	return mapToSlice(recommendationWithScores), nil
}

func (s *ReviewService) RecommendContentToUser(userId string) ([]domain.Recommendation, error) {
	recommendationWithScores := make(map[string]domain.RecommendationWithScore, initRecommendationWithScoreCapacity)
	reviews := s.store.GetReviewsByUser(userId)
	if len(reviews) == 0 {
		return nil, ErrNotFound
	}
	for _, userReview := range reviews { // review of user
		for _, otherReview := range s.store.GetReviewsByContent(userReview.ContentID) { // other review for the same film
			newRecommendationWithScores := s.getUserContent(otherReview.UserID)
			for key, value := range newRecommendationWithScores {
				current, ok := recommendationWithScores[key]
				if !ok || current.Score < value.Score {
					recommendationWithScores[key] = value
				}
			}
		}
	}

	exclude := s.store.GetReviewsByUserMap(userId)
	for key := range recommendationWithScores {
		if _, exists := exclude[key]; exists {
			delete(recommendationWithScores, key)
		}
	}
	return mapToSlice(recommendationWithScores), nil
}

func (s *ReviewService) getUserContent(userId string) map[string]domain.RecommendationWithScore {
	reviews := s.store.GetReviewsByUser(userId)
	recommendationWithScores := make(map[string]domain.RecommendationWithScore, len(reviews))
	for _, review := range reviews {
		recommendationWithScores[review.ContentID] = domain.RecommendationWithScore{
			Recommendation: domain.Recommendation{
				ID:    review.ContentID,
				Title: review.Title,
			},
			Score: review.Score,
		}
	}
	return recommendationWithScores
}

func (s *ReviewService) getUserContentExcludeContent(
	userId string,
	excludeContentId string,
) map[string]domain.RecommendationWithScore {
	reviews := s.store.GetReviewsByUser(userId)
	recommendationWithScores := make(map[string]domain.RecommendationWithScore, len(reviews))
	for _, review := range reviews {
		if review.ContentID == excludeContentId {
			continue
		}
		recommendationWithScores[review.ContentID] = domain.RecommendationWithScore{
			Recommendation: domain.Recommendation{
				ID:    review.ContentID,
				Title: review.Title,
			},
			Score: review.Score,
		}
	}
	return recommendationWithScores
}

func mapToSlice(recommendationsMap map[string]domain.RecommendationWithScore) []domain.Recommendation {
	recommendationsWithScore := make([]domain.RecommendationWithScore, 0, len(recommendationsMap))
	for _, rec := range recommendationsMap {
		recommendationsWithScore = append(recommendationsWithScore, rec)
	}
	sort.Slice(recommendationsWithScore, func(i, j int) bool {
		return recommendationsWithScore[i].Score > recommendationsWithScore[j].Score
	})
	recommendations := make([]domain.Recommendation, 0, len(recommendationsWithScore))
	for _, rec := range recommendationsWithScore {
		recommendations = append(recommendations, domain.Recommendation{ID: rec.ID, Title: rec.Title})
	}
	return recommendations
}
