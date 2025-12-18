package service

import (
	"net/http"
	"slices"
	"sort"

	"github.com/course-go/reelgoofy/internal/recommendation"
	"github.com/course-go/reelgoofy/internal/rest"
	"github.com/course-go/reelgoofy/internal/review"
	"github.com/google/uuid"
)

const (
	scoreThreshold         = 70
	genreRelevanceScore    = 2
	actorRelevanceScore    = 1
	directorRelevanceScore = 3
)

type RecommendationService struct {
	database *review.Database
}

func CreateRecommendationService(database *review.Database) *RecommendationService {
	return &RecommendationService{database}
}

func (service *RecommendationService) RecommendByContent(contentId string, offset int, limit int) (rest.Response, int) {
	validations, ok := checkPathVariablesValidity(contentId, "contentId", offset, limit)
	if !ok {
		return rest.Response{Status: rest.StatusFail, Data: validations}, http.StatusBadRequest
	}
	allFilms := service.database.GetAllFilms()
	targetFilm := allFilms[contentId]
	var scoredRecommendations []recommendation.ScoredRecommendation
	for _, film := range allFilms {
		score := calculateFilmSimilarity(targetFilm, film)
		if score > 0 && film.Id != contentId {
			recommend := recommendation.ScoredRecommendation{
				Recommendation: recommendation.Recommendation{Id: film.Id, Title: film.Title},
				Score:          score,
			}
			scoredRecommendations = append(scoredRecommendations, recommend)
		}
	}
	sort.Slice(scoredRecommendations, func(i, j int) bool {
		return scoredRecommendations[i].Score > scoredRecommendations[j].Score
	})
	finalRecommends := make([]recommendation.Recommendation, 0, limit)
	for _, score := range scoredRecommendations {
		finalRecommends = append(finalRecommends, score.Recommendation)
	}
	return rest.SuccessRecommendationResponse(applyOffsetAndLimit(offset, limit, finalRecommends)), http.StatusOK
}

func (service *RecommendationService) RecommendByUser(userId string, offset int, limit int) (rest.Response, int) {
	validations, ok := checkPathVariablesValidity(userId, "userId", offset, limit)
	if !ok {
		return rest.Response{Status: rest.StatusFail, Data: validations}, http.StatusBadRequest
	}
	seenFilms := make(map[string]bool)
	var highRatedReviews []review.Review
	for _, rvw := range service.database.GetReviewsByUser(userId) {
		seenFilms[rvw.ContentId] = true
		if rvw.Score > scoreThreshold {
			highRatedReviews = append(highRatedReviews, rvw)
		}
	}
	allFilms := service.database.GetAllFilms()
	userProfile := buildUserProfile(highRatedReviews, allFilms)
	var recommendScores []recommendation.ScoredRecommendation
	for _, film := range allFilms {
		if seenFilms[film.Id] {
			continue
		}
		score := calculateFilmRelevance(film, userProfile)
		if score > 0 {
			recommendScore := recommendation.ScoredRecommendation{
				Recommendation: recommendation.Recommendation{Id: film.Id, Title: film.Title},
				Score:          score,
			}
			recommendScores = append(recommendScores, recommendScore)
		}
	}

	sort.Slice(recommendScores, func(i, j int) bool {
		return recommendScores[i].Score > recommendScores[j].Score
	})

	finalRecommends := make([]recommendation.Recommendation, 0, limit)
	for _, score := range recommendScores {
		finalRecommends = append(finalRecommends, score.Recommendation)
	}
	return rest.SuccessRecommendationResponse(applyOffsetAndLimit(offset, limit, finalRecommends)), http.StatusOK
}

func checkPathVariablesValidity(id string, idName string, offset int, limit int) (map[string]string, bool) {
	errors := make(map[string]string)

	_, err := uuid.Parse(id)
	if err != nil {
		errors[idName] = rest.InvalidIDMessage
	}

	if offset < 0 {
		errors["offset"] = rest.InvalidOffsetMessage
	}

	if limit < 0 {
		errors["limit"] = rest.InvalidLimitMessage
	}

	if len(errors) > 0 {
		return errors, false
	}

	return nil, true
}

func calculateFilmSimilarity(targetFilm recommendation.Film, testedFilm recommendation.Film) int {
	var score int
	for _, genre := range testedFilm.Genres {
		if slices.Contains(targetFilm.Genres, genre) {
			score += genreRelevanceScore
		}
	}
	for _, actor := range testedFilm.Actors {
		if slices.Contains(targetFilm.Actors, actor) {
			score += actorRelevanceScore
		}
	}
	if testedFilm.Director == targetFilm.Director {
		score += directorRelevanceScore
	}
	return score
}

func applyOffsetAndLimit(
	offset int,
	limit int,
	recommends []recommendation.Recommendation,
) []recommendation.Recommendation {
	if offset >= len(recommends) {
		return []recommendation.Recommendation{}
	}
	offsetRecommends := recommends[offset:]
	if limit > len(offsetRecommends) {
		return offsetRecommends
	}
	return offsetRecommends[:limit]
}

func buildUserProfile(
	highRatedReviews []review.Review,
	allFilms map[string]recommendation.Film,
) recommendation.UserProfile {
	profile := recommendation.UserProfile{
		FavoriteGenres:    make(map[string]int),
		FavoriteActors:    make(map[string]int),
		FavoriteDirectors: make(map[string]int),
	}
	for _, rvw := range highRatedReviews {
		if film, ok := allFilms[rvw.ContentId]; ok {
			for _, genre := range film.Genres {
				profile.FavoriteGenres[genre]++
			}
			for _, actor := range film.Actors {
				profile.FavoriteActors[actor]++
			}
			profile.FavoriteDirectors[film.Director]++
		}
	}
	return profile
}

func calculateFilmRelevance(film recommendation.Film, profile recommendation.UserProfile) int {
	var score int

	for _, genre := range film.Genres {
		if count, ok := profile.FavoriteGenres[genre]; ok {
			score += genreRelevanceScore * count
		}
	}
	for _, actor := range film.Actors {
		if count, ok := profile.FavoriteActors[actor]; ok {
			score += actorRelevanceScore * count
		}
	}
	if count, ok := profile.FavoriteDirectors[film.Director]; ok {
		score += directorRelevanceScore * count
	}
	return score
}
