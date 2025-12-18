package model

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a user's review for a movie.
type Review struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	MovieID   uuid.UUID `json:"movieId"`
	Rating    int       `json:"rating"` // 1-5
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
