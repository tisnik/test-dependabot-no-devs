package models

import (
	"github.com/google/uuid"
)

// RawReview represents incoming review without an ID.
// Uses strings for validation at API boundary.
type RawReview struct {
	ContentID string `json:"contentId"`
	UserID    string `json:"userId"`
	Title     string `json:"title,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	Review    string `json:"review"`
	Score     int    `json:"score"`
}

// Review represents a validated review with generated UUID identifiers.
// All IDs are guaranteed to be valid UUIDs after creation.
type Review struct {
	ID        uuid.UUID `json:"id"`
	ContentID uuid.UUID `json:"contentId"`
	UserID    uuid.UUID `json:"userId"`
	Title     string    `json:"title,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	Review    string    `json:"review"`
	Score     int       `json:"score"`
}

// RawReviewsRequest represents the request body for bulk review ingestion.
// Wraps the reviews array in a JSend-compliant data structure.
type RawReviewsRequest struct {
	Data struct {
		Reviews []RawReview `json:"reviews"`
	} `json:"data"`
}

// ReviewsData wraps the reviews array for response.
type ReviewsData struct {
	Reviews []Review `json:"reviews"`
}

func (r *RawReview) Validate() map[string]string {
	errors := make(map[string]string)

	if r.ContentID == "" {
		errors["contentId"] = "Content ID is required."
	} else {
		_, err := uuid.Parse(r.ContentID)
		if err != nil {
			errors["contentId"] = "ID is not a valid UUID."
		}
	}

	if r.UserID == "" {
		errors["userId"] = "User ID is required."
	} else {
		_, err := uuid.Parse(r.UserID)
		if err != nil {
			errors["userId"] = "ID is not a valid UUID."
		}
	}

	if r.Review == "" {
		errors["review"] = "Review text is required."
	}

	if r.Score < 0 || r.Score > 100 {
		errors["score"] = "Score must be between 0 and 100."
	}

	return errors
}

// ToReview converts a RawReview to a Review with a generated ID.
// assumes validation has already been performed.
func (r *RawReview) ToReview() Review {
	contentID, _ := uuid.Parse(r.ContentID)
	userID, _ := uuid.Parse(r.UserID)

	return Review{
		ID:        uuid.New(),
		ContentID: contentID,
		UserID:    userID,
		Title:     r.Title,
		Duration:  r.Duration,
		Review:    r.Review,
		Score:     r.Score,
	}
}
