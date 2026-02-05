package recommendation

import (
	"errors"
	"slices"
	"sort"

	"github.com/course-go/reelgoofy/internal/entity"
	reviews "github.com/course-go/reelgoofy/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrContentNotFound  = errors.New("content with given ID does not exist")
	ErrUserHasNoReviews = errors.New("user has no reviews")
)

const TopReviews = 3

type RecommendationService struct {
	repository *reviews.Repository
}

func NewRecommendationService(repository *reviews.Repository) *RecommendationService {
	return &RecommendationService{repository: repository}
}

func (s *RecommendationService) RecommendByContentID(id uuid.UUID) ([]entity.Content, error) {
	cont, err := s.getContentByID(id)
	if err != nil {
		return nil, ErrContentNotFound
	}

	var recommendations []entity.Content
	for _, content := range s.getContentsInfo() {
		if content.ContentID == cont.ContentID {
			continue
		}
		if hasNonEmptyIntersection(cont.Genres, content.Genres) || hasNonEmptyIntersection(cont.Tags, content.Tags) {
			recommendations = append(recommendations, content)
		}
	}
	return recommendations, nil
}

func (s *RecommendationService) RecommendByUserID(userId uuid.UUID) ([]entity.Content, error) {
	revs := s.repository.GetReviews()

	var userReviews []entity.Review
	for _, rev := range revs {
		if rev.UserID == userId {
			userReviews = append(userReviews, rev)
		}
	}
	if len(userReviews) == 0 {
		return []entity.Content{}, ErrUserHasNoReviews
	}

	sort.Slice(userReviews, func(i, j int) bool {
		return userReviews[i].Score > userReviews[j].Score
	})

	topN := min(TopReviews, len(userReviews))
	topReviews := userReviews[:topN]

	var topGenres, topTags []string
	topIDs := make(map[uuid.UUID]struct{})
	for _, rev := range topReviews {
		topGenres = append(topGenres, rev.Genres...)
		topTags = append(topTags, rev.Tags...)
		topIDs[rev.ContentID] = struct{}{}
	}

	var recommendations []entity.Content
	for _, content := range s.getContentsInfo() {
		if _, isTop := topIDs[content.ContentID]; isTop {
			continue
		}
		if hasNonEmptyIntersection(topGenres, content.Genres) || hasNonEmptyIntersection(topTags, content.Tags) {
			recommendations = append(recommendations, content)
		}
	}

	return recommendations, nil
}

func (s *RecommendationService) getContentsInfo() []entity.Content {
	seen := make(map[uuid.UUID]struct{})
	var movies []entity.Content

	for _, rev := range s.repository.GetReviews() {
		if _, ok := seen[rev.ContentID]; !ok {
			seen[rev.ContentID] = struct{}{}
			movies = append(movies, entity.Content{
				ContentID:   rev.ContentID,
				Title:       rev.Title,
				Genres:      rev.Genres,
				Tags:        rev.Tags,
				Description: rev.Description,
				Director:    rev.Director,
				Actors:      rev.Actors,
				Origins:     rev.Origins,
				Duration:    rev.Duration,
				Released:    rev.Released,
			})
		}
	}
	return movies
}

func (s *RecommendationService) getContentByID(id uuid.UUID) (entity.Content, error) {
	for _, content := range s.getContentsInfo() {
		if content.ContentID == id {
			return content, nil
		}
	}
	return entity.Content{}, ErrContentNotFound
}

func hasNonEmptyIntersection(arr1, arr2 []string) bool {
	for _, str := range arr1 {
		if slices.Contains(arr2, str) {
			return true
		}
	}

	return false
}
