package service

import (
	"cmp"
	"slices"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/course-go/reelgoofy/internal/repository"
)

// settings for recommendation algorithm.
const (
	originRecommendationValue   = 1
	tagRecommendationValue      = 2
	actorRecommendationValue    = 3
	genreRecommendationValue    = 4
	directorRecommendationValue = 5
	maxReviewScore              = 100.0
)

type RecommendationService struct {
	repo *repository.ReviewRepository
}

func NewRecommendationService(repo *repository.ReviewRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) RecommendFromContent(
	contentId string,
	offset, limit int,
) ([]models.Recommendation, error) {
	allReviews := s.repo.GetAllReviews()

	separatedReviews := separateReviewsByContentId(allReviews)

	if len(separatedReviews[contentId]) == 0 {
		return []models.Recommendation{}, models.ErrNotFound
	}

	sourceReviewParams := aggregateReviews(separatedReviews[contentId])

	aggregatedReviewsParams := make([]models.RecommedationParameters, 0)
	for id, reviews := range separatedReviews {
		if id == contentId {
			continue
		}
		aggregatedReviewParams := aggregateReviews(reviews)
		aggregatedReviewParams.SimilarityScore = calculateSimilarityScore(sourceReviewParams, aggregatedReviewParams)
		aggregatedReviewsParams = append(aggregatedReviewsParams, aggregatedReviewParams)
	}

	slices.SortFunc(aggregatedReviewsParams, sortAggregatedReviews)

	recommendations := make([]models.Recommendation, len(aggregatedReviewsParams))

	for index, reviewParams := range aggregatedReviewsParams {
		recommendations[index] = models.Recommendation{
			ContentId: reviewParams.ContentId,
			Title:     reviewParams.Title,
		}
	}

	return paginate(recommendations, offset, limit), nil
}

func (s *RecommendationService) RecommendFromUser(userId string, offset, limit int) ([]models.Recommendation, error) {
	allReviews := s.repo.GetAllReviews()

	userReviews := getReviewsByUserId(allReviews, userId)

	if len(userReviews) == 0 {
		return []models.Recommendation{}, models.ErrNotFound
	}

	separatedReviews := separateReviewsByContentId(allReviews)

	userRecommendationParams := aggregateUserReviews(userReviews)

	// remove content user already reviewed
	for _, review := range userReviews {
		delete(separatedReviews, review.ContentID)
	}

	aggregatedReviewsParams := make([]models.RecommedationParameters, 0)
	for _, reviews := range separatedReviews {
		aggregatedReviewParams := aggregateReviews(reviews)
		aggregatedReviewParams.SimilarityScore = calculateSimilarityScoreForUser(
			userRecommendationParams,
			aggregatedReviewParams,
		)
		aggregatedReviewsParams = append(aggregatedReviewsParams, aggregatedReviewParams)
	}

	slices.SortFunc(aggregatedReviewsParams, sortAggregatedReviews)

	recommendations := make([]models.Recommendation, len(aggregatedReviewsParams))

	for index, reviewParams := range aggregatedReviewsParams {
		recommendations[index] = models.Recommendation{
			ContentId: reviewParams.ContentId,
			Title:     reviewParams.Title,
		}
	}

	return paginate(recommendations, offset, limit), nil
}

func separateReviewsByContentId(reviews []models.RawReview) map[string][]models.RawReview {
	separatedReviews := make(map[string][]models.RawReview)
	for _, review := range reviews {
		separatedReviews[review.ContentID] = append(separatedReviews[review.ContentID], review)
	}
	return separatedReviews
}

func getReviewsByUserId(reviews []models.RawReview, userId string) []models.RawReview {
	var separatedReviews []models.RawReview
	for _, review := range reviews {
		if review.UserID == userId {
			separatedReviews = append(separatedReviews, review)
		}
	}
	return separatedReviews
}

func calculateSimilarityScore(source, candidate models.RecommedationParameters) float64 {
	var score float64
	for _, genre := range candidate.Genres {
		if slices.Contains(source.Genres, genre) {
			score += genreRecommendationValue
		}
	}
	for _, tag := range candidate.Tags {
		if slices.Contains(source.Tags, tag) {
			score += tagRecommendationValue
		}
	}
	for _, actor := range candidate.Actors {
		if slices.Contains(source.Actors, actor) {
			score += actorRecommendationValue
		}
	}
	for _, origin := range candidate.Origins {
		if slices.Contains(source.Origins, origin) {
			score += originRecommendationValue
		}
	}
	for _, director := range candidate.Directors {
		if slices.Contains(source.Directors, director) {
			score += directorRecommendationValue
		}
	}
	return score
}

func calculateSimilarityScoreForUser(
	userSource models.RecommedationParametersForUser,
	candidate models.RecommedationParameters,
) float64 {
	var score float64
	for _, genre := range candidate.Genres {
		_, exists := userSource.GenresWithAvgScore[genre]
		if exists {
			score += genreRecommendationValue * (userSource.GenresWithAvgScore[genre] / maxReviewScore)
		}
	}
	for _, tag := range candidate.Tags {
		_, exists := userSource.TagsWithAvgScore[tag]
		if exists {
			score += tagRecommendationValue * (userSource.TagsWithAvgScore[tag] / maxReviewScore)
		}
	}
	for _, actor := range candidate.Actors {
		_, exists := userSource.ActorsWithAvgScore[actor]
		if exists {
			score += actorRecommendationValue * (userSource.ActorsWithAvgScore[actor] / maxReviewScore)
		}
	}
	for _, origin := range candidate.Origins {
		_, exists := userSource.OriginsWithAvgScore[origin]
		if exists {
			score += originRecommendationValue * (userSource.OriginsWithAvgScore[origin] / maxReviewScore)
		}
	}
	for _, director := range candidate.Directors {
		_, exists := userSource.DirectorsWithAvgScore[director]
		if exists {
			score += directorRecommendationValue * (userSource.DirectorsWithAvgScore[director] / maxReviewScore)
		}
	}
	return score
}

