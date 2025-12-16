package internal

import (
	"math"
	"sort"
	"sync"
)

type Storage struct {
	mu           sync.RWMutex // For thread safety of data storage
	reviews      map[string]Review
	avgScores    map[string]float64
	countReviews map[string]int
}

func NewStorage() *Storage {
	return &Storage{
		reviews:      make(map[string]Review),
		avgScores:    make(map[string]float64),
		countReviews: make(map[string]int),
	}
}

func (s *Storage) AddReview(r Review) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews[r.Id] = r
	if r.Score != nil {
		s.updateAvgScore(r.ContentId, float64(*r.Score), 1)
	}
}

func (s *Storage) DeleteReview(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, exists := s.reviews[id]
	if !exists {
		return
	}

	delete(s.reviews, id)
	if r.Score != nil {
		s.updateAvgScore(r.ContentId, -float64(*r.Score), -1)
	}
}

func (s *Storage) GetReviewById(id string) *Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, exists := s.reviews[id]
	if !exists {
		return nil
	}
	return &r
}

func (s *Storage) GetReviewsByContentId(contentId string) []Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []Review{}
	for _, r := range s.reviews {
		if r.ContentId == contentId {
			result = append(result, r)
		}
	}
	return result
}

func (s *Storage) GetReviewsByUserId(userId string) []Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []Review{}
	for _, r := range s.reviews {
		if r.UserId == userId {
			result = append(result, r)
		}
	}
	return result
}

/* ------- Recommendations algorithms ----- */

// RecommendContentToUser returns top movies, which user has not seen, order by avgScore.
func (s *Storage) RecommendContentToUser(userId string, limit *int, offset *int) []Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := map[string]bool{}
	for _, r := range s.reviews {
		if r.UserId == userId {
			seen[r.ContentId] = true
		}
	}

	contentMap := map[string]Review{}
	for _, r := range s.reviews {
		if !seen[r.ContentId] {
			if _, ok := contentMap[r.ContentId]; !ok {
				contentMap[r.ContentId] = r
			}
		}
	}

	// New score of review will be the avgScore, original score of review is not changed
	candidates := []Review{}
	for cid, r := range contentMap {
		candidates = append(candidates, copyAvgScore(r, s.avgScores[cid]))
	}

	sort.Slice(candidates, func(i, j int) bool {
		return *candidates[i].Score > *candidates[j].Score
	})

	return paginateReviews(candidates, offset, limit)
}

// RecommendContentToContent returns movies with similar genres.
func (s *Storage) RecommendContentToContent(contentId string, limit *int, offset *int) []Review {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var targetGenres map[string]bool
	found := false
	for _, r := range s.reviews {
		if r.ContentId == contentId {
			targetGenres = map[string]bool{}
			if r.Genres != nil {
				for _, g := range *r.Genres {
					targetGenres[g] = true
				}
			}
			found = true
			break
		}
	}
	if !found || len(targetGenres) == 0 {
		return nil
	}

	type candidate struct {
		review     Review
		similarity int
	}

	candidates := []candidate{}
	for _, r := range s.reviews {
		if r.ContentId == contentId {
			continue
		}
		sim := 0
		if r.Genres != nil {
			for _, g := range *r.Genres {
				if targetGenres[g] {
					sim++
				}
			}
		}
		if sim > 0 {
			candidates = append(candidates, candidate{review: r, similarity: sim})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].similarity > candidates[j].similarity
	})

	results := []Review{}
	for _, c := range candidates {
		results = append(results, c.review)
	}

	return paginateReviews(results, offset, limit)
}

/* ------ Helpers ----- */

func (s *Storage) updateAvgScore(contentId string, scoreDelta float64, countDelta int) {
	currentAvg := s.avgScores[contentId]
	currentCount := s.countReviews[contentId]

	newCount := currentCount + countDelta
	if newCount <= 0 {
		s.avgScores[contentId] = 0
		s.countReviews[contentId] = 0
		return
	}
	newAvg := (currentAvg*float64(currentCount) + scoreDelta) / float64(newCount)
	s.avgScores[contentId] = newAvg
	s.countReviews[contentId] = newCount
}

// copyAvgScore sets a review Score to avgScore of review, and to not overwrite original value of review by avgScore.
func copyAvgScore(r Review, avg float64) Review {
	rCopy := r
	rCopy.Score = new(int)
	*rCopy.Score = int(math.Ceil(avg))
	return rCopy
}

func paginateReviews(reviews []Review, offset, limit *int) []Review {
	start := 0
	if offset != nil && *offset >= 0 {
		start = *offset
	}
	if start > len(reviews) {
		start = len(reviews)
	}

	end := len(reviews)
	if limit != nil && *limit > 0 && start+*limit < len(reviews) {
		end = start + *limit
	}
	return reviews[start:end]
}
