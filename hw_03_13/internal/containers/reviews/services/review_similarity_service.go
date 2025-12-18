package services

import (
	"math"
	"slices"
	"sort"
	"time"

	"github.com/course-go/reelgoofy/internal/containers/reviews/dto"
	"github.com/course-go/reelgoofy/internal/containers/reviews/enums"
)

const (
	maxDuration            = 240
	yearNormalization      = 2000
	cosineSimilarityWeight = 0.7
	scoreWeight            = 0.3
	normalization          = 100
	// + 1 is for "other" countries, + 2 is for released date and duration.
	similarVectorStaticLength = 3
)

type RecommendationConfig struct {
	GenreWeight    float64
	ActorWeight    float64
	DirectorWeight float64
	TagWeight      float64
	MinScore       int
}

type ScoredReview struct {
	Review     dto.Review
	Similarity float64
}

// BuildVector takes review and converts certain parts related to content
// into a vector which is then used for finding similar reviews.
func BuildVector(rawReview dto.RawReview) dto.SimilarityVector {
	genres := enums.Genres
	countries := enums.Countries

	vec := make(dto.SimilarityVector, len(genres)+len(countries)+similarVectorStaticLength)

	// Genres
	for i, genre := range genres {
		vec[i] = 0
		if slices.IndexFunc(rawReview.Genres, func(g enums.Genre) bool { return g == genre }) != -1 {
			vec[i] = 1
		}
	}

	// Origins (only popular countries)
	noCountrySelected := true
	for i, country := range countries {
		vec[len(genres)+i] = 0
		if slices.IndexFunc(rawReview.Origins, func(o string) bool { return o == string(country) }) != -1 {
			vec[len(genres)+i] = 1
			noCountrySelected = false
		}
	}

	// If no popular country is found, go for "others"
	if noCountrySelected {
		vec[len(genres)+len(countries)] = 1
	}

	t, err := time.Parse("2006-01-02", rawReview.Released)
	if err != nil {
		t = time.Now()
	}

	// Release date
	vec[len(genres)+len(countries)+1] = float64(t.Year()) / yearNormalization

	// Duration normalization
	vec[len(genres)+len(countries)+2] = float64(rawReview.Duration) / maxDuration

	return vec
}

// CosineSimilarity calculates a similarity.
func CosineSimilarity(a, b dto.SimilarityVector) float64 {
	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// OrderReviewsBySimilarity reorganizes provided slice by similarity.
func OrderReviewsBySimilarity(
	needle dto.SimilarityVector,
	haystack []dto.Review,
	ignoreContentId *string,
) []ScoredReview {
	scored := make([]ScoredReview, 0, len(haystack))
	for _, r := range haystack {
		// Ignore the requested one
		if ignoreContentId != nil && *ignoreContentId == r.ContentId {
			continue
		}

		similarity := cosineSimilarityWeight*CosineSimilarity(
			needle,
			r.SimilarityVector,
		) + scoreWeight*float64(
			r.Score,
		)/normalization

		scored = append(scored, ScoredReview{
			Review:     r,
			Similarity: similarity,
		})
	}

	// Sort by similarity
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Similarity > scored[j].Similarity
	})

	return scored
}

// BuildUserSimilarityVector Build user's preference profile by averaging
// similarity vectors of top-rated reviews.
func BuildUserSimilarityVector(reviews []dto.Review) dto.SimilarityVector {
	profile := make(dto.SimilarityVector, len(reviews[0].SimilarityVector))
	for _, review := range reviews {
		for i, value := range review.SimilarityVector {
			profile[i] += value
		}
	}

	for i := range profile {
		profile[i] /= float64(len(reviews))
	}

	return profile
}
