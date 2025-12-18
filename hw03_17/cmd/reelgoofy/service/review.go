package service

import (
	"time"

	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/repository"
	"github.com/MiriamVenglikova/assignment-3-reelgoofy/cmd/reelgoofy/structures"
	"github.com/google/uuid"
)

const validID = "ID is not a valid UUID"

type ReviewService struct {
	reviewTable *repository.ReviewTable
}

func NewReviewService(r *repository.ReviewTable) *ReviewService {
	return &ReviewService{reviewTable: r}
}

func (s *ReviewService) UploadReview(raw structures.RawReview) (review structures.Review, errs map[string]any) {
	errs = s.validate(raw)
	if len(errs) > 0 {
		return structures.Review{}, errs
	}

	review = structures.Review{
		ID:        uuid.New().String(),
		RawReview: raw,
	}

	s.reviewTable.Add(review)
	return review, nil
}

func (s *ReviewService) DeleteReview(id string) (validate bool, errs map[string]any) {
	errs = make(map[string]any)
	_, err := uuid.Parse(id)
	if err != nil {
		errs["id"] = validID
		return false, errs
	}

	validate = s.reviewTable.Delete(id)
	if !validate {
		return false, nil
	}

	return true, nil
}

func (s *ReviewService) GetAllReviews() []structures.Review {
	return s.reviewTable.GetAll()
}

func (s *ReviewService) validate(raw structures.RawReview) map[string]any {
	errors := make(map[string]any)

	if raw.ContentID == "" {
		errors["contentId"] = "Content ID is required"
	}
	if raw.UserID == "" {
		errors["userId"] = "User ID is required"
	}

	_, err := uuid.Parse(raw.ContentID)
	if err != nil {
		errors["contentId"] = validID
	}
	_, err = uuid.Parse(raw.UserID)
	if err != nil {
		errors["userId"] = validID
	}
	if raw.Review == "" {
		errors["review"] = "Review text is required"
	}
	if raw.Score == 0 {
		errors["score"] = "Score is required"
	}
	if raw.Score != 0 && (raw.Score < 0 || raw.Score > 100) {
		errors["score"] = "Score must be between 1 and 100"
	}

	_, err = time.Parse("2006-01-02", raw.Released)
	if raw.Released != "" && err != nil {
		errors["released"] = "Invalid date format, correct: YYYY-MM-DD"
	}

	return errors
}
