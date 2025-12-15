package internal

import (
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("not found")

const (
	optimalGenres = 3
	optimalTags   = 5
)

type Storage struct {
	mut     sync.RWMutex
	reviews map[string]Review // reviewID -> Review

	contentIdx map[string]map[string]struct{} // contentID -> map of reviews (revId->Rev)
	userIdx    map[string]map[string]struct{} // userID -> map of reviews (revId->Rev)
}

func NewStorage() *Storage {
	return &Storage{
		reviews:    make(map[string]Review),
		contentIdx: make(map[string]map[string]struct{}),
		userIdx:    make(map[string]map[string]struct{}),
	}
}

func (s *Storage) AddReview(review Review) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.reviews[review.ID] = review

	_, ok := s.contentIdx[review.ContentID]
	if !ok {
		s.contentIdx[review.ContentID] = make(map[string]struct{})
	}
	s.contentIdx[review.ContentID][review.ID] = struct{}{}

	_, ok = s.userIdx[review.UserID]
	if !ok {
		s.userIdx[review.UserID] = make(map[string]struct{})
	}
	s.userIdx[review.UserID][review.ID] = struct{}{}
}

func (s *Storage) DeleteReview(reviewId string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	rew, ok := s.reviews[reviewId]
	if !ok {
		return ErrNotFound
	}

	delete(s.reviews, reviewId)

	// delete that review from content reviews
	contentMap, ok := s.contentIdx[rew.ContentID]
	if ok {
		delete(contentMap, reviewId)
		if len(contentMap) == 0 {
			delete(s.contentIdx, rew.ContentID)
		}
	}

	// delete that review from users reviews
	userMap, ok := s.userIdx[rew.UserID]
	if ok {
		delete(userMap, reviewId)
		if len(userMap) == 0 {
			delete(s.userIdx, rew.UserID)
		}
	}
	return nil
}

func (s *Storage) HasContent(id string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	_, ok := s.contentIdx[id]
	return ok
}

func (s *Storage) HasUser(id string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	_, ok := s.userIdx[id]
	return ok
}

func (s *Storage) GetReviewByUser(userId string) []Review {
	s.mut.RLock()
	defer s.mut.RUnlock()
	reviews := make([]Review, 0)
	if !s.HasUser(userId) {
		return reviews
	}

	for id := range s.userIdx[userId] {
		r, ok := s.reviews[id]
		if ok {
			reviews = append(reviews, r)
		}
	}
	return reviews
}

func (s *Storage) GetReviewsByContent(contentId string) []Review {
	s.mut.RLock()
	defer s.mut.RUnlock()
	reviews := make([]Review, 0)
	if !s.HasContent(contentId) {
		return reviews
	}

	for id := range s.contentIdx[contentId] {
		r, ok := s.reviews[id]
		if ok {
			reviews = append(reviews, r)
		}
	}
	return reviews
}

func (s *Storage) GetAllContentIDs() []string {
	s.mut.RLock()
	defer s.mut.RUnlock()
	contentIds := make([]string, 0, len(s.contentIdx))
	for id := range s.contentIdx {
		contentIds = append(contentIds, id)
	}
	return contentIds
}

func (s *Storage) GetReviewForContent(contentId string) (SynthReview, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if !s.HasContent(contentId) {
		return SynthReview{}, ErrNotFound
	}

	reviews, ok := s.contentIdx[contentId]
	if !ok || len(reviews) == 0 {
		return SynthReview{}, ErrNotFound
	}
	genreFreq := make(map[string]int)
	tagFreq := make(map[string]int)

	title := ""

	for reviewID := range reviews {
		r, exists := s.reviews[reviewID]
		if !exists {
			continue
		}
		if title == "" {
			title += r.Title
		}

		for _, g := range r.Genres {
			genreFreq[g]++
		}

		for _, t := range r.Tags {
			tagFreq[t]++
		}
	}
	// Select top 3 genres
	topGenres := topN(genreFreq, optimalGenres)

	// Select top 5 tags
	topTags := topN(tagFreq, optimalTags)

	synthRev := SynthReview{
		ContentId: contentId,
		Title:     title,
		Genres:    topGenres,
		Tags:      topTags,
	}
	return synthRev, nil
}

func topN(freq map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	arr := make([]kv, 0, len(freq))
	for k, v := range freq {
		arr = append(arr, kv{k, v})
	}

	// sort by frequency descending
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Value > arr[j].Value
	})

	// pick N items
	result := make([]string, 0, n)
	for i := 0; i < len(arr) && i < n; i++ {
		result = append(result, arr[i].Key)
	}
	return result
}
