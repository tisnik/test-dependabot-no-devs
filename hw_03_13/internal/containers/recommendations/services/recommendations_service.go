package services

import (
	"math"
	"sort"

	recommendationDto "github.com/course-go/reelgoofy/internal/containers/recommendations/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	"github.com/course-go/reelgoofy/internal/containers/reviews/services"
	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/google/uuid"
)

const (
	percentage            = 10
	recommendationsAmount = 3
)

// RecommendByContentId returns content recommendations based on similarity to the specified content.
// It finds similar items using cosine similarity of content features.
func RecommendByContentId(
	id uuid.UUID,
	repo repository.ReviewRepository,
	amount *int,
	offset *int,
) ([]recommendationDto.Recommendation, map[string]string) {
	review, ok := repo.GetFirstReviewByContentID(id)
	if !ok {
		return nil, map[string]string{
			"contentId": string(response.ContentIdNotFound),
		}
	}

	reviews := repo.GetReviews()
	scored := services.OrderReviewsBySimilarity(review.SimilarityVector, reviews, &review.ContentId)

	return formatRecommendations(scored, amount, offset), nil
}

// RecommendByUserId returns content recommendations based on similarity to user preferences.
// It takes into consideration user's top-scored reviews and finds similar items using cosine similarity.
func RecommendByUserId(
	id uuid.UUID,
	repo repository.ReviewRepository,
	amount *int,
	offset *int,
) ([]recommendationDto.Recommendation, map[string]string) {
	userReviews, ok := repo.GetReviewsByUserId(id)
	if !ok {
		return nil, map[string]string{
			"userId": string(response.UserIdNotFound),
		}
	}

	allReviews := repo.GetReviews()

	// Order user userReviews by achieved score
	sort.Slice(*userReviews, func(i, j int) bool {
		return (*userReviews)[i].Score > (*userReviews)[j].Score
	})

	// Get 10% of user's highest-scored reviews
	top10PercentReviews := max(1, len(*userReviews)/percentage)
	topReviews := (*userReviews)[:top10PercentReviews]

	// Build user's preference profile by averaging similarity vectors of top-rated reviews
	profile := services.BuildUserSimilarityVector(topReviews)

	// Order reviews and then filter out those, which were created by the user
	scored := services.OrderReviewsBySimilarity(profile, allReviews, nil)
	filtered := scored[:0]
	for _, s := range scored {
		if s.Review.UserId != id.String() {
			filtered = append(filtered, s)
		}
	}
	scored = filtered

	return formatRecommendations(scored, amount, offset), nil
}

// formatRecommendations returns a formatted response.
func formatRecommendations(scored []services.ScoredReview, amount, offset *int) []recommendationDto.Recommendation {
	limit := recommendationsAmount
	if amount != nil {
		limit = *amount
	}

	if limit > len(scored) {
		limit = len(scored)
	}

	start := 0
	if offset != nil && *offset >= 0 {
		start = *offset
	}

	if start > len(scored) {
		return []recommendationDto.Recommendation{}
	}

	end := start + limit
	end = int(math.Min(float64(end), float64(len(scored))))

	topReviews := scored[start:end]
	recommendations := make([]recommendationDto.Recommendation, 0, len(topReviews))
	for _, r := range topReviews {
		recommendations = append(recommendations, recommendationDto.Recommendation{
			ID:    r.Review.ID,
			Title: r.Review.Title,
		})
	}
	return recommendations
}
