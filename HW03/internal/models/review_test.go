package models_test

import (
	"testing"

	"github.com/course-go/reelgoofy/internal/models"
	"github.com/google/uuid"
)

func TestValidateRawReview_Valid(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Review:    "Great movie!",
		Score:     85,
	}

	errors := review.Validate()
	if len(errors) != 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
}

func TestValidateRawReview_MissingContentID(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		UserID: "550e8400-e29b-41d4-a716-446655440001",
		Review: "Great movie!",
		Score:  85,
	}

	errors := review.Validate()
	if _, exists := errors["contentId"]; !exists {
		t.Error("Should have error - missing contentId")
	}
}

func TestValidateRawReview_InvalidContentIDUUID(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "not-a-uuid",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Review:    "Great movie!",
		Score:     85,
	}

	errors := review.Validate()
	if _, exists := errors["contentId"]; !exists {
		t.Error("Should have error - invalid contentId UUID")
	}
}

func TestValidateRawReview_MissingUserID(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		Review:    "Great movie!",
		Score:     85,
	}

	errors := review.Validate()
	if _, exists := errors["userId"]; !exists {
		t.Error("Should have error - missing userId")
	}
}

func TestValidateRawReview_InvalidUserIDUUID(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "not-a-uuid",
		Review:    "Great movie!",
		Score:     85,
	}

	errors := review.Validate()
	if _, exists := errors["userId"]; !exists {
		t.Error("Should have error - invalid userId (UUID)")
	}
}

func TestValidateRawReview_MissingReview(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Score:     85,
	}

	errors := review.Validate()
	if _, exists := errors["review"]; !exists {
		t.Error("Should have error - missing review text")
	}
}

func TestValidateRawReview_ScoreTooLow(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Review:    "Great movie!",
		Score:     -1,
	}

	errors := review.Validate()
	if _, exists := errors["score"]; !exists {
		t.Error("Should have error for score < 0")
	}
}

func TestValidateRawReview_ScoreTooHigh(t *testing.T) {
	t.Parallel()
	review := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Review:    "Great movie!",
		Score:     101,
	}

	errors := review.Validate()
	if _, exists := errors["score"]; !exists {
		t.Error("Should have error for score > 100")
	}
}

func TestToReview(t *testing.T) {
	t.Parallel()
	rawReview := models.RawReview{
		ContentID: "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "550e8400-e29b-41d4-a716-446655440001",
		Title:     "The Matrix",
		Duration:  136,
		Review:    "Mind-blowing movie!",
		Score:     98,
	}

	review := rawReview.ToReview()

	if review.ID == uuid.Nil {
		t.Error("Review ID should be generated")
	}
	if review.ContentID.String() != rawReview.ContentID {
		t.Error("ContentID should match")
	}
	if review.UserID.String() != rawReview.UserID {
		t.Error("UserID should match")
	}
	if review.Title != rawReview.Title {
		t.Error("Title should match")
	}
	if review.Duration != rawReview.Duration {
		t.Error("Duration should match")
	}
	if review.Review != rawReview.Review {
		t.Error("Review text should match")
	}
	if review.Score != rawReview.Score {
		t.Error("Score should match")
	}
}
