package recommend

import (
	"sort"

	"github.com/course-go/reelgoofy/internal/model"
	"github.com/course-go/reelgoofy/internal/repository"
)

type Engine struct {
	S *repository.InMemoryRepository
}

func NewEngine(s *repository.InMemoryRepository) *Engine { return &Engine{S: s} }

// RecommendContentToContent returns other content ranked purely by average score (highest first).
func (e *Engine) RecommendContentToContent(contentID string, limit, offset int) []model.Recommendation {
	if len(e.S.ReviewsForContent(contentID)) == 0 {
		return nil
	}
	type scored struct {
		id, title string
		avg       int
	}
	ids := e.S.AllContentIDs()
	list := make([]scored, 0, len(ids))
	for _, otherID := range ids {
		if otherID == contentID {
			continue
		}
		revs := e.S.ReviewsForContent(otherID)
		if len(revs) == 0 {
			continue
		}
		total := 0
		for _, r := range revs {
			total += r.Score
		}
		list = append(list, scored{id: otherID, title: revs[0].Title, avg: total / len(revs)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].avg > list[j].avg })
	if offset >= len(list) {
		return []model.Recommendation{}
	}
	if limit <= 0 {
		limit = 20
	}
	end := offset + limit
	end = min(end, len(list))
	recs := make([]model.Recommendation, 0, end-offset)
	for _, sc := range list[offset:end] {
		recs = append(recs, model.Recommendation{ID: sc.id, Title: sc.title})
	}
	return recs
}

// RecommendContentToUser recommends unseen content for a user ranked by average score only.
func (e *Engine) RecommendContentToUser(userID string, limit, offset int) []model.Recommendation {
	userRevs := e.S.ReviewsForUser(userID)
	if len(userRevs) == 0 {
		return nil
	}
	reviewed := make(map[string]struct{}, len(userRevs))
	for _, r := range userRevs {
		reviewed[r.ContentID] = struct{}{}
	}
	type scored struct {
		id, title string
		avg       int
	}
	ids := e.S.AllContentIDs()
	list := make([]scored, 0, len(ids))
	for _, cid := range ids {
		if _, seen := reviewed[cid]; seen {
			continue
		}
		revs := e.S.ReviewsForContent(cid)
		if len(revs) == 0 {
			continue
		}
		total := 0
		for _, r := range revs {
			total += r.Score
		}
		list = append(list, scored{id: cid, title: revs[0].Title, avg: total / len(revs)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].avg > list[j].avg })
	if offset >= len(list) {
		return []model.Recommendation{}
	}
	if limit <= 0 {
		limit = 20
	}
	end := offset + limit
	end = min(end, len(list))
	recs := make([]model.Recommendation, 0, end-offset)
	for _, sc := range list[offset:end] {
		recs = append(recs, model.Recommendation{ID: sc.id, Title: sc.title})
	}
	return recs
}
