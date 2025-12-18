package api

import (
	"context"
	"slices"

	"github.com/google/uuid"
)

func sliceWithLimitAndOffset(recs []Recommendation, limit *int, offset *int) []Recommendation {
	o := 0
	if offset != nil {
		o = *offset
		if o >= len(recs) {
			return []Recommendation{}
		}
	}
	if limit == nil {
		return recs[o:]
	}
	return recs[o:min(o+*limit, len(recs))]
}

func isValidLimitAndOffset(limit *int, offset *int) bool {
	if (limit != nil && *limit < 0) || (offset != nil && *offset < 0) {
		return false
	}
	return true
}

// RecommendContentToContent
// ---------------------------------------------------------
// GET /recommendations/content/{contentId}/content
// --------------------------------------------------------.
func (api *API) RecommendContentToContent( //nolint:ireturn
	ctx context.Context, request RecommendContentToContentRequestObject,
) (RecommendContentToContentResponseObject, error) {
	limit := request.Params.Limit
	offset := request.Params.Offset
	api.Logger.Info("getting recs for content", "ContentId", request.ContentId,
		"Params.Limit", limit, "Params.Offset", offset)

	if !isValidLimitAndOffset(limit, offset) {
		return RecommendContentToContent400JSONResponse{
			Status: Fail,
			Data: map[string]any{
				"Params": "Limit and Offset can not be negative",
			},
		}, nil
	}

	contentId, err := uuid.Parse(request.ContentId)
	if err != nil {
		api.Logger.Error("invalid content id provided", "ContentId", request.ContentId)
		return RecommendContentToContent400JSONResponse{ //nolint:nilerr
			Status: Fail,
			Data: map[string]any{
				"contentID": "ID is not a valid UUID.",
			},
		}, nil
	}

	recs := api.getContentByContent(contentId, limit, offset)

	api.Logger.Info("Content recommendation search finished", "ContentId", contentId, "TotalFound", len(recs))

	if len(recs) == 0 {
		return RecommendContentToContent404JSONResponse{
			Status: Fail,
			Data: map[string]any{
				"contentId": "Content with such ID not found.",
			},
		}, nil
	}

	return RecommendContentToContent200JSONResponse{
		Status: Success,
		Data: Recommendations{
			Recommendations: &recs,
		},
	}, nil
}

func (api *API) getContentByContent(contentId uuid.UUID, limit *int, offset *int) []Recommendation {
	// I know it is very inefficient to do it this way, although it is the
	// most straightforward for this task.

	// This algorithm suggests content with at least one common genre and an average rating of >= 80.
	type stat struct {
		title      *string
		genres     *[]string
		scoreSum   int
		scoreCount int
	}
	contentScore := make(map[uuid.UUID]*stat)

	wantedGenres := &[]string{}
	for _, reviews := range api.Database {
		// I’m aware that iterating through reviews copies the entire struct, which is inefficient.
		// Would be better to implement db to use pointers or implement read sql db
		for _, r := range reviews {
			if r.ContentId == contentId {
				wantedGenres = r.Genres
			}
			ptrStat, ok := contentScore[r.ContentId]
			if !ok {
				ptrStat = &stat{
					title:      r.Title,
					genres:     r.Genres,
					scoreSum:   0,
					scoreCount: 0,
				}
				contentScore[r.ContentId] = ptrStat
			}
			ptrStat.scoreSum += r.Score
			ptrStat.scoreCount += 1
		}
	}

	api.Logger.Info("Identified source genres", "ContentId", contentId, "Genres", wantedGenres)

	recs := []Recommendation{}
	for id, content := range contentScore {
		if !haveCommonGenres(content.genres, wantedGenres) {
			continue
		}
		avgScore := float64(content.scoreSum) / float64(content.scoreCount)
		if avgScore >= qualityTreshold {
			recs = append(recs, Recommendation{Id: id.String(), Title: content.title})
		}
	}

	api.Logger.Info("Filtering complete", "CandidatesScanned", len(contentScore), "QualifiedRecs", len(recs))

	return sliceWithLimitAndOffset(recs, limit, offset)
}

func haveCommonGenres(genres *[]string, wantedGenres *[]string) bool {
	if wantedGenres == nil || len(*wantedGenres) == 0 {
		return true
	}

	if genres == nil || len(*genres) == 0 {
		return false
	}

	for _, elem1 := range *genres {
		if slices.Contains(*wantedGenres, elem1) {
			return true
		}
	}
	return false
}

// RecommendContentToUser
// ---------------------------------------------------------
// GET /recommendations/users/{userId}/content
// --------------------------------------------------------.
func (api *API) RecommendContentToUser( //nolint:ireturn
	ctx context.Context,
	request RecommendContentToUserRequestObject,
) (RecommendContentToUserResponseObject, error) {
	limit := request.Params.Limit
	offset := request.Params.Offset
	if !isValidLimitAndOffset(limit, offset) {
		return RecommendContentToUser400JSONResponse{
			Status: Fail,
			Data: map[string]any{
				"Params": "Limit and Offset can not be negative",
			},
		}, nil
	}

	api.Logger.Info("getting recs for content", "UserId", request.UserId,
		"Params.Limit", limit, "Params.Offset", offset)
	userId, err := uuid.Parse(request.UserId)
	if err != nil {
		api.Logger.Error("invalid content id provided", "UserId", request.UserId)
		return RecommendContentToUser400JSONResponse{ //nolint:nilerr
			Status: Fail,
			Data: map[string]any{
				"UserId": "ID is not a valid UUID.",
			},
		}, nil
	}

	recs := api.getContentByUser(userId, limit, offset)

	api.Logger.Info("User recommendation search finished", "UserId", userId, "TotalFound", len(recs))

	if len(recs) == 0 {
		return RecommendContentToUser404JSONResponse{
			Status: Fail,
			Data: map[string]any{
				"UserId": "User with such ID not found.",
			},
		}, nil
	}

	return RecommendContentToUser200JSONResponse{
		Status: Success,
		Data: Recommendations{
			Recommendations: &recs,
		},
	}, nil
}

func (api *API) getContentByUser(userId uuid.UUID, limit *int, offset *int) []Recommendation {
	recs := []Recommendation{}
	// Could also be done in a more efficient way
	for _, review := range api.Database[userId] {
		if review.Score >= qualityTreshold {
			recs = append(recs, Recommendation{Id: review.ContentId.String(), Title: review.Title})
		}
	}

	api.Logger.Info("Processed user reviews", "UserId", userId,
		"TotalReviews", len(api.Database[userId]), "QualifiedRecs", len(recs))

	return sliceWithLimitAndOffset(recs, limit, offset)
}