func aggregateReviews(reviews []models.RawReview) models.RecommedationParameters {
	var aggregatedReviews models.RecommedationParameters

	uniqueGenres := make(map[string]bool)
	uniqueTags := make(map[string]bool)
	uniqueActors := make(map[string]bool)
	uniqueOrigins := make(map[string]bool)
	uniqueDirectors := make(map[string]bool)
	var sumScore int

	// get unique parts
	for _, review := range reviews {
		for _, genre := range review.Genres {
			if genre != "" {
				uniqueGenres[genre] = true
			}
		}
		for _, tag := range review.Tags {
			if tag != "" {
				uniqueTags[tag] = true
			}
		}
		for _, actor := range review.Actors {
			if actor != "" {
				uniqueActors[actor] = true
			}
		}
		for _, origin := range review.Origins {
			if origin != "" {
				uniqueOrigins[origin] = true
			}
		}

		if review.Director != "" {
			uniqueDirectors[review.Director] = true
		}

		sumScore += review.Score
		aggregatedReviews.ContentId = review.ContentID
		aggregatedReviews.Title = review.Title
	}

	// fill RecommedationParameters struct
	for uniqueGenre := range uniqueGenres {
		aggregatedReviews.Genres = append(aggregatedReviews.Genres, uniqueGenre)
	}
	for uniqueTag := range uniqueTags {
		aggregatedReviews.Tags = append(aggregatedReviews.Tags, uniqueTag)
	}
	for uniqueActor := range uniqueActors {
		aggregatedReviews.Actors = append(aggregatedReviews.Actors, uniqueActor)
	}
	for uniqueOrigin := range uniqueOrigins {
		aggregatedReviews.Origins = append(aggregatedReviews.Origins, uniqueOrigin)
	}
	for uniqueDirector := range uniqueDirectors {
		aggregatedReviews.Directors = append(aggregatedReviews.Directors, uniqueDirector)
	}
	aggregatedReviews.AvgScore = float64(sumScore) / float64(len(reviews))

	return aggregatedReviews
}

func aggregateUserReviews(reviews []models.RawReview) models.RecommedationParametersForUser {
	aggregatedReviews := models.RecommedationParametersForUser{
		GenresWithAvgScore:    make(map[string]float64),
		TagsWithAvgScore:      make(map[string]float64),
		ActorsWithAvgScore:    make(map[string]float64),
		OriginsWithAvgScore:   make(map[string]float64),
		DirectorsWithAvgScore: make(map[string]float64),
	}
	uniqueGenresWithScores := make(map[string][]int)
	uniqueTagsWithScores := make(map[string][]int)
	uniqueActorsWithScores := make(map[string][]int)
	uniqueOriginsWithScores := make(map[string][]int)
	uniqueDirectorsWithScores := make(map[string][]int)

	for _, review := range reviews {
		for _, genre := range review.Genres {
			if genre != "" {
				uniqueGenresWithScores[genre] = append(uniqueGenresWithScores[genre], review.Score)
			}
		}
		for _, tag := range review.Tags {
			if tag != "" {
				uniqueTagsWithScores[tag] = append(uniqueTagsWithScores[tag], review.Score)
			}
		}
		for _, actor := range review.Actors {
			if actor != "" {
				uniqueActorsWithScores[actor] = append(uniqueActorsWithScores[actor], review.Score)
			}
		}
		for _, origin := range review.Origins {
			if origin != "" {
				uniqueOriginsWithScores[origin] = append(uniqueOriginsWithScores[origin], review.Score)
			}
		}

		if review.Director != "" {
			uniqueDirectorsWithScores[review.Director] = append(
				uniqueDirectorsWithScores[review.Director], review.Score,
			)
		}
	}
	for genre, scores := range uniqueGenresWithScores {
		aggregatedReviews.GenresWithAvgScore[genre] = avg(scores)
	}
	for tag, scores := range uniqueTagsWithScores {
		aggregatedReviews.TagsWithAvgScore[tag] = avg(scores)
	}
	for actor, scores := range uniqueActorsWithScores {
		aggregatedReviews.ActorsWithAvgScore[actor] = avg(scores)
	}
	for origin, scores := range uniqueOriginsWithScores {
		aggregatedReviews.OriginsWithAvgScore[origin] = avg(scores)
	}
	for director, scores := range uniqueDirectorsWithScores {
		aggregatedReviews.DirectorsWithAvgScore[director] = avg(scores)
	}
	return aggregatedReviews
}

func avg(slice []int) float64 {
	var total int
	for _, value := range slice {
		total += value
	}
	return float64(total) / float64(len(slice))
}

func sortAggregatedReviews(i, j models.RecommedationParameters) int {
	finalScoreI := i.AvgScore + i.SimilarityScore
	finalScoreJ := j.AvgScore + j.SimilarityScore

	diff := cmp.Compare(finalScoreJ, finalScoreI)
	if diff == 0 {
		return cmp.Compare(j.SimilarityScore, i.SimilarityScore)
	}

	return diff
}
