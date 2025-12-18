package service

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/repository"
	"github.com/medvedovan/reelgoofy-hw3/internal/utils"
)

type RecommendationService struct {
	logger     *slog.Logger
	validate   *validator.Validate
	repository *repository.Repository
}

func NewRecommendationService(
	logger *slog.Logger,
	validate *validator.Validate,
	repository *repository.Repository,
) (rs *RecommendationService) {
	return &RecommendationService{
		logger:     logger,
		validate:   validate,
		repository: repository,
	}
}

var ErrContentNotFound = errors.New("content with given UUID not found")

func (rs *RecommendationService) RecommendContentToContent(
	contentId string,
	params model.RecommendContentToContentParams,
) ([]model.Recommendation, error) {
	err := rs.validate.Var(contentId, "uuid")
	if err != nil {
		return nil, fmt.Errorf("invalid contentId: %w", err)
	}

	contentProfile := rs.repository.GetContentProfileById(contentId)

	if contentProfile == nil {
		return nil, ErrContentNotFound
	}

	contentProfiles := rs.repository.GetContentProfiles()
	recommendations := make([]model.Recommendation, 0, len(contentProfiles))

	for cId, cp := range contentProfiles {
		if cId == contentId {
			continue
		}

		if (cp.Director != nil && contentProfile.Director != nil &&
			*cp.Director == *contentProfile.Director) ||
			utils.Contains(contentProfile.Actors, cp.Actors) ||
			utils.Contains(contentProfile.Genres, cp.Genres) ||
			utils.Contains(contentProfile.Tags, cp.Tags) {
			r := model.Recommendation{
				Id:    cId,
				Title: cp.Title,
			}
			recommendations = append(recommendations, r)
		}
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Id < recommendations[j].Id
	})

	paginatedRecommendations := utils.Paginate(recommendations, params.Offset, params.Limit)
	return paginatedRecommendations, nil
}

func (rs *RecommendationService) RecommendContentToUser(
	userId string,
	params model.RecommendContentToUserParams,
) ([]model.Recommendation, error) {
	err := rs.validate.Var(userId, "uuid")
	if err != nil {
		return nil, fmt.Errorf("invalid userId: %w", err)
	}

	userPreferences := rs.repository.GetUserPreferences(userId)

	contentProfiles := rs.repository.GetContentProfiles()
	recommendations := make([]model.Recommendation, 0, len(contentProfiles))

	for cId, cp := range contentProfiles {
		if cp.Score < repository.MinScore {
			continue
		}

		// recommendation for user with no previous reviews
		if userPreferences == nil {
			r := model.Recommendation{
				Id:    cId,
				Title: cp.Title,
			}
			recommendations = append(recommendations, r)
			continue
		}

		if utils.Contains(&[]string{cId}, userPreferences.SeenContentId) {
			continue
		}

		if utils.Contains(&[]string{*cp.Director}, userPreferences.Directors) ||
			utils.Contains(userPreferences.Actors, cp.Actors) ||
			utils.Contains(userPreferences.Genres, cp.Genres) ||
			utils.Contains(userPreferences.Tags, cp.Tags) {
			r := model.Recommendation{
				Id:    cId,
				Title: cp.Title,
			}
			recommendations = append(recommendations, r)
		}
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Id < recommendations[j].Id
	})

	paginatedRecommendations := utils.Paginate(recommendations, params.Offset, params.Limit)
	return paginatedRecommendations, nil
}
