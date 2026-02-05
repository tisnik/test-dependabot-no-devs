package request

import "github.com/google/uuid"

type RawReview struct {
	ContentID   uuid.UUID `json:"contentId"             validate:"required"`
	UserID      uuid.UUID `json:"userId"                validate:"required"`
	Title       string    `json:"title,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	Director    string    `json:"director,omitempty"`
	Actors      []string  `json:"actors,omitempty"`
	Origins     []string  `json:"origins,omitempty"`
	Duration    int       `json:"duration,omitempty"`
	Released    string    `json:"released,omitempty"`
	Review      string    `json:"review"                validate:"required"`
	Score       int       `json:"score"                 validate:"required"`
}

type ReviewData struct {
	Reviews []RawReview `json:"reviews"`
}

type RawReviewsRequest struct {
	Data ReviewData `json:"data"`
}
