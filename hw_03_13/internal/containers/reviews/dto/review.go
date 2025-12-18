package dto

import (
	"github.com/course-go/reelgoofy/internal/containers/reviews/enums"
)

const (
	MaxScore = 100
	MinScore = 1
)

// Review represents a data with UUID.
type Review struct {
	RawReview
	SimilarityVector `json:"-"`

	ID string `json:"id"`
}

// SimilarityVector is used for comparison between different Reviews.
type SimilarityVector []float64

// RawReview represents a DTO for new Review creation.
type RawReview struct {
	ContentId   string        `json:"contentId"             validate:"required,uuid4"`
	UserId      string        `json:"userId"                validate:"required,uuid4"`
	Title       string        `json:"title,omitempty"       validate:"omitempty,max=200"`
	Genres      []enums.Genre `json:"genres,omitempty"      validate:"omitempty"`
	Tags        []string      `json:"tags,omitempty"        validate:"omitempty,max=50"`
	Description string        `json:"description,omitempty" validate:"omitempty,max=1000"`
	Director    string        `json:"director,omitempty"    validate:"omitempty,max=100"`
	Actors      []string      `json:"actors,omitempty"      validate:"omitempty"`
	Origins     []string      `json:"origins,omitempty"     validate:"omitempty,dive,max=100"`
	Duration    int           `json:"duration,omitempty"    validate:"omitempty,min=1"`
	Released    string        `json:"released"              validate:"omitempty,datetime=2006-01-02"`
	Review      string        `json:"review"                validate:"required,max=5000"`
	Score       int           `json:"score"                 validate:"required,min=0,max=100"`
}

// CanBeValidated converts struct to a valid interface and so it can be validated.
func (r RawReview) CanBeValidated() {}
