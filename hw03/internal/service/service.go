package service

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/Markek1/reelgoofy/internal/domain"
	"github.com/Markek1/reelgoofy/internal/repository"
	"github.com/google/uuid"
)

const MinScoreThreshold = 50

type Service struct {
	repository *repository.MemoryRepository
	logger     *slog.Logger
}

func NewService(repository *repository.MemoryRepository, logger *slog.Logger) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}

func (s *Service) IngestReviews(rawReviews []domain.RawReview) ([]domain.Review, error) {
	s.logger.Info("Ingesting reviews", "count", len(rawReviews))
	reviews := make([]domain.Review, 0, len(rawReviews))
	for _, rawReview := range rawReviews {
		id := domain.ReviewID(uuid.New().String())

		review := domain.Review{
			ID:        id,
			RawReview: rawReview,
		}
		err := s.repository.Save(review)
		if err != nil {
			s.logger.Error("Failed to save review", "error", err)
			return nil, fmt.Errorf("failed to save review: %w", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (s *Service) DeleteReview(id string) error {
	s.logger.Info("Deleting review", "id", id)
	err := uuid.Validate(id)
	if err != nil {
		return fmt.Errorf("invalid review id: %w", err)
	}
	err = s.repository.Delete(domain.ReviewID(id))
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

// Helper struct for sorting results.
type scoredContent struct {
	ID    domain.ContentID
	Title string
	Score int
}

// RecommendContent recommends content based on genre similarity.
func (s *Service) RecommendContent(contentID domain.ContentID, limit, offset int) ([]domain.Recommendation, error) {
	s.logger.Info("Recommending content", "contentID", contentID, "limit", limit, "offset", offset)
	reviews, err := s.repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	scoredMap := s.recommendContentWithScore(contentID, reviews)

	scoredList := make([]scoredContent, 0, len(scoredMap))
	for _, sc := range scoredMap {
		scoredList = append(scoredList, sc)
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	return s.sliceResults(scoredList, limit, offset), nil
}

func (s *Service) RecommendUser(userID domain.UserID, limit, offset int) ([]domain.Recommendation, error) {
	s.logger.Info("Recommending for user", "userID", userID, "limit", limit, "offset", offset)
	reviews, err := s.repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	// Find contents the user has already watched and liked
	watchedContent := make(map[domain.ContentID]bool)
	for _, r := range reviews {
		if r.UserID == userID && r.Score > MinScoreThreshold {
			watchedContent[r.ContentID] = true
		}
	}

	totalScores := make(map[domain.ContentID]scoredContent)

	for watchedID := range watchedContent {
		recs := s.recommendContentWithScore(watchedID, reviews)

		for id, rec := range recs {
			if watchedContent[id] {
				continue
			}
			if existing, found := totalScores[id]; found {
				existing.Score += rec.Score
				totalScores[id] = existing
			} else {
				totalScores[id] = rec
			}
		}
	}

	scoredList := make([]scoredContent, 0, len(totalScores))
	for _, sc := range totalScores {
		scoredList = append(scoredList, sc)
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	return s.sliceResults(scoredList, limit, offset), nil
}

func (s *Service) sliceResults(scoredList []scoredContent, limit, offset int) []domain.Recommendation {
	total := len(scoredList)

	if offset >= total {
		return []domain.Recommendation{}
	}

	end := total
	if limit > 0 {
		end = min(offset+limit, total)
	}

	sliced := scoredList[offset:end]
	result := make([]domain.Recommendation, len(sliced))
	for i, sc := range sliced {
		result[i] = domain.Recommendation{
			ID:    sc.ID,
			Title: sc.Title,
		}
	}
	return result
}

// Calculates recommendation scores based on genre overlap.
func (s *Service) recommendContentWithScore(
	targetContentID domain.ContentID,
	allReviews []domain.Review,
) map[domain.ContentID]scoredContent {
	targetGenres := make(map[string]bool)
	titles := make(map[domain.ContentID]string)

	for _, r := range allReviews {
		titles[r.ContentID] = r.Title
		if r.ContentID == targetContentID {
			for _, g := range r.Genres {
				targetGenres[g] = true
			}
		}
	}

	scores := make(map[domain.ContentID]int)

	for _, r := range allReviews {
		if r.ContentID == targetContentID {
			continue
		}

		matchCount := 0
		for _, g := range r.Genres {
			if targetGenres[g] {
				matchCount++
			}
		}

		if matchCount > 0 {
			scores[r.ContentID] = matchCount
		}
	}

	results := make(map[domain.ContentID]scoredContent)
	for id, score := range scores {
		results[id] = scoredContent{
			ID:    id,
			Title: titles[id],
			Score: score,
		}
	}
	return results
}
