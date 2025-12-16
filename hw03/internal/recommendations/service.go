package recommendations

import (
	"fmt"
	"sort"

	apperrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/reviews"
)

type RecommendationService struct {
	repo reviews.ReviewRepository
}

type contentScore struct {
	contentID string
	title     string
	score     int
}

func NewRecommendationService(repo reviews.ReviewRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) RecommendContentToContent(
	contentID string,
	limit int,
	offset int,
) (dto.RecommendationsDTO, error) {
	err := validateParams(limit, offset)
	if err != nil {
		return dto.RecommendationsDTO{}, err
	}

	contentRevs, err := s.repo.GetReviewsByContentID(contentID)
	if err != nil {
		return dto.RecommendationsDTO{}, apperrors.NewNotFoundError(
			"contentId",
			"Content with such ID not found.",
		)
	}
	if len(contentRevs) == 0 {
		return dto.RecommendationsDTO{}, apperrors.NewNotFoundError(
			"contentId",
			"Content with such ID doesn't have any reviews.",
		)
	}

	profile, err := s.buildContentProfile(contentID)
	if err != nil {
		return dto.RecommendationsDTO{}, fmt.Errorf("failed to build content profile: %w", err)
	}

	result, err := s.getRecommendations(profile, limit, offset)
	if err != nil {
		return dto.RecommendationsDTO{}, fmt.Errorf("failed to get recommendations: %w", err)
	}

	return dto.RecommendationsDTO{
		Recommendations: result,
	}, nil
}

func (s *RecommendationService) RecommendContentToUser(
	userID string,
	limit int,
	offset int,
) (dto.RecommendationsDTO, error) {
	err := validateParams(limit, offset)
	if err != nil {
		return dto.RecommendationsDTO{}, err
	}

	userRevs, err := s.repo.GetReviewsByUserID(userID)
	if err != nil {
		return dto.RecommendationsDTO{}, apperrors.NewNotFoundError(
			"userId",
			"User with such ID not found.",
		)
	}
	if len(userRevs) == 0 {
		return dto.RecommendationsDTO{}, apperrors.NewNotFoundError(
			"userId",
			"User with such ID doesn't have any reviews.",
		)
	}

	profile, err := s.buildUserProfile(userID)
	if err != nil {
		return dto.RecommendationsDTO{}, fmt.Errorf("failed to build user profile: %w", err)
	}

	result, err := s.getRecommendations(profile, limit, offset)
	if err != nil {
		return dto.RecommendationsDTO{}, fmt.Errorf("failed to get recommendations: %w", err)
	}

	return dto.RecommendationsDTO{
		Recommendations: result,
	}, nil
}

func validateParams(limit int, offset int) error {
	if limit < 0 {
		return apperrors.NewValidationError("limit", "Limit must be non-negative integer")
	}
	if offset < 0 {
		return apperrors.NewValidationError("offset", "Offset must be non-negative integer")
	}
	return nil
}

func (s *RecommendationService) getRecommendations(
	profile RecommendationProfile,
	limit int,
	offset int,
) ([]dto.RecommendationDTO, error) {
	allContents, err := s.repo.GetAllContents()
	if err != nil {
		return []dto.RecommendationDTO{}, fmt.Errorf("failed to get all contents: %w", err)
	}

	scores := make([]contentScore, 0, len(allContents))

	for _, content := range allContents {
		if profile.ExcludeIDs[content.ContentID] {
			continue
		}

		score := s.calculateScore(profile, content)

		if score > 0 {
			scores = append(scores, contentScore{
				contentID: content.ContentID,
				title:     content.Title,
				score:     score,
			})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	paginated := paginate(scores, limit, offset)

	result := make([]dto.RecommendationDTO, 0, len(paginated))
	for _, item := range paginated {
		result = append(result, mapToDTO(item))
	}

	return result, nil
}

func (s *RecommendationService) calculateScore(p RecommendationProfile, c reviews.Content) int {
	score := 0

	for _, g := range c.Genres {
		score += p.PreferredGenres[g]
	}

	for _, t := range c.Tags {
		score += p.PreferredTags[t]
	}

	if c.Director != "" {
		score += p.PreferredDirectors[c.Director]
	}

	for _, a := range c.Actors {
		score += p.PreferredActors[a]
	}

	return score
}

func paginate(items []contentScore, limit, offset int) []contentScore {
	total := len(items)
	if offset > total {
		return []contentScore{}
	}

	end := offset + limit

	return items[offset:min(end, total)]
}

func mapToDTO(item contentScore) dto.RecommendationDTO {
	return dto.RecommendationDTO{
		ID:    item.contentID,
		Title: item.title,
	}
}
