package model

import "github.com/google/uuid"

// Recommendation represents a movie recommendation for a user.
type Recommendation struct {
	MovieID uuid.UUID `json:"movieId"`
	Score   float64   `json:"score"`
}
