package service

import (
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/repository"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
)

type RecommendationService struct {
	reviewTable *repository.ReviewTable
}

func NewRecommendationService(r *repository.ReviewTable) *RecommendationService {
	return &RecommendationService{reviewTable: r}
}

func (s *RecommendationService) RecommendContentToContent(
	contentId string,
	limit, offset int,
) (structures.Recommendations, map[string]any) {
	all := s.reviewTable.GetAll()
	errs := make(map[string]any)

	genresMap := make(map[string]struct{})
	foundContent := false

	for _, r := range all {
		if r.ContentID == contentId {
			foundContent = true
			for _, g := range r.Genres {
				genresMap[g] = struct{}{}
			}
			break
		}
	}

	if !foundContent {
		errs["contentId"] = "Content with such ID not found."
		return structures.Recommendations{}, errs
	}

	recs := []structures.Recommendation{}
	seen := make(map[string]struct{})

	for _, r := range all {
		if r.ContentID == contentId || (len(genresMap) == 0 && r.Score < 80) {
			continue
		}
		if len(genresMap) == 0 && r.Score >= 80 { // if no genres, recommend highly rated content
			_, ok := seen[r.ContentID]
			if !ok {
				recs = append(recs, structures.Recommendation{ID: r.ID, Title: r.Title})
				seen[r.ContentID] = struct{}{}
			}
			continue
		}

		for _, g := range r.Genres { // recommend if shares any genre
			_, ok := genresMap[g]
			if ok {
				_, ok := seen[r.ContentID]
				if !ok {
					recs = append(recs, structures.Recommendation{ID: r.ID, Title: r.Title})
					seen[r.ContentID] = struct{}{}
				}
				break
			}
		}
	}

	if offset >= len(recs) {
		return structures.Recommendations{Recommendations: []structures.Recommendation{}}, nil
	}
	end := min(offset+limit, len(recs))
	return structures.Recommendations{Recommendations: recs[offset:end]}, nil
}

func (s *RecommendationService) RecommendContentToUser(
	userId string,
	limit, offset int,
) (structures.Recommendations, map[string]any) {
	all := s.reviewTable.GetAll()
	var foundUser bool
	errs := make(map[string]any)

	averageUsersScore := 0
	count := 0

	for _, r := range all {
		if r.UserID == userId {
			foundUser = true
			averageUsersScore += r.Score
			count++
		}
	}

	if !foundUser {
		errs["userId"] = "User with such ID not found."
		return structures.Recommendations{}, errs
	}

	averageUsersScore /= count
	recs := []structures.Recommendation{}
	seen := make(map[string]struct{})

	for _, r := range all {
		if r.UserID == userId {
			continue
		}
		if r.Score >= averageUsersScore { // recommend content with score higher than user's average
			_, ok := seen[r.ContentID]
			if !ok {
				recs = append(recs, structures.Recommendation{ID: r.ID, Title: r.Title})
				seen[r.ContentID] = struct{}{}
			}
			continue
		}
	}

	if offset >= len(recs) {
		return structures.Recommendations{Recommendations: []structures.Recommendation{}}, nil
	}
	end := min(offset+limit, len(recs))
	return structures.Recommendations{Recommendations: recs[offset:end]}, nil
}
