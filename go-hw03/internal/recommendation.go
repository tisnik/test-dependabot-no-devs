package internal

import (
	"sort"
	"strings"
)

type scoredMovie struct {
	id    string
	title string
	score int
}

const (
	genrePoint = 2
	tagPoint   = 1
)

func RecommendContentToContent(storage *Storage, contentID string, limit, offset int) []Recommendation {
	if !storage.HasContent(contentID) {
		return nil
	}

	baseReviews, err := storage.GetReviewForContent(contentID)
	if err != nil {
		return nil
	}

	baseGenres := toLowerCase(baseReviews.Genres)
	baseTags := toLowerCase(baseReviews.Tags)

	candidates := make([]scoredMovie, 0)

	for _, targetId := range storage.GetAllContentIDs() {
		if contentID == targetId {
			continue
		}

		targetReview, err := storage.GetReviewForContent(targetId)
		if err != nil {
			continue
		}

		score := 0
		targetGenres := toLowerCase(targetReview.Genres)
		for genre := range baseGenres {
			_, ok := targetGenres[genre]
			if ok {
				score += genrePoint
			}
		}

		targetTags := toLowerCase(targetReview.Tags)
		for tag := range baseTags {
			_, ok := targetTags[tag]
			if ok {
				score += tagPoint
			}
		}

		if score > 0 {
			candidates = append(candidates, scoredMovie{id: targetId, title: targetReview.Title, score: score})
		}
	}

	// limit, offset
	return cutByParams(candidates, limit, offset)
}

func RecommendContentToUser(storage *Storage, userID string, limit, offset int) []Recommendation {
	if !storage.HasUser(userID) {
		return nil
	}

	userReviews := storage.GetReviewByUser(userID)
	seen := make(map[string]struct{})

	userGenres := make(map[string]struct{})
	userTags := make(map[string]struct{})

	for _, review := range userReviews {
		for _, genre := range review.Genres {
			userGenres[strings.ToLower(genre)] = struct{}{}
		}

		for _, tag := range review.Tags {
			userTags[strings.ToLower(tag)] = struct{}{}
		}
		seen[review.ContentID] = struct{}{}
	}

	candidates := make([]scoredMovie, 0)
	for _, targetId := range storage.GetAllContentIDs() {
		_, ok := seen[targetId]
		if ok {
			continue
		}

		review, err := storage.GetReviewForContent(targetId)
		if err != nil {
			continue
		}

		score := 0
		for _, genre := range review.Genres {
			_, ok := userGenres[strings.ToLower(genre)]
			if ok {
				score += genrePoint
			}
		}

		for _, tag := range review.Tags {
			_, ok := userTags[strings.ToLower(tag)]
			if ok {
				score += tagPoint
			}
		}

		if score > 0 {
			candidates = append(candidates, scoredMovie{id: targetId, title: review.Title, score: score})
		}
	}

	// limit, offset
	return cutByParams(candidates, limit, offset)
}

func cutByParams(candidates []scoredMovie, limit, offset int) []Recommendation {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].title < candidates[j].title
		}
		return candidates[i].score > candidates[j].score
	})

	start := min(offset, len(candidates))

	end := min(start+limit, len(candidates))

	result := make([]Recommendation, 0, end-start)
	for _, content := range candidates[start:end] {
		result = append(result, Recommendation{Id: content.id, Title: content.title})
	}
	return result
}

func toLowerCase(properties []string) map[string]struct{} {
	newProperties := make(map[string]struct{}, len(properties))
	for _, prop := range properties {
		newProperties[strings.ToLower(prop)] = struct{}{}
	}
	return newProperties
}
