package api

import (
	"context"
	"time"

	"github.com/course-go/reelgoofy/internal/db"
	"github.com/google/uuid"
)

// IngestReviews
// ---------------------------------------------------------
// POST /reviews
// --------------------------------------------------------.
func (api *API) IngestReviews( //nolint:ireturn
	ctx context.Context,
	request IngestReviewsRequestObject,
) (IngestReviewsResponseObject, error) {
	errorData := api.validateIngestRequest(request.Body.Data)
	if len(errorData) != 0 {
		return IngestReviews400JSONResponse{
			Status: Fail,
			Data:   errorData,
		}, nil
	}

	api.Logger.Info("Ingesting reviews", "review count", len(*request.Body.Data.Reviews))

	reviews := []Review{}
	for _, review := range *request.Body.Data.Reviews {
		userId, _ := uuid.Parse(review.UserId)
		contentId, _ := uuid.Parse(review.ContentId)

		reviewId := uuid.New()
		api.Database[userId] = append(api.Database[userId], db.Review{
			ReviewId:  reviewId,
			Title:     review.Title,
			ContentId: contentId,
			Genres:    review.Genres,
			Score:     review.Score,
		})
		reviews = append(reviews, convertReview(review, reviewId.String()))
	}

	api.Logger.Info("Review ingestion completed successfully", "ingestedCount", len(reviews))

	return IngestReviews201JSONResponse{
		Status: Success,
		Data:   Reviews{Reviews: &reviews},
	}, nil
}

func (api *API) validateIngestRequest(data *RawReviews) (errorData map[string]any) {
	errorData = make(map[string]any)
	if data == nil || data.Reviews == nil {
		errorData["html body"] = "no reviews in html body provided"
		return errorData
	}

	for _, review := range *data.Reviews {
		contentId, errContent := uuid.Parse(review.ContentId)
		if errContent != nil {
			errorData["contentId"] = "ID is not a valid UUID."
		}

		userId, err := uuid.Parse(review.UserId)
		if err != nil {
			errorData["userId"] = "ID is not a valid UUID."
		} else if errContent == nil { // if it is valid user id and valid content id
			for _, usrReview := range api.Database[userId] {
				if usrReview.ContentId == contentId {
					errorData["1 content 2 reviews"] = "User can't review one content multiple times."
				}
			}
		}
		if review.Released != nil {
			_, err = time.Parse(time.DateOnly, *review.Released)
			if err != nil {
				errorData["released"] = "Invalid date formats."
			}
		}

		if review.Score < 0 {
			errorData["score"] = "Score must be positive integer."
		}
	}
	return errorData
}

func convertReview(raw RawReview, id string) Review {
	return Review{
		Id:          id,
		ContentId:   raw.ContentId,
		Review:      raw.Review,
		Score:       raw.Score,
		UserId:      raw.UserId,
		Actors:      raw.Actors,
		Description: raw.Description,
		Director:    raw.Director,
		Duration:    raw.Duration,
		Genres:      raw.Genres,
		Origins:     raw.Origins,
		Released:    raw.Released,
		Tags:        raw.Tags,
		Title:       raw.Title,
	}
}

// DeleteReview
// ---------------------------------------------------------
// DELETE /reviews/{reviewId}
// --------------------------------------------------------.
func (api *API) DeleteReview( //nolint:ireturn
	ctx context.Context,
	request DeleteReviewRequestObject,
) (DeleteReviewResponseObject, error) {
	api.Logger.Info("Processing delete review request", "reviewId", request.ReviewId)

	id, err := uuid.Parse(request.ReviewId)
	if err != nil {
		return DeleteReview400JSONResponse{ //nolint:nilerr
			Status: Fail,
			Data: map[string]any{
				"reviewId": "ID is not a valid UUID.",
			},
		}, nil
	}

	hasDeleted := api.deleteReviewFromDb(id)
	if !hasDeleted {
		return DeleteReview404JSONResponse{
			Status: Fail,
			Data: map[string]any{
				"reviewId": "Review with such ID not found.",
			},
		}, nil
	}

	api.Logger.Info("Review deleted successfully", "reviewId", id)

	return DeleteReview200JSONResponse{
		Status: Success,
		Data:   Reviews{},
	}, nil
}

func (api *API) deleteReviewFromDb(reviewId uuid.UUID) bool {
	for contentId, reviews := range api.Database {
		for i, r := range reviews {
			if r.ReviewId == reviewId {
				reviews[i] = reviews[len(reviews)-1]
				api.Database[contentId] = reviews[:len(reviews)-1]
				return true
			}
		}
	}
	return false
}
