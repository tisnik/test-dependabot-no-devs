package recommender

import (
	"sort"
	"strings"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/storage"
	"github.com/google/uuid"
)

const (
	// wordMatchWeight determines the scoring impact of matching title words between
	// the user's profile and candidate content. Set to 2.0 to prioritize content
	// with similar titles, as word overlap is a strong indicator of content similarity.
	wordMatchWeight = 2.0

	// durationDiffThreshold (in minutes) defines the maximum duration difference
	// between candidate content and the user's average preferred duration to award
	// a similarity bonus. Set to 30 minutes to match content of similar length,
	// as users typically prefer consistent content duration.
	durationDiffThreshold = 30

	// highScoreThreshold defines the minimum score (out of 100) for content to
	// receive a quality bonus. Set to 70 to boost highly-rated content in
	// recommendations, ensuring quality standards are maintained.
	highScoreThreshold = 70
)

// Recommender generates content recommendations based on user preferences and
// content similarity using a profile-based scoring algorithm.
type Recommender struct {
	store *storage.ReviewStore
}

func NewRecommender(store *storage.ReviewStore) *Recommender {
	return &Recommender{store: store}
}

type contentScore struct {
	contentID uuid.UUID
	title     string
	score     float64
}

// Recommend generates recommendations based on content similarity.
func (r *Recommender) Recommend(contentID, userID uuid.UUID) []models.Recommendation {
	var targetReviews []models.Review
	var excludeContentID uuid.UUID

	if contentID != uuid.Nil {
		targetReviews = r.store.GetByContentID(contentID)
		excludeContentID = contentID
	} else if userID != uuid.Nil {
		targetReviews = r.store.GetByUserID(userID)
	}

	if len(targetReviews) == 0 {
		return []models.Recommendation{}
	}

	profile := BuildProfile(targetReviews)

	allReviews := r.store.GetAll()
	contentScores := make(map[uuid.UUID]*contentScore)

	for _, review := range allReviews {
		if review.ContentID == excludeContentID {
			continue
		}

		score := CalculateScore(profile, review)

		if existing, exists := contentScores[review.ContentID]; exists {
			existing.score += score
		} else {
			contentScores[review.ContentID] = &contentScore{
				contentID: review.ContentID,
				title:     review.Title,
				score:     score,
			}
		}
	}

	scores := make([]contentScore, 0, len(contentScores))
	for _, cs := range contentScores {
		scores = append(scores, *cs)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	recommendations := make([]models.Recommendation, 0, len(scores))
	for _, cs := range scores {
		recommendations = append(recommendations, models.Recommendation{
			ID:    cs.contentID,
			Title: cs.title,
		})
	}

	return recommendations
}

// Profile represents a user or content profile for recommendations.
type Profile struct {
	TitleWords   map[string]int
	AvgDuration  float64
	AvgScore     float64
	ReviewsCount int
}

// BuildProfile builds a profile from a set of reviews.
func BuildProfile(reviews []models.Review) Profile {
	prof := Profile{
		TitleWords: make(map[string]int),
	}

	totalDuration := 0
	totalScore := 0

	for _, review := range reviews {
		if review.Title != "" {
			for word := range strings.FieldsSeq(strings.ToLower(review.Title)) {
				prof.TitleWords[word]++
			}
		}

		totalDuration += review.Duration
		totalScore += review.Score
		prof.ReviewsCount++
	}

	if prof.ReviewsCount > 0 {
		prof.AvgDuration = float64(totalDuration) / float64(prof.ReviewsCount)
		prof.AvgScore = float64(totalScore) / float64(prof.ReviewsCount)
	}

	return prof
}

// CalculateScore calculates a similarity score between a profile and a review.
func CalculateScore(p Profile, review models.Review) float64 {
	score := 0.0

	if review.Title != "" {
		for word := range strings.FieldsSeq(strings.ToLower(review.Title)) {
			if count, exists := p.TitleWords[word]; exists {
				score += float64(count) * wordMatchWeight
			}
		}
	}

	if review.Duration > 0 && p.AvgDuration > 0 {
		durationDiff := float64(review.Duration) - p.AvgDuration
		if durationDiff < 0 {
			durationDiff = -durationDiff
		}
		// Smaller difference = higher score
		if durationDiff < durationDiffThreshold {
			score += 1.0
		}
	}

	if review.Score >= highScoreThreshold {
		score += 1.0
	}

	return score
}
