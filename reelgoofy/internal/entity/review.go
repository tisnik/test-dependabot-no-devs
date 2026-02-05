package entity

import "github.com/google/uuid"

type Review struct {
	ID          uuid.UUID `json:"id,omitempty"`
	ContentID   uuid.UUID `json:"contentId"`
	UserID      uuid.UUID `json:"userId"`
	Title       string    `json:"title,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	Director    string    `json:"director,omitempty"`
	Actors      []string  `json:"actors,omitempty"`
	Origins     []string  `json:"origins,omitempty"`
	Duration    int       `json:"duration,omitempty"`
	Released    string    `json:"released,omitempty"`
	Review      string    `json:"review"`
	Score       int       `json:"score"`
}
