package entity

import "github.com/google/uuid"

type Content struct {
	ContentID   uuid.UUID `json:"contentId"`
	Title       string    `json:"title,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	Director    string    `json:"director,omitempty"`
	Actors      []string  `json:"actors,omitempty"`
	Origins     []string  `json:"origins,omitempty"`
	Duration    int       `json:"duration,omitempty"`
	Released    string    `json:"released,omitempty"`
}
