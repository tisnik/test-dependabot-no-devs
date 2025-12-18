package service

import (
	"net/http"
	"time"

	"github.com/course-go/reelgoofy/internal/rest"
	"github.com/course-go/reelgoofy/internal/review"
	"github.com/google/uuid"
)

type ReviewService struct {
	database *review.Database
}

func CreateReviewService(database *review.Database) *ReviewService {
	return &ReviewService{database}
}

func (service *ReviewService) GetAllReviews() []review.Review {
	return service.database.GetAllReviews()
}

func (service *ReviewService) AddReviews(reviewsReq []review.RawReview) (rest.Response, int) {
	resultReviews := make([]review.Review, 0, len(reviewsReq))
	for _, reviewReq := range reviewsReq {
		validatorResult, ok := checkFormat(reviewReq)
		if !ok {
			return rest.Response{Status: rest.StatusFail, Data: validatorResult}, http.StatusBadRequest
		}
		newReview := review.Review{Id: uuid.NewString(), RawReview: reviewReq}
		resultReviews = append(resultReviews, newReview)
	}
	service.database.AddAllReviews(resultReviews)
	return rest.SuccessReviewResponse(resultReviews), http.StatusCreated
}

func (service *ReviewService) GetReviewsByUser(userId string) []review.Review {
	return service.database.GetReviewsByUser(userId)
}

func (service *ReviewService) GetReviewsByFilm(contentId string) []review.Review {
	return service.database.GetReviewsByFilm(contentId)
}

func (service *ReviewService) DeleteReview(id string) (rest.Response, int) {
	_, err := uuid.Parse(id)
	if err != nil {
		return rest.InvalidUUIDResponse("contentId"), http.StatusBadRequest
	}
	_, ok := service.database.GetReview(id)
	if !ok {
		return rest.DeleteReviewNotFoundResponse(), http.StatusNotFound
	}
	service.database.DeleteReview(id)
	return rest.Response{Status: rest.StatusSuccess, Data: nil}, http.StatusOK
}

func checkFormat(reviewReq review.RawReview) (map[string]string, bool) {
	errors := make(map[string]string)

	_, err := uuid.Parse(reviewReq.ContentId)
	if err != nil {
		errors["contentId"] = rest.InvalidIDMessage
	}

	_, err = uuid.Parse(reviewReq.UserId)
	if err != nil {
		errors["userId"] = rest.InvalidIDMessage
	}

	if reviewReq.Released != "" {
		_, err = time.Parse(time.DateOnly, reviewReq.Released)
		if err != nil {
			errors["released"] = rest.InvalidDateFormatMessage
		}
	}

	if len(errors) > 0 {
		return errors, false
	}

	return nil, true
}
