package recommendations

import (
	"fmt"

	"github.com/course-go/reelgoofy/internal/reviews"
)

const weight = 100

type RecommendationProfile struct {
	ExcludeIDs         map[string]bool
	PreferredGenres    map[string]int
	PreferredTags      map[string]int
	PreferredDirectors map[string]int
	PreferredActors    map[string]int
}

func newRecommendationProfile() RecommendationProfile {
	return RecommendationProfile{
		ExcludeIDs:         make(map[string]bool),
		PreferredGenres:    make(map[string]int),
		PreferredTags:      make(map[string]int),
		PreferredDirectors: make(map[string]int),
		PreferredActors:    make(map[string]int),
	}
}

func (s *RecommendationService) buildContentProfile(contentID string) (RecommendationProfile, error) {
	source, err := s.repo.GetContent(contentID)
	if err != nil {
		return RecommendationProfile{}, fmt.Errorf("failed to get content: %w", err)
	}

	profile := newRecommendationProfile()
	profile.ExcludeIDs[contentID] = true
	s.addToProfile(&profile, source, weight)

	return profile, nil
}

func (s *RecommendationService) buildUserProfile(userID string) (RecommendationProfile, error) {
	userReviews, err := s.repo.GetReviewsByUserID(userID)
	if err != nil {
		return RecommendationProfile{}, fmt.Errorf("failed to get user reviews: %w", err)
	}

	profile := newRecommendationProfile()

	for _, rev := range userReviews {
		source, err := s.repo.GetContent(rev.ContentID)
		if err != nil {
			continue
		}

		profile.ExcludeIDs[rev.ContentID] = true
		s.addToProfile(&profile, source, rev.Score)
	}

	return profile, nil
}

func (*RecommendationService) addToProfile(p *RecommendationProfile, c reviews.Content, weight int) {
	for _, g := range c.Genres {
		p.PreferredGenres[g] += weight
	}

	for _, t := range c.Tags {
		p.PreferredTags[t] += weight
	}

	if c.Director != "" {
		p.PreferredDirectors[c.Director] += weight
	}

	for _, a := range c.Actors {
		p.PreferredActors[a] += weight
	}
}
